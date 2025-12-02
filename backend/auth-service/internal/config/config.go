package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Service
	ServiceName string
	ServicePort string

	// Database
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string

	// JWT
	JWTSecret      string
	JWTExpiryHours int

	// Email
	SMTPHost     string
	SMTPPort     int
	SMTPEmail    string
	SMTPPassword string

	// Verification
	EmailCodeTTL int // in minutes
}

func Load() *Config {
	jwtExpiry, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	emailCodeTTL, _ := strconv.Atoi(getEnv("EMAIL_CODE_TTL_MINUTES", "10"))

	return &Config{
		ServiceName:      getEnv("SERVICE_NAME", "auth-service"),
		ServicePort:      getEnv("SERVICE_PORT", "8081"),
		PostgresHost:     getEnv("POSTGRES_HOST", "postgres"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "fintrack_user"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", ""),
		PostgresDB:       getEnv("POSTGRES_DB", "fintrack"),
		RedisHost:        getEnv("REDIS_HOST", "redis"),
		RedisPort:        getEnv("REDIS_PORT", "6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		JWTSecret:        getEnv("JWT_SECRET", ""),
		JWTExpiryHours:   jwtExpiry,
		SMTPHost:         getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:         smtpPort,
		SMTPEmail:        getEnv("SMTP_EMAIL", ""),
		SMTPPassword:     getEnv("SMTP_PASSWORD", ""),
		EmailCodeTTL:     emailCodeTTL,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
