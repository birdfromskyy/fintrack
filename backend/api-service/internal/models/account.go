package models

import (
	"time"
)

type Account struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	Balance   float64   `json:"balance" db:"balance"`
	IsDefault bool      `json:"is_default" db:"is_default"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateAccountRequest struct {
	Name      string  `json:"name" binding:"required,min=1,max=100"`
	Balance   float64 `json:"balance"`
	IsDefault bool    `json:"is_default"`
}

type UpdateAccountRequest struct {
	Name    string  `json:"name" binding:"omitempty,min=1,max=100"`
	Balance float64 `json:"balance"`
}

type AccountStats struct {
	TotalIncome    float64 `json:"total_income"`
	TotalExpense   float64 `json:"total_expense"`
	CurrentBalance float64 `json:"current_balance"`
}
