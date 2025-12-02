package models

import (
	"time"
)

type Transaction struct {
	ID          string    `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	AccountID   string    `json:"account_id" db:"account_id"`
	CategoryID  string    `json:"category_id" db:"category_id"`
	Type        string    `json:"type" db:"type"` // income or expense
	Amount      float64   `json:"amount" db:"amount"`
	Description string    `json:"description" db:"description"`
	Date        time.Time `json:"date" db:"date"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Joined fields
	AccountName   string `json:"account_name,omitempty" db:"account_name"`
	CategoryName  string `json:"category_name,omitempty" db:"category_name"`
	CategoryIcon  string `json:"category_icon,omitempty" db:"category_icon"`
	CategoryColor string `json:"category_color,omitempty" db:"category_color"`
}

type CreateTransactionRequest struct {
	AccountID   string  `json:"account_id" binding:"required,uuid"`
	CategoryID  string  `json:"category_id" binding:"required,uuid"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
	Date        string  `json:"date" binding:"required"` // Changed from time.Time to string
}

type UpdateTransactionRequest struct {
	AccountID   string  `json:"account_id" binding:"omitempty,uuid"`
	CategoryID  string  `json:"category_id" binding:"omitempty,uuid"`
	Amount      float64 `json:"amount" binding:"omitempty,gt=0"`
	Description string  `json:"description"`
	Date        string  `json:"date"` // Changed from time.Time to string
}

type TransactionFilter struct {
	UserID     string
	AccountID  string
	CategoryID string
	Type       string
	DateFrom   time.Time
	DateTo     time.Time
	Limit      int
	Offset     int
}
