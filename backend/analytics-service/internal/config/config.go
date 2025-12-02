package config

import (
	"os"
)

type Config struct {
	// Service
	ServiceName string
	ServicePort string

	// PostgreSQL
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string

	// ClickHouse
	ClickHouseHost     string
	ClickHousePort     string
	ClickHouseUser     string
	ClickHousePassword string
	ClickHouseDB       string

	// JWT
	JWTSecret string
}

func Load() *Config {
	return &Config{
		ServiceName:        getEnv("SERVICE_NAME", "analytics-service"),
		ServicePort:        getEnv("SERVICE_PORT", "8083"),
		PostgresHost:       getEnv("POSTGRES_HOST", "postgres"),
		PostgresPort:       getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:       getEnv("POSTGRES_USER", "fintrack_user"),
		PostgresPassword:   getEnv("POSTGRES_PASSWORD", ""),
		PostgresDB:         getEnv("POSTGRES_DB", "fintrack"),
		ClickHouseHost:     getEnv("CLICKHOUSE_HOST", "clickhouse"),
		ClickHousePort:     getEnv("CLICKHOUSE_PORT", "9000"),
		ClickHouseUser:     getEnv("CLICKHOUSE_USER", "default"),
		ClickHousePassword: getEnv("CLICKHOUSE_PASSWORD", ""),
		ClickHouseDB:       getEnv("CLICKHOUSE_DB", "fintrack_analytics"),
		JWTSecret:          getEnv("JWT_SECRET", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
