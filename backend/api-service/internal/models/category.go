package models

import (
	"time"
)

type Category struct {
	ID        string    `json:"id" db:"id"`
	UserID    *string   `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	Type      string    `json:"type" db:"type"` // income or expense
	Icon      string    `json:"icon" db:"icon"`
	Color     string    `json:"color" db:"color"`
	IsSystem  bool      `json:"is_system" db:"is_system"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateCategoryRequest struct {
	Name  string `json:"name" binding:"required,min=1,max=100"`
	Type  string `json:"type" binding:"required,oneof=income expense"`
	Icon  string `json:"icon"`
	Color string `json:"color" binding:"required"`
}

type UpdateCategoryRequest struct {
	Name  string `json:"name" binding:"omitempty,min=1,max=100"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

type CategoryStats struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Type         string  `json:"type"`
	Total        float64 `json:"total"`
	Count        int     `json:"count"`
	Percentage   float64 `json:"percentage"`
}
