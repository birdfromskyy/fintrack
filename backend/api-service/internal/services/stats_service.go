package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type StatsService struct {
	db *sql.DB
}

func NewStatsService(db *sql.DB) *StatsService {
	return &StatsService{db: db}
}

type Summary struct {
	TotalIncome       float64 `json:"total_income"`
	TotalExpense      float64 `json:"total_expense"`
	Balance           float64 `json:"balance"`
	AccountsCount     int     `json:"accounts_count"`
	TransactionsCount int     `json:"transactions_count"`
}

type MonthlyStats struct {
	Month        string  `json:"month"`
	Year         int     `json:"year"`
	Income       float64 `json:"income"`
	Expense      float64 `json:"expense"`
	Balance      float64 `json:"balance"`
	Transactions int     `json:"transactions"`
}

type DailyBalance struct {
	Date    time.Time `json:"date"`
	Balance float64   `json:"balance"`
	Income  float64   `json:"income"`
	Expense float64   `json:"expense"`
}

func (s *StatsService) GetSummary(ctx context.Context, userID string) (*Summary, error) {
	summary := &Summary{}

	// Get total balance from all accounts
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(balance), 0), COUNT(*) FROM accounts WHERE user_id = $1`,
		userID).Scan(&summary.Balance, &summary.AccountsCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts summary: %w", err)
	}

	// Get total income
	err = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM transactions 
         WHERE user_id = $1 AND type = 'income'`,
		userID).Scan(&summary.TotalIncome)
	if err != nil {
		return nil, fmt.Errorf("failed to get total income: %w", err)
	}

	// Get total expense
	err = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM transactions 
         WHERE user_id = $1 AND type = 'expense'`,
		userID).Scan(&summary.TotalExpense)
	if err != nil {
		return nil, fmt.Errorf("failed to get total expense: %w", err)
	}

	// Get transactions count
	err = s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM transactions WHERE user_id = $1`,
		userID).Scan(&summary.TransactionsCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions count: %w", err)
	}

	return summary, nil
}

func (s *StatsService) GetMonthlyStats(ctx context.Context, userID string, months int) ([]*MonthlyStats, error) {
	if months <= 0 {
		months = 12
	}

	startDate := time.Now().AddDate(0, -months+1, 0).Format("2006-01-01")

	query := `
        WITH months AS (
            SELECT 
                DATE_TRUNC('month', t.date) as month,
                t.type,
                SUM(t.amount) as total,
                COUNT(t.id) as count
            FROM transactions t
            WHERE t.user_id = $1 AND t.date >= $2::date
            GROUP BY DATE_TRUNC('month', t.date), t.type
        )
        SELECT 
            TO_CHAR(month, 'Month') as month_name,
            EXTRACT(YEAR FROM month) as year,
            COALESCE(SUM(CASE WHEN type = 'income' THEN total END), 0) as income,
            COALESCE(SUM(CASE WHEN type = 'expense' THEN total END), 0) as expense,
            SUM(count) as transactions
        FROM months
        GROUP BY month, month_name, year
        ORDER BY month DESC`

	rows, err := s.db.QueryContext(ctx, query, userID, startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly stats: %w", err)
	}
	defer rows.Close()

	var stats []*MonthlyStats
	for rows.Next() {
		var stat MonthlyStats
		err := rows.Scan(&stat.Month, &stat.Year, &stat.Income, &stat.Expense, &stat.Transactions)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monthly stat: %w", err)
		}
		stat.Balance = stat.Income - stat.Expense
		stats = append(stats, &stat)
	}

	return stats, nil
}

func (s *StatsService) GetBalanceHistory(ctx context.Context, userID string, days int) ([]*DailyBalance, error) {
	if days <= 0 {
		days = 30
	}

	startDate := time.Now().AddDate(0, 0, -days+1)

	query := `
        WITH daily_transactions AS (
            SELECT 
                DATE(date) as day,
                SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) as income,
                SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END) as expense
            FROM transactions
            WHERE user_id = $1 AND date >= $2
            GROUP BY DATE(date)
        ),
        date_series AS (
            SELECT generate_series($2::date, CURRENT_DATE, '1 day'::interval)::date AS day
        )
        SELECT 
            ds.day,
            COALESCE(dt.income, 0) as income,
            COALESCE(dt.expense, 0) as expense
        FROM date_series ds
        LEFT JOIN daily_transactions dt ON ds.day = dt.day
        ORDER BY ds.day`

	rows, err := s.db.QueryContext(ctx, query, userID, startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance history: %w", err)
	}
	defer rows.Close()

	// Get initial balance
	var initialBalance float64
	err = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE user_id = $1`,
		userID).Scan(&initialBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial balance: %w", err)
	}

	// Calculate balance before the period
	var priorIncome, priorExpense float64
	err = s.db.QueryRowContext(ctx,
		`SELECT 
            COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)
         FROM transactions 
         WHERE user_id = $1 AND date < $2`,
		userID, startDate).Scan(&priorIncome, &priorExpense)
	if err != nil {
		return nil, fmt.Errorf("failed to get prior transactions: %w", err)
	}

	runningBalance := initialBalance - (priorIncome - priorExpense)

	var history []*DailyBalance
	for rows.Next() {
		var daily DailyBalance
		err := rows.Scan(&daily.Date, &daily.Income, &daily.Expense)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily balance: %w", err)
		}

		runningBalance += daily.Income - daily.Expense
		daily.Balance = runningBalance

		history = append(history, &daily)
	}

	return history, nil
}

func (s *StatsService) GetCategoryBreakdown(ctx context.Context, userID string, transactionType string, period string) (map[string]interface{}, error) {
	var startDate time.Time

	switch period {
	case "week":
		startDate = time.Now().AddDate(0, 0, -7)
	case "month":
		startDate = time.Now().AddDate(0, -1, 0)
	case "year":
		startDate = time.Now().AddDate(-1, 0, 0)
	default:
		startDate = time.Now().AddDate(0, -1, 0) // default to month
	}

	query := `
        SELECT 
            c.name,
            c.color,
            c.icon,
            COALESCE(SUM(t.amount), 0) as total,
            COUNT(t.id) as count
        FROM categories c
        LEFT JOIN transactions t ON c.id = t.category_id 
            AND t.user_id = $1 
            AND t.type = $2 
            AND t.date >= $3
        WHERE (c.user_id = $1 OR c.is_system = true) AND c.type = $2
        GROUP BY c.id, c.name, c.color, c.icon
        HAVING COUNT(t.id) > 0
        ORDER BY total DESC`

	rows, err := s.db.QueryContext(ctx, query, userID, transactionType, startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get category breakdown: %w", err)
	}
	defer rows.Close()

	categories := []map[string]interface{}{}
	var total float64

	for rows.Next() {
		var name, color, icon string
		var amount float64
		var count int

		err := rows.Scan(&name, &color, &icon, &amount, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}

		categories = append(categories, map[string]interface{}{
			"name":   name,
			"color":  color,
			"icon":   icon,
			"amount": amount,
			"count":  count,
		})

		total += amount
	}

	// Calculate percentages
	for _, cat := range categories {
		if total > 0 {
			cat["percentage"] = (cat["amount"].(float64) / total) * 100
		} else {
			cat["percentage"] = 0
		}
	}

	return map[string]interface{}{
		"categories": categories,
		"total":      total,
		"period":     period,
		"type":       transactionType,
	}, nil
}
