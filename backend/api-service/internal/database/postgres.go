package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"api-service/internal/config"

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
		`CREATE TABLE IF NOT EXISTS transactions (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
            category_id UUID NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
            type VARCHAR(20) NOT NULL CHECK (type IN ('income', 'expense')),
            amount DECIMAL(15, 2) NOT NULL CHECK (amount > 0),
            description TEXT,
            date DATE NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );`,

		`CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_category_id ON transactions(category_id);`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date);`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(type);`,

		`DROP TRIGGER IF EXISTS update_transactions_updated_at ON transactions;`,
		`CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions
								FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();`,
		`CREATE TABLE IF NOT EXISTS user_actions (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				action VARCHAR(50) NOT NULL,
				entity VARCHAR(50) NOT NULL,
				entity_id UUID,
				details TEXT,
				ip VARCHAR(45),
				user_agent TEXT,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,

		`CREATE INDEX IF NOT EXISTS idx_user_actions_user_id ON user_actions(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_user_actions_created_at ON user_actions(created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_user_actions_action ON user_actions(action);`,
		`CREATE INDEX IF NOT EXISTS idx_user_actions_entity ON user_actions(entity);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}
