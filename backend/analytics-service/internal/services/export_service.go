package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"time"

	"analytics-service/internal/models"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ExportService struct {
	postgresDB   *sql.DB
	clickhouseDB driver.Conn
}

func NewExportService(postgresDB *sql.DB, clickhouseDB driver.Conn) *ExportService {
	return &ExportService{
		postgresDB:   postgresDB,
		clickhouseDB: clickhouseDB,
	}
}

func (s *ExportService) ExportTransactionsCSV(ctx context.Context, req *models.ExportRequest) ([]byte, error) {
	query := `
        SELECT 
            t.id,
            t.date,
            c.name as category,
            c.type,
            t.amount,
            t.description,
            a.name as account
        FROM transactions t
        JOIN categories c ON t.category_id = c.id
        JOIN accounts a ON t.account_id = a.id
        WHERE t.user_id = $1`

	args := []interface{}{req.UserID}
	argNum := 1

	if !req.DateFrom.IsZero() {
		argNum++
		query += fmt.Sprintf(" AND t.date >= $%d", argNum)
		args = append(args, req.DateFrom)
	}

	if !req.DateTo.IsZero() {
		argNum++
		query += fmt.Sprintf(" AND t.date <= $%d", argNum)
		args = append(args, req.DateTo)
	}

	if req.AccountID != "" {
		argNum++
		query += fmt.Sprintf(" AND t.account_id = $%d", argNum)
		args = append(args, req.AccountID)
	}

	if req.Type != "" {
		argNum++
		query += fmt.Sprintf(" AND c.type = $%d", argNum)
		args = append(args, req.Type)
	}

	query += " ORDER BY t.date DESC, t.created_at DESC"

	rows, err := s.postgresDB.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("Failed to query transactions: %v", err)
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	// Create CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"ID", "Дата", "Категория", "Тип", "Сумма", "Описание", "Счет"}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data
	rowCount := 0
	for rows.Next() {
		var (
			id          string
			date        time.Time
			category    string
			txType      string
			amount      float64
			description sql.NullString
			account     string
		)

		err := rows.Scan(&id, &date, &category, &txType, &amount, &description, &account)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		// ✅ Форматируем сумму с учётом типа (расходы с минусом)
		amountFormatted := fmt.Sprintf("%.2f", amount)
		if txType == "expense" {
			amountFormatted = fmt.Sprintf("-%.2f", amount)
		}

		record := []string{
			id,
			date.Format("2006-01-02"),
			category,
			txType,
			amountFormatted, // ← ИСПРАВЛЕНО
			description.String,
			account,
		}

		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write CSV record: %w", err)
		}
		rowCount++
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	log.Printf("Exported %d transactions for user %s", rowCount, req.UserID)

	// Log export action
	s.logExport(ctx, req.UserID, "transactions", req.Format)

	return buf.Bytes(), nil
}

func (s *ExportService) GenerateReport(ctx context.Context, userID string, period string) (*models.Report, error) {
	report := &models.Report{
		Title:       fmt.Sprintf("Финансовый отчет - %s", period),
		Period:      period,
		GeneratedAt: time.Now(),
		Summary:     make(map[string]interface{}),
		Charts:      []models.ChartData{},
		Tables:      []models.TableData{},
	}

	// Parse period
	var startDate, endDate time.Time
	now := time.Now()

	switch period {
	case "week":
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case "month":
		startDate = now.AddDate(0, -1, 0)
		endDate = now
	case "quarter":
		startDate = now.AddDate(0, -3, 0)
		endDate = now
	case "year":
		startDate = now.AddDate(-1, 0, 0)
		endDate = now
	default:
		startDate = now.AddDate(0, -1, 0)
		endDate = now
	}

	// Get summary data
	summary, err := s.getSummaryData(ctx, userID, startDate, endDate)
	if err == nil {
		report.Summary = summary
	}

	// Get category breakdown chart data
	categoryChart, err := s.getCategoryChartData(ctx, userID, startDate, endDate)
	if err == nil {
		report.Charts = append(report.Charts, *categoryChart)
	}

	// Get trend chart data
	trendChart, err := s.getTrendChartData(ctx, userID, startDate, endDate)
	if err == nil {
		report.Charts = append(report.Charts, *trendChart)
	}

	// Get top transactions table
	topTransactions, err := s.getTopTransactionsTable(ctx, userID, startDate, endDate)
	if err == nil {
		report.Tables = append(report.Tables, *topTransactions)
	}

	// Get account balances table
	accountsTable, err := s.getAccountsTable(ctx, userID)
	if err == nil {
		report.Tables = append(report.Tables, *accountsTable)
	}

	// Log report generation
	s.logExport(ctx, userID, "report", "pdf")

	return report, nil
}

