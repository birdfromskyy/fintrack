package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"api-service/internal/models"

	"github.com/google/uuid"
)

type CategoryService struct {
	db *sql.DB
}

func NewCategoryService(db *sql.DB) *CategoryService {
	return &CategoryService{db: db}
}

func (s *CategoryService) CreateCategory(ctx context.Context, userID string, req *models.CreateCategoryRequest) (*models.Category, error) {
	// Check if category with same name already exists for user
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1 AND user_id = $2 AND type = $3)`,
		req.Name, userID, req.Type).Scan(&exists)

	if err != nil {
		return nil, fmt.Errorf("failed to check category existence: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("category with this name already exists")
	}

	category := &models.Category{
		ID:        uuid.New().String(),
		UserID:    &userID,
		Name:      req.Name,
		Type:      req.Type,
		Icon:      req.Icon,
		Color:     req.Color,
		IsSystem:  false,
		CreatedAt: time.Now(),
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO categories (id, user_id, name, type, icon, color, is_system, created_at) 
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		category.ID, category.UserID, category.Name, category.Type,
		category.Icon, category.Color, category.IsSystem, category.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

func (s *CategoryService) GetCategories(ctx context.Context, userID string) ([]*models.Category, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, name, type, icon, color, is_system, created_at 
         FROM categories 
         WHERE user_id = $1 OR is_system = true 
         ORDER BY is_system DESC, type, name`,
		userID)

	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		var c models.Category
		err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Type,
			&c.Icon, &c.Color, &c.IsSystem, &c.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, &c)
	}

	return categories, nil
}

func (s *CategoryService) GetCategoriesByType(ctx context.Context, userID, categoryType string) ([]*models.Category, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, name, type, icon, color, is_system, created_at 
         FROM categories 
         WHERE (user_id = $1 OR is_system = true) AND type = $2
         ORDER BY is_system DESC, name`,
		userID, categoryType)

	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		var c models.Category
		err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Type,
			&c.Icon, &c.Color, &c.IsSystem, &c.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, &c)
	}

	return categories, nil
}

func (s *CategoryService) GetCategory(ctx context.Context, userID, categoryID string) (*models.Category, error) {
	var c models.Category
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, type, icon, color, is_system, created_at 
         FROM categories 
         WHERE id = $1 AND (user_id = $2 OR is_system = true)`,
		categoryID, userID).Scan(&c.ID, &c.UserID, &c.Name, &c.Type,
		&c.Icon, &c.Color, &c.IsSystem, &c.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &c, nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, userID, categoryID string, req *models.UpdateCategoryRequest) (*models.Category, error) {
	// Check if category exists and belongs to user (not system)
	var isSystem bool
	var ownerID *string
	err := s.db.QueryRowContext(ctx,
		`SELECT is_system, user_id FROM categories WHERE id = $1`,
		categoryID).Scan(&isSystem, &ownerID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to check category: %w", err)
	}

	if isSystem {
		return nil, fmt.Errorf("cannot modify system category")
	}

	if ownerID == nil || *ownerID != userID {
		return nil, fmt.Errorf("category not found")
	}

	// Build update query
	updateFields := make(map[string]interface{})
	if req.Name != "" {
		updateFields["name"] = req.Name
	}
	if req.Icon != "" {
		updateFields["icon"] = req.Icon
	}
	if req.Color != "" {
		updateFields["color"] = req.Color
	}

	if len(updateFields) == 0 {
		return s.GetCategory(ctx, userID, categoryID)
	}

	// Execute update
	query := `UPDATE categories SET `
	args := []interface{}{}
	i := 1

	for field, value := range updateFields {
		if i > 1 {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", field, i)
		args = append(args, value)
		i++
	}

	query += fmt.Sprintf(" WHERE id = $%d AND user_id = $%d", i, i+1)
	args = append(args, categoryID, userID)

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return s.GetCategory(ctx, userID, categoryID)
}

func (s *CategoryService) DeleteCategory(ctx context.Context, userID, categoryID string) error {
	// Check if category exists and belongs to user (not system)
	var isSystem bool
	var ownerID *string
	err := s.db.QueryRowContext(ctx,
		`SELECT is_system, user_id FROM categories WHERE id = $1`,
		categoryID).Scan(&isSystem, &ownerID)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("category not found")
		}
		return fmt.Errorf("failed to check category: %w", err)
	}

	if isSystem {
		return fmt.Errorf("cannot delete system category")
	}

	if ownerID == nil || *ownerID != userID {
		return fmt.Errorf("category not found")
	}

	// Check if category has transactions
	var transactionCount int
	err = s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM transactions WHERE category_id = $1`,
		categoryID).Scan(&transactionCount)

	if err != nil {
		return fmt.Errorf("failed to check transactions: %w", err)
	}

	if transactionCount > 0 {
		return fmt.Errorf("cannot delete category with existing transactions")
	}

	// Delete category
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM categories WHERE id = $1 AND user_id = $2`,
		categoryID, userID)

	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

func (s *CategoryService) GetCategoryStats(ctx context.Context, userID string, startDate, endDate time.Time) ([]*models.CategoryStats, error) {
	query := `
        SELECT 
            c.id as category_id,
            c.name as category_name,
            c.type,
            COALESCE(SUM(t.amount), 0) as total,
            COUNT(t.id) as count
        FROM categories c
        LEFT JOIN transactions t ON c.id = t.category_id 
            AND t.user_id = $1 
            AND t.date >= $2 
            AND t.date <= $3
        WHERE c.user_id = $1 OR c.is_system = true
        GROUP BY c.id, c.name, c.type
        HAVING COUNT(t.id) > 0
        ORDER BY total DESC`

	rows, err := s.db.QueryContext(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get category stats: %w", err)
	}
	defer rows.Close()

	var stats []*models.CategoryStats
	var totalExpense, totalIncome float64

	// First pass to collect data and calculate totals
	for rows.Next() {
		var stat models.CategoryStats
		err := rows.Scan(&stat.CategoryID, &stat.CategoryName, &stat.Type, &stat.Total, &stat.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category stat: %w", err)
		}

		if stat.Type == "expense" {
			totalExpense += stat.Total
		} else {
			totalIncome += stat.Total
		}

		stats = append(stats, &stat)
	}

	// Second pass to calculate percentages
	for _, stat := range stats {
		if stat.Type == "expense" && totalExpense > 0 {
			stat.Percentage = (stat.Total / totalExpense) * 100
		} else if stat.Type == "income" && totalIncome > 0 {
			stat.Percentage = (stat.Total / totalIncome) * 100
		}
	}

	return stats, nil
}
