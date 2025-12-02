package clickhouse

import (
	"context"
	"fmt"
	"log"
	"time"

	"analytics-service/internal/config"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

func Connect(cfg *config.Config) (driver.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", cfg.ClickHouseHost, cfg.ClickHousePort)},
		Auth: clickhouse.Auth{
			Database: cfg.ClickHouseDB,
			Username: cfg.ClickHouseUser,
			Password: cfg.ClickHousePassword,
		},
		Debug: false,
		Debugf: func(format string, v ...interface{}) {
			log.Printf(format, v...)
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:      time.Duration(10) * time.Second,
		MaxOpenConns:     5,
		MaxIdleConns:     5,
		ConnMaxLifetime:  time.Duration(10) * time.Minute,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
		BlockBufferSize:  10,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	ctx := context.Background()
	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	log.Println("Successfully connected to ClickHouse")
	return conn, nil
}

func InitializeTables(conn driver.Conn) error {
	ctx := context.Background()

	queries := []string{
		// User actions table
		`CREATE TABLE IF NOT EXISTS user_actions (
            id UUID DEFAULT generateUUIDv4(),
            user_id String,
            action String,
            entity String,
            entity_id String,
            details String,
            ip String,
            user_agent String,
            timestamp DateTime DEFAULT now(),
            date Date DEFAULT toDate(timestamp)
        ) ENGINE = MergeTree()
        PARTITION BY toYYYYMM(date)
        ORDER BY (user_id, timestamp)
        TTL date + INTERVAL 90 DAY`,

		// Transaction analytics table
		`CREATE TABLE IF NOT EXISTS transaction_analytics (
            id UUID DEFAULT generateUUIDv4(),
            user_id String,
            account_id String,
            category_id String,
            category_name String,
            type String,
            amount Float64,
            date Date,
            timestamp DateTime DEFAULT now(),
            hour UInt8 DEFAULT toHour(timestamp),
            weekday UInt8 DEFAULT toDayOfWeek(date)
        ) ENGINE = MergeTree()
        PARTITION BY toYYYYMM(date)
        ORDER BY (user_id, date, category_id)`,

		// Aggregated daily stats
		`CREATE TABLE IF NOT EXISTS daily_stats (
            date Date,
            user_id String,
            total_income Float64,
            total_expense Float64,
            transaction_count UInt32,
            unique_categories UInt32,
            max_transaction Float64,
            min_transaction Float64,
            avg_transaction Float64
        ) ENGINE = SummingMergeTree()
        PARTITION BY toYYYYMM(date)
        ORDER BY (user_id, date)`,

		// Category performance
		`CREATE TABLE IF NOT EXISTS category_performance (
            date Date,
            user_id String,
            category_id String,
            category_name String,
            type String,
            total_amount Float64,
            transaction_count UInt32,
            avg_amount Float64,
            trend Float64
        ) ENGINE = ReplacingMergeTree()
        PARTITION BY toYYYYMM(date)
        ORDER BY (user_id, date, category_id)`,

		// User behavior patterns
		`CREATE TABLE IF NOT EXISTS user_patterns (
            user_id String,
            pattern_type String,
            pattern_value String,
            frequency UInt32,
            last_occurrence DateTime,
            confidence Float64
        ) ENGINE = ReplacingMergeTree()
        ORDER BY (user_id, pattern_type, pattern_value)`,

		// System metrics
		`CREATE TABLE IF NOT EXISTS system_metrics (
            timestamp DateTime,
            metric_name String,
            metric_value Float64,
            tags String
        ) ENGINE = MergeTree()
        PARTITION BY toYYYYMM(timestamp)
        ORDER BY (metric_name, timestamp)
        TTL timestamp + INTERVAL 30 DAY`,
	}

	for _, query := range queries {
		if err := conn.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	log.Println("ClickHouse tables initialized successfully")
	return nil
}
