package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"

	"analytics-service/internal/models"
)

type AnalyticsService struct {
	postgresDB   *sql.DB
	clickhouseDB interface{}
}

func NewAnalyticsService(postgresDB *sql.DB, clickhouseDB interface{}) *AnalyticsService {
	return &AnalyticsService{
		postgresDB:   postgresDB,
		clickhouseDB: clickhouseDB,
	}
}

func (s *AnalyticsService) GetOverview(ctx context.Context, userID string, period string) (*models.Overview, error) {
	log.Printf("=== GetOverview START for user %s, period %s ===", userID, period)

	overview := &models.Overview{
		Period:          period,
		TotalIncome:     0,
		TotalExpense:    0,
		NetIncome:       0,
		SavingsRate:     0,
		TopCategories:   []models.CategoryStat{},
		AccountBalances: []models.AccountBalance{},
	}

	// Get current date and calculate period
	now := time.Now()
	var startDate, endDate time.Time

	switch period {
	case "week":
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case "month":
		// First day of current month
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = now
	case "quarter":
		startDate = now.AddDate(0, -3, 0)
		endDate = now
	case "year":
		startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		endDate = now
	default:
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = now
	}

	log.Printf("Date range: %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// First, let's check if we have any transactions at all
	var totalTransactions int
	checkQuery := `SELECT COUNT(*) FROM transactions WHERE user_id = $1`
	err := s.postgresDB.QueryRowContext(ctx, checkQuery, userID).Scan(&totalTransactions)
	if err != nil {
		log.Printf("Error counting total transactions: %v", err)
	} else {
		log.Printf("User has %d total transactions", totalTransactions)
	}

	// Get income and expense for the period
	query := `
        SELECT 
            COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) as total_income,
            COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) as total_expense,
            COUNT(*) as period_transactions
        FROM transactions
        WHERE user_id = $1 
        AND date >= $2 
        AND date <= $3`

	var periodTransactions int
	err = s.postgresDB.QueryRowContext(
		ctx,
		query,
		userID,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	).Scan(
		&overview.TotalIncome,
		&overview.TotalExpense,
		&periodTransactions,
	)

	if err != nil {
		log.Printf("ERROR getting income/expense: %v", err)
		log.Printf("Query: %s", query)
		log.Printf("Params: userID=%s, startDate=%s, endDate=%s",
			userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	} else {
		log.Printf("Period transactions: %d, Income: %.2f, Expense: %.2f",
			periodTransactions, overview.TotalIncome, overview.TotalExpense)
	}

	// Calculate net income and savings rate
	overview.NetIncome = overview.TotalIncome - overview.TotalExpense
	if overview.TotalIncome > 0 {
		overview.SavingsRate = (overview.NetIncome / overview.TotalIncome) * 100
	}

	// Get top categories for the period
	categoryQuery := `
        SELECT 
            c.id,
            c.name,
            c.type,
            c.icon,
            c.color,
            COALESCE(SUM(t.amount), 0) as total_amount,
            COUNT(t.id) as transaction_count
        FROM categories c
        LEFT JOIN transactions t ON c.id = t.category_id 
            AND t.user_id = $1 
            AND t.date >= $2 
            AND t.date <= $3
        WHERE (c.user_id = $1 OR c.is_system = true)
        GROUP BY c.id, c.name, c.type, c.icon, c.color
        HAVING COUNT(t.id) > 0
        ORDER BY total_amount DESC
        LIMIT 10`

	rows, err := s.postgresDB.QueryContext(
		ctx,
		categoryQuery,
		userID,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)

	if err != nil {
		log.Printf("ERROR getting categories: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var stat models.CategoryStat
			var icon, color sql.NullString

			err := rows.Scan(
				&stat.CategoryID,
				&stat.CategoryName,
				&stat.Type,
				&icon,
				&color,
				&stat.Amount,
				&stat.Count,
			)
			if err != nil {
				log.Printf("Error scanning category: %v", err)
				continue
			}

			// Calculate percentage
			if stat.Type == "income" && overview.TotalIncome > 0 {
				stat.Percentage = (stat.Amount / overview.TotalIncome) * 100
			} else if stat.Type == "expense" && overview.TotalExpense > 0 {
				stat.Percentage = (stat.Amount / overview.TotalExpense) * 100
			}

			log.Printf("Category: %s, Type: %s, Amount: %.2f, Count: %d",
				stat.CategoryName, stat.Type, stat.Amount, stat.Count)

			overview.TopCategories = append(overview.TopCategories, stat)
		}
		log.Printf("Found %d categories with transactions", len(overview.TopCategories))
	}

	// Get account balances
	accountQuery := `
        SELECT 
            a.id,
            a.name,
            a.balance
        FROM accounts a
        WHERE a.user_id = $1
        ORDER BY a.is_default DESC, a.balance DESC`

	rows, err = s.postgresDB.QueryContext(ctx, accountQuery, userID)
	if err != nil {
		log.Printf("ERROR getting accounts: %v", err)
	} else {
		defer rows.Close()
		var totalBalance float64
		for rows.Next() {
			var balance models.AccountBalance
			err := rows.Scan(&balance.AccountID, &balance.AccountName, &balance.Balance)
			if err != nil {
				log.Printf("Error scanning account: %v", err)
				continue
			}
			totalBalance += balance.Balance
			overview.AccountBalances = append(overview.AccountBalances, balance)
		}

		// Calculate percentages
		for i := range overview.AccountBalances {
			if totalBalance > 0 {
				overview.AccountBalances[i].Percentage = (overview.AccountBalances[i].Balance / totalBalance) * 100
			}
		}
		log.Printf("Found %d accounts, total balance: %.2f", len(overview.AccountBalances), totalBalance)
	}

	// Get month comparison
	comparison, err := s.getMonthComparison(ctx, userID, startDate)
	if err != nil {
		log.Printf("Error getting month comparison: %v", err)
	} else {
		overview.MonthComparison = comparison
		if comparison != nil {
			log.Printf("Month comparison - Income change: %.2f%%, Expense change: %.2f%%",
				comparison.IncomeChange, comparison.ExpenseChange)
		}
	}

	log.Printf("=== GetOverview END - Income: %.2f, Expense: %.2f, Categories: %d ===",
		overview.TotalIncome, overview.TotalExpense, len(overview.TopCategories))

	return overview, nil
}

func (s *AnalyticsService) GetTrends(ctx context.Context, userID string, days int) ([]*models.Trend, error) {
	log.Printf("=== GetTrends START for user %s, days %d ===", userID, days)

	if days <= 0 {
		days = 30
	}

	trends := []*models.Trend{}
	startDate := time.Now().AddDate(0, 0, -days)

	// Create a map of all dates in range
	dateMap := make(map[string]*models.Trend)
	for d := startDate; !d.After(time.Now()); d = d.AddDate(0, 0, 1) {
		dateMap[d.Format("2006-01-02")] = &models.Trend{
			Date:    d,
			Income:  0,
			Expense: 0,
			Balance: 0,
		}
	}

	// Get transactions for the period
	query := `
        SELECT 
            date,
            type,
            SUM(amount) as total
        FROM transactions
        WHERE user_id = $1 AND date >= $2
        GROUP BY date, type
        ORDER BY date`

	rows, err := s.postgresDB.QueryContext(ctx, query, userID, startDate.Format("2006-01-02"))
	if err != nil {
		log.Printf("Error getting trends: %v", err)
		return trends, nil
	}
	defer rows.Close()

	// Fill in actual transaction data
	transactionCount := 0
	for rows.Next() {
		var date time.Time
		var txType string
		var amount float64

		err := rows.Scan(&date, &txType, &amount)
		if err != nil {
			log.Printf("Error scanning trend: %v", err)
			continue
		}

		dateStr := date.Format("2006-01-02")
		if trend, exists := dateMap[dateStr]; exists {
			if txType == "income" {
				trend.Income = amount
			} else {
				trend.Expense = amount
			}
			transactionCount++
		}
	}

	log.Printf("Found transactions for %d days", transactionCount)

	// Convert map to sorted slice and calculate running balance
	var runningBalance float64

	// Get initial balance
	var initialBalance float64
	err = s.postgresDB.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE user_id = $1`,
		userID).Scan(&initialBalance)
	if err == nil {
		runningBalance = initialBalance
	}

	// Sort dates and create trends array
	for d := startDate; !d.After(time.Now()); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		if trend, exists := dateMap[dateStr]; exists {
			runningBalance += trend.Income - trend.Expense
			trend.Balance = runningBalance
			trends = append(trends, trend)
		}
	}

	log.Printf("=== GetTrends END - Generated %d trend points ===", len(trends))
	return trends, nil
}

func (s *AnalyticsService) GetForecast(ctx context.Context, userID string, months int) (*models.Forecast, error) {
	log.Printf("=== GetForecast START for user %s, months %d ===", userID, months)

	if months <= 0 {
		months = 3
	}

	forecast := &models.Forecast{
		Period:           fmt.Sprintf("%d месяца", months),
		PredictedIncome:  0,
		PredictedExpense: 0,
		PredictedBalance: 0,
		Confidence:       0,
		BasedOnMonths:    0,
	}

	// Get historical monthly data
	query := `
        SELECT 
            DATE_TRUNC('month', date) as month,
            SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) as income,
            SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END) as expense
        FROM transactions
        WHERE user_id = $1
        GROUP BY DATE_TRUNC('month', date)
        ORDER BY month DESC
        LIMIT 6`

	rows, err := s.postgresDB.QueryContext(ctx, query, userID)
	if err != nil {
		log.Printf("Error getting forecast data: %v", err)
		return forecast, nil
	}
	defer rows.Close()

	var monthlyData []struct {
		Income  float64
		Expense float64
	}

	for rows.Next() {
		var month time.Time
		var income, expense float64
		err := rows.Scan(&month, &income, &expense)
		if err != nil {
			continue
		}
		monthlyData = append(monthlyData, struct {
			Income  float64
			Expense float64
		}{income, expense})
		log.Printf("Month: %s, Income: %.2f, Expense: %.2f",
			month.Format("2006-01"), income, expense)
	}

	if len(monthlyData) == 0 {
		log.Printf("No historical data for forecast")
		return forecast, nil
	}

	// Calculate averages
	var totalIncome, totalExpense float64
	for _, data := range monthlyData {
		totalIncome += data.Income
		totalExpense += data.Expense
	}

	avgMonthlyIncome := totalIncome / float64(len(monthlyData))
	avgMonthlyExpense := totalExpense / float64(len(monthlyData))

	forecast.PredictedIncome = avgMonthlyIncome * float64(months)
	forecast.PredictedExpense = avgMonthlyExpense * float64(months)
	forecast.PredictedBalance = forecast.PredictedIncome - forecast.PredictedExpense
	forecast.BasedOnMonths = len(monthlyData)

	// Calculate confidence based on data availability
	if len(monthlyData) >= 6 {
		forecast.Confidence = 85
	} else if len(monthlyData) >= 3 {
		forecast.Confidence = 65
	} else if len(monthlyData) >= 1 {
		forecast.Confidence = 45
	} else {
		forecast.Confidence = 20
	}

	log.Printf("=== GetForecast END - Predicted Income: %.2f, Expense: %.2f, Confidence: %.0f%% ===",
		forecast.PredictedIncome, forecast.PredictedExpense, forecast.Confidence)

	return forecast, nil
}

func (s *AnalyticsService) GetInsights(ctx context.Context, userID string) ([]*models.Insight, error) {
	log.Printf("=== GetInsights START for user %s ===", userID)

	insights := []*models.Insight{}

	// Insight 1: Biggest expense category this month
	var categoryName string
	var categoryAmount float64
	err := s.postgresDB.QueryRowContext(ctx, `
        SELECT c.name, SUM(t.amount) as total
        FROM transactions t
        JOIN categories c ON t.category_id = c.id
        WHERE t.user_id = $1 
        AND t.type = 'expense' 
        AND t.date >= DATE_TRUNC('month', CURRENT_DATE)
        GROUP BY c.id, c.name
        ORDER BY total DESC
        LIMIT 1`,
		userID).Scan(&categoryName, &categoryAmount)

	if err == nil && categoryAmount > 0 {
		insights = append(insights, &models.Insight{
			Type:  "expense_analysis",
			Title: "Наибольшие расходы",
			Description: fmt.Sprintf("Категория '%s' - ваши наибольшие расходы в этом месяце (%.0f ₽)",
				categoryName, categoryAmount),
			Value:    categoryAmount,
			Priority: "high",
			Date:     time.Now(),
		})
		log.Printf("Added insight: Biggest expense - %s: %.2f", categoryName, categoryAmount)
	}

	// Insight 2: Spending trend comparison
	var currentMonthExpense, lastMonthExpense float64
	err = s.postgresDB.QueryRowContext(ctx, `
        SELECT 
            COALESCE(SUM(CASE WHEN date >= DATE_TRUNC('month', CURRENT_DATE) 
                THEN amount END), 0) as current_month,
            COALESCE(SUM(CASE WHEN date >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month') 
                AND date < DATE_TRUNC('month', CURRENT_DATE) 
                THEN amount END), 0) as last_month
        FROM transactions
        WHERE user_id = $1 AND type = 'expense'`,
		userID).Scan(&currentMonthExpense, &lastMonthExpense)

	if err == nil {
		if lastMonthExpense > 0 {
			change := ((currentMonthExpense - lastMonthExpense) / lastMonthExpense) * 100
			trend := &models.Insight{
				Type:  "trend_analysis",
				Title: "Динамика расходов",
				Value: change,
				Date:  time.Now(),
			}

			if change > 20 {
				trend.Description = fmt.Sprintf("Расходы увеличились на %.0f%% по сравнению с прошлым месяцем", change)
				trend.Priority = "high"
			} else if change < -10 {
				trend.Description = fmt.Sprintf("Отлично! Расходы снизились на %.0f%% по сравнению с прошлым месяцем",
					math.Abs(change))
				trend.Priority = "low"
			} else {
				trend.Description = fmt.Sprintf("Расходы изменились на %.0f%% по сравнению с прошлым месяцем", change)
				trend.Priority = "medium"
			}

			insights = append(insights, trend)
			log.Printf("Added insight: Spending trend - %.2f%%", change)
		}
	}

	// Insight 3: Savings rate
	var monthIncome, monthExpense float64
	err = s.postgresDB.QueryRowContext(ctx, `
        SELECT 
            COALESCE(SUM(CASE WHEN type = 'income' THEN amount END), 0),
            COALESCE(SUM(CASE WHEN type = 'expense' THEN amount END), 0)
        FROM transactions
        WHERE user_id = $1 AND date >= DATE_TRUNC('month', CURRENT_DATE)`,
		userID).Scan(&monthIncome, &monthExpense)

	if err == nil && monthIncome > 0 {
		savingsRate := ((monthIncome - monthExpense) / monthIncome) * 100
		savings := &models.Insight{
			Type:  "savings_analysis",
			Title: "Уровень сбережений",
			Value: savingsRate,
			Date:  time.Now(),
		}

		if savingsRate > 30 {
			savings.Description = fmt.Sprintf("Отлично! Вы сберегаете %.0f%% от доходов", savingsRate)
			savings.Priority = "low"
		} else if savingsRate > 10 {
			savings.Description = fmt.Sprintf("Вы сберегаете %.0f%% от доходов", savingsRate)
			savings.Priority = "medium"
		} else if savingsRate > 0 {
			savings.Description = fmt.Sprintf("Вы сберегаете только %.0f%% от доходов. Попробуйте увеличить", savingsRate)
			savings.Priority = "high"
		} else {
			savings.Description = "Внимание! Ваши расходы превышают доходы"
			savings.Priority = "high"
		}

		insights = append(insights, savings)
		log.Printf("Added insight: Savings rate - %.2f%%", savingsRate)
	}

	log.Printf("=== GetInsights END - Generated %d insights ===", len(insights))
	return insights, nil
}

func (s *AnalyticsService) GetCashflow(ctx context.Context, userID string, startDate, endDate time.Time) ([]*models.Cashflow, error) {
	// Simplified implementation
	return []*models.Cashflow{}, nil
}

func (s *AnalyticsService) getMonthComparison(ctx context.Context, userID string, currentStart time.Time) (*models.Comparison, error) {
	prevStart := currentStart.AddDate(0, -1, 0)
	prevEnd := currentStart.AddDate(0, 0, -1)

	var currIncome, currExpense, prevIncome, prevExpense float64

	// Current period
	err := s.postgresDB.QueryRowContext(ctx, `
        SELECT 
            COALESCE(SUM(CASE WHEN type = 'income' THEN amount END), 0),
            COALESCE(SUM(CASE WHEN type = 'expense' THEN amount END), 0)
        FROM transactions
        WHERE user_id = $1 AND date >= $2`,
		userID, currentStart.Format("2006-01-02")).Scan(&currIncome, &currExpense)

	if err != nil {
		return nil, err
	}

	// Previous period
	err = s.postgresDB.QueryRowContext(ctx, `
        SELECT 
            COALESCE(SUM(CASE WHEN type = 'income' THEN amount END), 0),
            COALESCE(SUM(CASE WHEN type = 'expense' THEN amount END), 0)
        FROM transactions
        WHERE user_id = $1 AND date >= $2 AND date <= $3`,
		userID, prevStart.Format("2006-01-02"), prevEnd.Format("2006-01-02")).Scan(&prevIncome, &prevExpense)

	if err != nil {
		return nil, err
	}

	comparison := &models.Comparison{
		IncomeDiff:    currIncome - prevIncome,
		ExpenseDiff:   currExpense - prevExpense,
		IncomeChange:  0,
		ExpenseChange: 0,
	}

	if prevIncome > 0 {
		comparison.IncomeChange = ((currIncome - prevIncome) / prevIncome) * 100
	}
	if prevExpense > 0 {
		comparison.ExpenseChange = ((currExpense - prevExpense) / prevExpense) * 100
	}

	return comparison, nil
}

func (s *AnalyticsService) StartAggregationWorker(ctx context.Context) {
	// Disabled
}