func (s *ExportService) ExportSummaryCSV(ctx context.Context, userID string, startDate, endDate time.Time) ([]byte, error) {
	// Get summary statistics
	var totalIncome, totalExpense float64
	var transactionCount int

	err := s.postgresDB.QueryRowContext(ctx, `
        SELECT 
            COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0),
            COUNT(*)
        FROM transactions
        WHERE user_id = $1 AND date >= $2 AND date <= $3`,
		userID, startDate, endDate).Scan(&totalIncome, &totalExpense, &transactionCount)

	if err != nil {
		return nil, fmt.Errorf("failed to get summary: %w", err)
	}

	// Get category breakdown
	categoryQuery := `
        SELECT 
            c.name,
            c.type,
            SUM(t.amount) as total,
            COUNT(t.id) as count
        FROM transactions t
        JOIN categories c ON t.category_id = c.id
        WHERE t.user_id = $1 AND t.date >= $2 AND t.date <= $3
        GROUP BY c.id, c.name, c.type
        ORDER BY c.type, total DESC`

	rows, err := s.postgresDB.QueryContext(ctx, categoryQuery, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	// Create CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write summary header
	writer.Write([]string{"Сводка за период", startDate.Format("2006-01-02"), "по", endDate.Format("2006-01-02")})
	writer.Write([]string{})
	writer.Write([]string{"Показатель", "Значение"})
	writer.Write([]string{"Общий доход", fmt.Sprintf("%.2f", totalIncome)})
	writer.Write([]string{"Общий расход", fmt.Sprintf("%.2f", totalExpense)})
	writer.Write([]string{"Баланс", fmt.Sprintf("%.2f", totalIncome-totalExpense)})
	writer.Write([]string{"Количество транзакций", fmt.Sprintf("%d", transactionCount)})
	writer.Write([]string{})

	// Write category breakdown
	writer.Write([]string{"Категория", "Тип", "Сумма", "Количество"})

	for rows.Next() {
		var name, txType string
		var total float64
		var count int

		err := rows.Scan(&name, &txType, &total, &count)
		if err != nil {
			continue
		}

		writer.Write([]string{name, txType, fmt.Sprintf("%.2f", total), fmt.Sprintf("%d", count)})
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	// Log export
	s.logExport(ctx, userID, "summary", "csv")

	return buf.Bytes(), nil
}

// Helper functions
func (s *ExportService) getSummaryData(ctx context.Context, userID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	var totalIncome, totalExpense float64
	var transactionCount int
	var uniqueCategories int

	err := s.postgresDB.QueryRowContext(ctx, `
        SELECT 
            COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0),
            COUNT(*),
            COUNT(DISTINCT category_id)
        FROM transactions
        WHERE user_id = $1 AND date >= $2 AND date <= $3`,
		userID, startDate, endDate).Scan(&totalIncome, &totalExpense, &transactionCount, &uniqueCategories)

	if err != nil {
		return nil, err
	}

	savingsRate := 0.0
	if totalIncome > 0 {
		savingsRate = ((totalIncome - totalExpense) / totalIncome) * 100
	}

	return map[string]interface{}{
		"total_income":      totalIncome,
		"total_expense":     totalExpense,
		"net_income":        totalIncome - totalExpense,
		"savings_rate":      savingsRate,
		"transaction_count": transactionCount,
		"unique_categories": uniqueCategories,
		"avg_transaction":   (totalIncome + totalExpense) / float64(transactionCount),
		"period_start":      startDate.Format("2006-01-02"),
		"period_end":        endDate.Format("2006-01-02"),
	}, nil
}

func (s *ExportService) getCategoryChartData(ctx context.Context, userID string, startDate, endDate time.Time) (*models.ChartData, error) {
	query := `
        SELECT 
            c.name,
            c.type,
            SUM(t.amount) as total
        FROM transactions t
        JOIN categories c ON t.category_id = c.id
        WHERE t.user_id = $1 AND t.date >= $2 AND t.date <= $3
        GROUP BY c.id, c.name, c.type
        ORDER BY total DESC
        LIMIT 10`

	rows, err := s.postgresDB.QueryContext(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chart := &models.ChartData{
		Type:   "pie",
		Title:  "Распределение по категориям",
		Labels: []string{},
		Data:   []map[string]interface{}{},
	}

	for rows.Next() {
		var name, txType string
		var total float64

		err := rows.Scan(&name, &txType, &total)
		if err != nil {
			continue
		}

		chart.Labels = append(chart.Labels, name)
		chart.Data = append(chart.Data, map[string]interface{}{
			"value": total,
			"type":  txType,
		})
	}

	return chart, nil
}

func (s *ExportService) getTrendChartData(ctx context.Context, userID string, startDate, endDate time.Time) (*models.ChartData, error) {
	query := `
        SELECT 
            DATE(date) as day,
            SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) as income,
            SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END) as expense
        FROM transactions
        WHERE user_id = $1 AND date >= $2 AND date <= $3
        GROUP BY DATE(date)
        ORDER BY day`

	rows, err := s.postgresDB.QueryContext(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chart := &models.ChartData{
		Type:   "line",
		Title:  "Динамика доходов и расходов",
		Labels: []string{},
		Data:   []map[string]interface{}{},
	}

	incomeData := []float64{}
	expenseData := []float64{}

	for rows.Next() {
		var day time.Time
		var income, expense float64

		err := rows.Scan(&day, &income, &expense)
		if err != nil {
			continue
		}

		chart.Labels = append(chart.Labels, day.Format("02.01"))
		incomeData = append(incomeData, income)
		expenseData = append(expenseData, expense)
	}

	chart.Data = append(chart.Data, map[string]interface{}{
		"label": "Доходы",
		"data":  incomeData,
	})
	chart.Data = append(chart.Data, map[string]interface{}{
		"label": "Расходы",
		"data":  expenseData,
	})

	return chart, nil
}

func (s *ExportService) getTopTransactionsTable(ctx context.Context, userID string, startDate, endDate time.Time) (*models.TableData, error) {
	query := `
        SELECT 
            t.date,
            c.name,
            t.amount,
            t.description,
            a.name
        FROM transactions t
        JOIN categories c ON t.category_id = c.id
        JOIN accounts a ON t.account_id = a.id
        WHERE t.user_id = $1 AND t.date >= $2 AND t.date <= $3
        ORDER BY t.amount DESC
        LIMIT 20`

	rows, err := s.postgresDB.QueryContext(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	table := &models.TableData{
		Title:   "Топ-20 транзакций",
		Headers: []string{"Дата", "Категория", "Сумма", "Описание", "Счет"},
		Rows:    [][]interface{}{},
	}

	for rows.Next() {
		var date time.Time
		var category string
		var amount float64
		var description sql.NullString
		var account string

		err := rows.Scan(&date, &category, &amount, &description, &account)
		if err != nil {
			continue
		}

		row := []interface{}{
			date.Format("2006-01-02"),
			category,
			fmt.Sprintf("%.2f", amount),
			description.String,
			account,
		}
		table.Rows = append(table.Rows, row)
	}

	return table, nil
}

func (s *ExportService) getAccountsTable(ctx context.Context, userID string) (*models.TableData, error) {
	query := `
        SELECT 
            name,
            balance,
            is_default,
            created_at
        FROM accounts
        WHERE user_id = $1
        ORDER BY is_default DESC, balance DESC`

	rows, err := s.postgresDB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	table := &models.TableData{
		Title:   "Счета",
		Headers: []string{"Название", "Баланс", "Основной", "Создан"},
		Rows:    [][]interface{}{},
	}

	for rows.Next() {
		var name string
		var balance float64
		var isDefault bool
		var createdAt time.Time

		err := rows.Scan(&name, &balance, &isDefault, &createdAt)
		if err != nil {
			continue
		}

		defaultStr := "Нет"
		if isDefault {
			defaultStr = "Да"
		}

		row := []interface{}{
			name,
			fmt.Sprintf("%.2f", balance),
			defaultStr,
			createdAt.Format("2006-01-02"),
		}
		table.Rows = append(table.Rows, row)
	}

	return table, nil
}

func (s *ExportService) logExport(ctx context.Context, userID, exportType, format string) {
	// ✅ Проверяем что ClickHouse подключён
	if s.clickhouseDB == nil {
		log.Printf("ClickHouse not connected, skipping export log")
		return
	}

	query := `INSERT INTO user_actions (user_id, action, entity, details) VALUES (?, ?, ?, ?)`
	details := fmt.Sprintf("type: %s, format: %s", exportType, format)
	err := s.clickhouseDB.Exec(ctx, query, userID, "export", exportType, details)
	if err != nil {
		log.Printf("Failed to log export action: %v", err)
	}
}
