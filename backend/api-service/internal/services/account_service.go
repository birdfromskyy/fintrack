package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"api-service/internal/models"

	"github.com/google/uuid"
)

type AccountService struct {
	db         *sql.DB
	logService *LogService
}

func NewAccountService(db *sql.DB, logService *LogService) *AccountService {
	return &AccountService{
		db:         db,
		logService: logService,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, userID string, req *models.CreateAccountRequest) (*models.Account, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// If setting as default, unset other defaults
	if req.IsDefault {
		_, err = tx.ExecContext(ctx,
			`UPDATE accounts SET is_default = false WHERE user_id = $1`,
			userID)
		if err != nil {
			return nil, fmt.Errorf("failed to unset default accounts: %w", err)
		}
	}

	account := &models.Account{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      req.Name,
		Balance:   req.Balance,
		IsDefault: req.IsDefault,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO accounts (id, user_id, name, balance, is_default, created_at, updated_at) 
         VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		account.ID, account.UserID, account.Name, account.Balance,
		account.IsDefault, account.CreatedAt, account.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// ✅ Логирование
	logDetails := map[string]interface{}{
		"action": "created",
		"data": map[string]interface{}{
			"id":         account.ID,
			"name":       account.Name,
			"balance":    account.Balance,
			"is_default": account.IsDefault,
		},
	}
	detailsJSON, _ := json.Marshal(logDetails)

	go s.logService.Log(context.Background(), &UserAction{
		UserID:   userID,
		Action:   "create",
		Entity:   "account",
		EntityID: account.ID,
		Details:  string(detailsJSON),
	})

	return account, nil
}

func (s *AccountService) GetAccounts(ctx context.Context, userID string) ([]*models.Account, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, name, balance, is_default, created_at, updated_at 
         FROM accounts WHERE user_id = $1 ORDER BY is_default DESC, created_at ASC`,
		userID)

	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*models.Account
	for rows.Next() {
		var a models.Account
		err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Balance,
			&a.IsDefault, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, &a)
	}

	return accounts, nil
}

func (s *AccountService) GetAccount(ctx context.Context, userID, accountID string) (*models.Account, error) {
	var a models.Account
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, balance, is_default, created_at, updated_at 
         FROM accounts WHERE id = $1 AND user_id = $2`,
		accountID, userID).Scan(&a.ID, &a.UserID, &a.Name, &a.Balance,
		&a.IsDefault, &a.CreatedAt, &a.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &a, nil
}

func (s *AccountService) UpdateAccount(ctx context.Context, userID, accountID string, req *models.UpdateAccountRequest) (*models.Account, error) {
	// ✅ ШАГ 1: Получаем СТАРЫЙ аккаунт (до изменений)
	oldAccount, err := s.GetAccount(ctx, userID, accountID)
	if err != nil {
		return nil, err // Уже содержит "account not found" если не найден
	}

	// ✅ ШАГ 2: Build update query + отслеживаем изменения
	updateFields := make(map[string]interface{})
	changes := make(map[string]map[string]interface{}) // ← Для логов

	// Сравниваем Name
	if req.Name != "" && req.Name != oldAccount.Name {
		updateFields["name"] = req.Name
		changes["name"] = map[string]interface{}{
			"old": oldAccount.Name,
			"new": req.Name,
		}
	}

	// Сравниваем Balance
	if req.Balance != 0 && req.Balance != oldAccount.Balance {
		updateFields["balance"] = req.Balance
		changes["balance"] = map[string]interface{}{
			"old": oldAccount.Balance,
			"new": req.Balance,
		}
	}

	// Если изменений нет — ничего не делаем
	if len(updateFields) == 0 {
		return oldAccount, nil
	}

	updateFields["updated_at"] = time.Now()

	// ✅ ШАГ 3: Execute update
	query := `UPDATE accounts SET `
	args := []interface{}{}
	i := 1

	for field, value := range updateFields {
		if i > 1 {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", field, i)
		args = append(args, value)
		i++
	}

	query += fmt.Sprintf(" WHERE id = $%d AND user_id = $%d", i, i+1)
	args = append(args, accountID, userID)

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	// ✅ ШАГ 4: Логирование с деталями "было → стало"
	if len(changes) > 0 {
		logDetails := map[string]interface{}{
			"action":  "updated",
			"changes": changes,
		}
		detailsJSON, _ := json.Marshal(logDetails)

		go s.logService.Log(context.Background(), &UserAction{
			UserID:   userID,
			Action:   "update",
			Entity:   "account",
			EntityID: accountID,
			Details:  string(detailsJSON),
		})
	}

	return s.GetAccount(ctx, userID, accountID)
}

func (s *AccountService) DeleteAccount(ctx context.Context, userID, accountID string) error {
	// Check if it's the only account
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM accounts WHERE user_id = $1`,
		userID).Scan(&count)

	if err != nil {
		return fmt.Errorf("failed to count accounts: %w", err)
	}

	if count <= 1 {
		return fmt.Errorf("cannot delete the only account")
	}

	// Check if account has transactions
	var transactionCount int
	err = s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM transactions WHERE account_id = $1`,
		accountID).Scan(&transactionCount)

	if err != nil {
		return fmt.Errorf("failed to check transactions: %w", err)
	}

	if transactionCount > 0 {
		return fmt.Errorf("cannot delete account with existing transactions")
	}

	// ✅ ШАГ 1: Сохраняем данные аккаунта ДО удаления (для логов)
	var accountName string
	var balance float64
	var isDefault bool
	err = s.db.QueryRowContext(ctx,
		`SELECT name, balance, is_default FROM accounts WHERE id = $1 AND user_id = $2`,
		accountID, userID).Scan(&accountName, &balance, &isDefault)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("account not found")
		}
		return fmt.Errorf("failed to get account info: %w", err)
	}

	// ✅ ШАГ 2: Начинаем транзакцию
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete account
	result, err := tx.ExecContext(ctx,
		`DELETE FROM accounts WHERE id = $1 AND user_id = $2`,
		accountID, userID)

	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account not found")
	}

	// If it was default, set another as default
	if isDefault {
		_, err = tx.ExecContext(ctx,
			`UPDATE accounts SET is_default = true 
			WHERE user_id = $1 
			ORDER BY created_at ASC 
			LIMIT 1`,
			userID)

		if err != nil {
			return fmt.Errorf("failed to set new default account: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// ✅ ШАГ 3: Логирование с деталями удалённого аккаунта
	logDetails := map[string]interface{}{
		"action": "deleted",
		"data": map[string]interface{}{
			"name":       accountName,
			"balance":    balance,
			"is_default": isDefault,
		},
	}
	detailsJSON, _ := json.Marshal(logDetails)

	go s.logService.Log(context.Background(), &UserAction{
		UserID:   userID,
		Action:   "delete",
		Entity:   "account",
		EntityID: accountID,
		Details:  string(detailsJSON),
	})

	return nil
}

func (s *AccountService) SetDefaultAccount(ctx context.Context, userID, accountID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if account exists and belongs to user
	var exists bool
	err = tx.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1 AND user_id = $2)`,
		accountID, userID).Scan(&exists)

	if err != nil {
		return fmt.Errorf("failed to check account: %w", err)
	}

	if !exists {
		return fmt.Errorf("account not found")
	}

	// Unset all defaults
	_, err = tx.ExecContext(ctx,
		`UPDATE accounts SET is_default = false WHERE user_id = $1`,
		userID)

	if err != nil {
		return fmt.Errorf("failed to unset default accounts: %w", err)
	}

	// Set new default
	_, err = tx.ExecContext(ctx,
		`UPDATE accounts SET is_default = true WHERE id = $1 AND user_id = $2`,
		accountID, userID)

	if err != nil {
		return fmt.Errorf("failed to set default account: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *AccountService) GetAccountStats(ctx context.Context, userID, accountID string) (*models.AccountStats, error) {
	var stats models.AccountStats

	// Get current balance
	err := s.db.QueryRowContext(ctx,
		`SELECT balance FROM accounts WHERE id = $1 AND user_id = $2`,
		accountID, userID).Scan(&stats.CurrentBalance)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	// Get total income
	err = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM transactions 
         WHERE account_id = $1 AND user_id = $2 AND type = 'income'`,
		accountID, userID).Scan(&stats.TotalIncome)

	if err != nil {
		return nil, fmt.Errorf("failed to get total income: %w", err)
	}

	// Get total expense
	err = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM transactions 
         WHERE account_id = $1 AND user_id = $2 AND type = 'expense'`,
		accountID, userID).Scan(&stats.TotalExpense)

	if err != nil {
		return nil, fmt.Errorf("failed to get total expense: %w", err)
	}

	return &stats, nil
}
