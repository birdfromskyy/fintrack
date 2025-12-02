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

type TransactionService struct {
	db         *sql.DB
	logService *LogService
}

func NewTransactionService(db *sql.DB, logService *LogService) *TransactionService {
	return &TransactionService{
		db:         db,
		logService: logService,
	}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, userID string, req *models.CreateTransactionRequest) (*models.Transaction, error) {
	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get category type to determine transaction type
	var categoryType string
	err = tx.QueryRowContext(ctx,
		`SELECT type FROM categories WHERE id = $1 AND (user_id = $2 OR is_system = true)`,
		req.CategoryID, userID).Scan(&categoryType)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category type: %w", err)
	}

	// Verify account belongs to user
	var accountExists bool
	err = tx.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1 AND user_id = $2)`,
		req.AccountID, userID).Scan(&accountExists)

	if err != nil {
		return nil, fmt.Errorf("failed to verify account: %w", err)
	}

	if !accountExists {
		return nil, fmt.Errorf("account not found")
	}

	// Parse date properly - expecting "YYYY-MM-DD" format
	var transactionDate time.Time
	if req.Date != "" {
		transactionDate, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
		}
	} else {
		transactionDate = time.Now()
	}

	// Create transaction
	transaction := &models.Transaction{
		ID:          uuid.New().String(),
		UserID:      userID,
		AccountID:   req.AccountID,
		CategoryID:  req.CategoryID,
		Type:        categoryType,
		Amount:      req.Amount,
		Description: req.Description,
		Date:        transactionDate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO transactions (id, user_id, account_id, category_id, type, amount, description, date, created_at, updated_at) 
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		transaction.ID, transaction.UserID, transaction.AccountID, transaction.CategoryID,
		transaction.Type, transaction.Amount, transaction.Description, transaction.Date,
		transaction.CreatedAt, transaction.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Update account balance
	if categoryType == "income" {
		_, err = tx.ExecContext(ctx,
			`UPDATE accounts SET balance = balance + $1 WHERE id = $2`,
			req.Amount, req.AccountID)
	} else {
		_, err = tx.ExecContext(ctx,
			`UPDATE accounts SET balance = balance - $1 WHERE id = $2`,
			req.Amount, req.AccountID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to update account balance: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// ✅ Детальное логирование создания
	logDetails := map[string]interface{}{
		"action": "created",
		"data": map[string]interface{}{
			"id":          transaction.ID,
			"type":        transaction.Type,
			"amount":      transaction.Amount,
			"description": transaction.Description,
			"date":        transaction.Date.Format("2006-01-02"),
			"account_id":  transaction.AccountID,
			"category_id": transaction.CategoryID,
		},
	}
	detailsJSON, _ := json.Marshal(logDetails)

	go s.logService.Log(context.Background(), &UserAction{
		UserID:   userID,
		Action:   "create",
		Entity:   "transaction",
		EntityID: transaction.ID,
		Details:  string(detailsJSON),
	})

	return transaction, nil

}

func (s *TransactionService) GetTransactions(ctx context.Context, filter *models.TransactionFilter) ([]*models.Transaction, error) {
	query := `
        SELECT 
            t.id, t.user_id, t.account_id, t.category_id, t.type, 
            t.amount, t.description, t.date, t.created_at, t.updated_at,
            a.name as account_name,
            c.name as category_name, c.icon as category_icon, c.color as category_color
        FROM transactions t
        JOIN accounts a ON t.account_id = a.id
        JOIN categories c ON t.category_id = c.id
        WHERE t.user_id = $1`

	args := []interface{}{filter.UserID}
	argCount := 1

	if filter.AccountID != "" {
		argCount++
		query += fmt.Sprintf(" AND t.account_id = $%d", argCount)
		args = append(args, filter.AccountID)
	}

	if filter.CategoryID != "" {
		argCount++
		query += fmt.Sprintf(" AND t.category_id = $%d", argCount)
		args = append(args, filter.CategoryID)
	}

	if filter.Type != "" {
		argCount++
		query += fmt.Sprintf(" AND t.type = $%d", argCount)
		args = append(args, filter.Type)
	}

	if !filter.DateFrom.IsZero() {
		argCount++
		query += fmt.Sprintf(" AND t.date >= $%d", argCount)
		args = append(args, filter.DateFrom)
	}

	if !filter.DateTo.IsZero() {
		argCount++
		query += fmt.Sprintf(" AND t.date <= $%d", argCount)
		args = append(args, filter.DateTo)
	}

	query += " ORDER BY t.date DESC, t.created_at DESC"

	if filter.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		var t models.Transaction
		err := rows.Scan(
			&t.ID, &t.UserID, &t.AccountID, &t.CategoryID, &t.Type,
			&t.Amount, &t.Description, &t.Date, &t.CreatedAt, &t.UpdatedAt,
			&t.AccountName, &t.CategoryName, &t.CategoryIcon, &t.CategoryColor,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, &t)
	}

	return transactions, nil
}

func (s *TransactionService) GetTransaction(ctx context.Context, userID, transactionID string) (*models.Transaction, error) {
	var t models.Transaction
	err := s.db.QueryRowContext(ctx,
		`SELECT 
            t.id, t.user_id, t.account_id, t.category_id, t.type, 
            t.amount, t.description, t.date, t.created_at, t.updated_at,
            a.name as account_name,
            c.name as category_name, c.icon as category_icon, c.color as category_color
        FROM transactions t
        JOIN accounts a ON t.account_id = a.id
        JOIN categories c ON t.category_id = c.id
        WHERE t.id = $1 AND t.user_id = $2`,
		transactionID, userID).Scan(
		&t.ID, &t.UserID, &t.AccountID, &t.CategoryID, &t.Type,
		&t.Amount, &t.Description, &t.Date, &t.CreatedAt, &t.UpdatedAt,
		&t.AccountName, &t.CategoryName, &t.CategoryIcon, &t.CategoryColor,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &t, nil
}

func (s *TransactionService) UpdateTransaction(ctx context.Context, userID, transactionID string, req *models.UpdateTransactionRequest) (*models.Transaction, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current transaction
	var oldTransaction models.Transaction
	err = tx.QueryRowContext(ctx,
		`SELECT id, user_id, account_id, category_id, type, amount, description, date 
         FROM transactions WHERE id = $1 AND user_id = $2`,
		transactionID, userID).Scan(
		&oldTransaction.ID, &oldTransaction.UserID, &oldTransaction.AccountID,
		&oldTransaction.CategoryID, &oldTransaction.Type, &oldTransaction.Amount,
		&oldTransaction.Description, &oldTransaction.Date,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	changes := make(map[string]map[string]interface{})

	// Revert old transaction from account balance
	if oldTransaction.Type == "income" {
		_, err = tx.ExecContext(ctx,
			`UPDATE accounts SET balance = balance - $1 WHERE id = $2`,
			oldTransaction.Amount, oldTransaction.AccountID)
	} else {
		_, err = tx.ExecContext(ctx,
			`UPDATE accounts SET balance = balance + $1 WHERE id = $2`,
			oldTransaction.Amount, oldTransaction.AccountID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to revert old balance: %w", err)
	}

	// Update transaction fields
	if req.AccountID != "" && req.AccountID != oldTransaction.AccountID {
		changes["account_id"] = map[string]interface{}{
			"old": oldTransaction.AccountID,
			"new": req.AccountID,
		}
		oldTransaction.AccountID = req.AccountID
	}

	if req.CategoryID != "" && req.CategoryID != oldTransaction.CategoryID {
		// Get new category type
		var categoryType string
		err = tx.QueryRowContext(ctx,
			`SELECT type FROM categories WHERE id = $1 AND (user_id = $2 OR is_system = true)`,
			req.CategoryID, userID).Scan(&categoryType)
		if err != nil {
			return nil, fmt.Errorf("failed to get category type: %w", err)
		}

		changes["category_id"] = map[string]interface{}{
			"old": oldTransaction.CategoryID,
			"new": req.CategoryID,
		}
		changes["type"] = map[string]interface{}{
			"old": oldTransaction.Type,
			"new": categoryType,
		}

		oldTransaction.CategoryID = req.CategoryID
		oldTransaction.Type = categoryType
	}

	if req.Amount > 0 && req.Amount != oldTransaction.Amount {
		changes["amount"] = map[string]interface{}{
			"old": oldTransaction.Amount,
			"new": req.Amount,
		}
		oldTransaction.Amount = req.Amount
	}

	if req.Description != "" && req.Description != oldTransaction.Description {
		changes["description"] = map[string]interface{}{
			"old": oldTransaction.Description,
			"new": req.Description,
		}
		oldTransaction.Description = req.Description
	}

	if req.Date != "" {
		transactionDate, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
		}
		if !transactionDate.Equal(oldTransaction.Date) {
			changes["date"] = map[string]interface{}{
				"old": oldTransaction.Date.Format("2006-01-02"),
				"new": transactionDate.Format("2006-01-02"),
			}
			oldTransaction.Date = transactionDate
		}
	}

	oldTransaction.UpdatedAt = time.Now()

	// Update transaction in database
	_, err = tx.ExecContext(ctx,
		`UPDATE transactions SET account_id = $1, category_id = $2, type = $3, 
         amount = $4, description = $5, date = $6, updated_at = $7 
         WHERE id = $8`,
		oldTransaction.AccountID, oldTransaction.CategoryID, oldTransaction.Type,
		oldTransaction.Amount, oldTransaction.Description, oldTransaction.Date,
		oldTransaction.UpdatedAt, transactionID)

	if err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	// Apply new transaction to account balance
	if oldTransaction.Type == "income" {
		_, err = tx.ExecContext(ctx,
			`UPDATE accounts SET balance = balance + $1 WHERE id = $2`,
			oldTransaction.Amount, oldTransaction.AccountID)
	} else {
		_, err = tx.ExecContext(ctx,
			`UPDATE accounts SET balance = balance - $1 WHERE id = $2`,
			oldTransaction.Amount, oldTransaction.AccountID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to update new balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// ✅ Логирование только если были изменения
	if len(changes) > 0 {
		logDetails := map[string]interface{}{
			"action":  "updated",
			"changes": changes,
		}
		detailsJSON, _ := json.Marshal(logDetails)

		go s.logService.Log(context.Background(), &UserAction{
			UserID:   userID,
			Action:   "update",
			Entity:   "transaction",
			EntityID: transactionID,
			Details:  string(detailsJSON),
		})
	}

	return &oldTransaction, nil

}

func (s *TransactionService) DeleteTransaction(ctx context.Context, userID, transactionID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get transaction details
	var accountID string
	var transactionType string
	var amount float64

	err = tx.QueryRowContext(ctx,
		`SELECT account_id, type, amount FROM transactions WHERE id = $1 AND user_id = $2`,
		transactionID, userID).Scan(&accountID, &transactionType, &amount)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("transaction not found")
		}
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	// Delete transaction
	result, err := tx.ExecContext(ctx,
		`DELETE FROM transactions WHERE id = $1 AND user_id = $2`,
		transactionID, userID)

	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction not found")
	}

	// Update account balance
	if transactionType == "income" {
		_, err = tx.ExecContext(ctx,
			`UPDATE accounts SET balance = balance - $1 WHERE id = $2`,
			amount, accountID)
	} else {
		_, err = tx.ExecContext(ctx,
			`UPDATE accounts SET balance = balance + $1 WHERE id = $2`,
			amount, accountID)
	}

	if err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// ✅ Детальное логирование удаления
	logDetails := map[string]interface{}{
		"action": "deleted",
		"data": map[string]interface{}{
			"account_id": accountID,
			"type":       transactionType,
			"amount":     amount,
		},
	}
	detailsJSON, _ := json.Marshal(logDetails)

	go s.logService.Log(context.Background(), &UserAction{
		UserID:   userID,
		Action:   "delete",
		Entity:   "transaction",
		EntityID: transactionID,
		Details:  string(detailsJSON),
	})

	return nil
}
