package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"auth-service/internal/config"

	_ "github.com/lib/pq"
)

func ConnectPostgres(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresDB,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create tables if not exists
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL")
	return db, nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,

		`CREATE TABLE IF NOT EXISTS users (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            email VARCHAR(255) UNIQUE NOT NULL,
            password_hash VARCHAR(255) NOT NULL,
            verified BOOLEAN DEFAULT FALSE,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );`,

		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`,

		`CREATE TABLE IF NOT EXISTS email_verifications (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            code VARCHAR(6) NOT NULL,
            expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );`,

		`CREATE INDEX IF NOT EXISTS idx_email_verifications_user_id ON email_verifications(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_email_verifications_code ON email_verifications(code);`,

		`CREATE TABLE IF NOT EXISTS accounts (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            name VARCHAR(100) NOT NULL,
            balance DECIMAL(15, 2) DEFAULT 0,
            is_default BOOLEAN DEFAULT FALSE,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );`,

		`CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);`,

		`CREATE TABLE IF NOT EXISTS categories (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            user_id UUID REFERENCES users(id) ON DELETE CASCADE,
            name VARCHAR(100) NOT NULL,
            type VARCHAR(20) NOT NULL CHECK (type IN ('income', 'expense')),
            icon VARCHAR(50),
            color VARCHAR(7),
            is_system BOOLEAN DEFAULT FALSE,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );`,

		`CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_categories_type ON categories(type);`,

		// Trigger for updating updated_at
		`CREATE OR REPLACE FUNCTION update_updated_at_column()
        RETURNS TRIGGER AS $$
        BEGIN
            NEW.updated_at = CURRENT_TIMESTAMP;
            RETURN NEW;
        END;
        $$ language 'plpgsql';`,

		`DROP TRIGGER IF EXISTS update_users_updated_at ON users;`,
		`CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();`,

		`DROP TRIGGER IF EXISTS update_accounts_updated_at ON accounts;`,
		`CREATE TRIGGER update_accounts_updated_at BEFORE UPDATE ON accounts
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	// Insert default categories if they don't exist
	if err := insertDefaultCategories(db); err != nil {
		return fmt.Errorf("failed to insert default categories: %w", err)
	}

	return nil
}

func insertDefaultCategories(db *sql.DB) error {
	// Check if default categories already exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM categories WHERE is_system = true").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // Default categories already exist
	}

	defaultCategories := []struct {
		Name  string
		Type  string
		Icon  string
		Color string
	}{
		// Income categories
		{"Зарплата", "income", "💰", "#4CAF50"},
		{"Фриланс", "income", "💻", "#8BC34A"},
		{"Инвестиции", "income", "📈", "#00BCD4"},
		{"Подарки", "income", "🎁", "#E91E63"},
		{"Другие доходы", "income", "💵", "#9C27B0"},

		// Expense categories
		{"Продукты", "expense", "🛒", "#FF5722"},
		{"Транспорт", "expense", "🚗", "#795548"},
		{"Жилье", "expense", "🏠", "#607D8B"},
		{"Развлечения", "expense", "🎮", "#FF9800"},
		{"Здоровье", "expense", "💊", "#F44336"},
		{"Одежда", "expense", "👕", "#3F51B5"},
		{"Образование", "expense", "📚", "#009688"},
		{"Рестораны", "expense", "🍽️", "#FFC107"},
		{"Коммунальные услуги", "expense", "💡", "#9E9E9E"},
		{"Другие расходы", "expense", "📦", "#673AB7"},
	}

	query := `INSERT INTO categories (name, type, icon, color, is_system) VALUES ($1, $2, $3, $4, true)`

	for _, cat := range defaultCategories {
		if _, err := db.Exec(query, cat.Name, cat.Type, cat.Icon, cat.Color); err != nil {
			return fmt.Errorf("failed to insert category %s: %w", cat.Name, err)
		}
	}

	log.Println("Default categories inserted successfully")
	return nil
}
