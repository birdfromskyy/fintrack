package services

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type MetricsService struct {
	clickhouseDB driver.Conn
}

func NewMetricsService(clickhouseDB driver.Conn) *MetricsService {
	return &MetricsService{
		clickhouseDB: clickhouseDB,
	}
}

func (s *MetricsService) GetDashboardMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Total users (from user_actions)
	var totalUsers int64
	err := s.clickhouseDB.QueryRow(ctx,
		`SELECT COUNT(DISTINCT user_id) FROM user_actions WHERE timestamp >= now() - INTERVAL 30 DAY`,
	).Scan(&totalUsers)
	if err == nil {
		metrics["active_users_30d"] = totalUsers
	}

	// Daily active users
	var dailyActiveUsers int64
	err = s.clickhouseDB.QueryRow(ctx,
		`SELECT COUNT(DISTINCT user_id) FROM user_actions WHERE timestamp >= today()`,
	).Scan(&dailyActiveUsers)
	if err == nil {
		metrics["daily_active_users"] = dailyActiveUsers
	}

	// Total transactions today (from transaction_analytics)
	var todayTransactions int64
	err = s.clickhouseDB.QueryRow(ctx,
		`SELECT COUNT(*) FROM transaction_analytics WHERE date = today()`,
	).Scan(&todayTransactions)
	if err == nil {
		metrics["today_transactions"] = todayTransactions
	}

	// Average transaction amount
	var avgAmount float64
	err = s.clickhouseDB.QueryRow(ctx,
		`SELECT AVG(amount) FROM transaction_analytics WHERE date >= today() - 30`,
	).Scan(&avgAmount)
	if err == nil {
		metrics["avg_transaction_amount"] = avgAmount
	}

	// System health metrics
	systemMetrics, _ := s.getSystemMetrics(ctx)
	if systemMetrics != nil {
		metrics["system"] = systemMetrics
	}

	return metrics, nil
}

func (s *MetricsService) GetUserMetrics(ctx context.Context, period string) ([]map[string]interface{}, error) {
	days := 30
	switch period {
	case "week":
		days = 7
	case "month":
		days = 30
	case "quarter":
		days = 90
	case "year":
		days = 365
	}

	query := `
        SELECT 
            toDate(timestamp) as date,
            COUNT(DISTINCT user_id) as active_users,
            COUNT(*) as total_actions
        FROM user_actions
        WHERE timestamp >= now() - INTERVAL ? DAY
        GROUP BY date
        ORDER BY date DESC`

	rows, err := s.clickhouseDB.Query(ctx, query, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get user metrics: %w", err)
	}
	defer rows.Close()

	var metrics []map[string]interface{}
	for rows.Next() {
		var date time.Time
		var activeUsers, totalActions int64

		err := rows.Scan(&date, &activeUsers, &totalActions)
		if err != nil {
			continue
		}

		metrics = append(metrics, map[string]interface{}{
			"date":          date.Format("2006-01-02"),
			"active_users":  activeUsers,
			"total_actions": totalActions,
		})
	}

	return metrics, nil
}

func (s *MetricsService) GetTransactionMetrics(ctx context.Context, period string) (map[string]interface{}, error) {
	days := 30
	switch period {
	case "day":
		days = 1
	case "week":
		days = 7
	case "month":
		days = 30
	case "year":
		days = 365
	}

	// Volume metrics
	volumeQuery := `
        SELECT 
            type,
            SUM(amount) as total_amount,
            COUNT(*) as count,
            AVG(amount) as avg_amount,
            MAX(amount) as max_amount,
            MIN(amount) as min_amount
        FROM transaction_analytics
        WHERE date >= today() - ?
        GROUP BY type`

	rows, err := s.clickhouseDB.Query(ctx, volumeQuery, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get volume metrics: %w", err)
	}
	defer rows.Close()

	volumeMetrics := make(map[string]interface{})
	for rows.Next() {
		var txType string
		var totalAmount, avgAmount, maxAmount, minAmount float64
		var count int64

		err := rows.Scan(&txType, &totalAmount, &count, &avgAmount, &maxAmount, &minAmount)
		if err != nil {
			continue
		}

		volumeMetrics[txType] = map[string]interface{}{
			"total":   totalAmount,
			"count":   count,
			"average": avgAmount,
			"max":     maxAmount,
			"min":     minAmount,
		}
	}

	// Hourly distribution
	hourlyQuery := `
        SELECT 
            hour,
            COUNT(*) as count
        FROM transaction_analytics
        WHERE date >= today() - ?
        GROUP BY hour
        ORDER BY hour`

	rows, err = s.clickhouseDB.Query(ctx, hourlyQuery, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get hourly distribution: %w", err)
	}
	defer rows.Close()

	hourlyDist := make([]map[string]interface{}, 0, 24)
	for rows.Next() {
		var hour uint8
		var count int64

		err := rows.Scan(&hour, &count)
		if err != nil {
			continue
		}

		hourlyDist = append(hourlyDist, map[string]interface{}{
			"hour":  hour,
			"count": count,
		})
	}

	// Category distribution
	categoryQuery := `
        SELECT 
            category_name,
            type,
            SUM(amount) as total,
            COUNT(*) as count
        FROM transaction_analytics
        WHERE date >= today() - ?
        GROUP BY category_name, type
        ORDER BY total DESC
        LIMIT 20`

	rows, err = s.clickhouseDB.Query(ctx, categoryQuery, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get category distribution: %w", err)
	}
	defer rows.Close()

	var categoryDist []map[string]interface{}
	for rows.Next() {
		var categoryName, txType string
		var total float64
		var count int64

		err := rows.Scan(&categoryName, &txType, &total, &count)
		if err != nil {
			continue
		}

		categoryDist = append(categoryDist, map[string]interface{}{
			"category": categoryName,
			"type":     txType,
			"total":    total,
			"count":    count,
		})
	}

	return map[string]interface{}{
		"period":                period,
		"volume":                volumeMetrics,
		"hourly_distribution":   hourlyDist,
		"category_distribution": categoryDist,
	}, nil
}

func (s *MetricsService) getSystemMetrics(ctx context.Context) (map[string]interface{}, error) {
	// Get latest system metrics
	query := `
        SELECT 
            metric_name,
            AVG(metric_value) as avg_value,
            MAX(metric_value) as max_value,
            MIN(metric_value) as min_value
        FROM system_metrics
        WHERE timestamp >= now() - INTERVAL 1 HOUR
        GROUP BY metric_name`

	rows, err := s.clickhouseDB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := make(map[string]interface{})
	for rows.Next() {
		var metricName string
		var avgValue, maxValue, minValue float64

		err := rows.Scan(&metricName, &avgValue, &maxValue, &minValue)
		if err != nil {
			continue
		}

		metrics[metricName] = map[string]interface{}{
			"average": avgValue,
			"max":     maxValue,
			"min":     minValue,
		}
	}

	return metrics, nil
}

func (s *MetricsService) RecordSystemMetric(ctx context.Context, name string, value float64, tags string) error {
	query := `INSERT INTO system_metrics (timestamp, metric_name, metric_value, tags) VALUES (?, ?, ?, ?)`

	err := s.clickhouseDB.Exec(ctx, query, time.Now(), name, value, tags)
	if err != nil {
		return fmt.Errorf("failed to record system metric: %w", err)
	}

	return nil
}
