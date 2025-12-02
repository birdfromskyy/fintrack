package services

import (
	"context"
	"database/sql"
	"fmt"
)

type LogService struct {
	db *sql.DB
}

func NewLogService(db *sql.DB) *LogService {
	return &LogService{db: db}
}

type UserAction struct {
	UserID    string
	Action    string // create, update, delete, view
	Entity    string // transaction, account, category
	EntityID  string
	Details   string
	IP        string
	UserAgent string
}

func (s *LogService) Log(ctx context.Context, action *UserAction) error {
	query := `
		INSERT INTO user_actions (user_id, action, entity, entity_id, details, ip, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := s.db.ExecContext(ctx, query,
		action.UserID,
		action.Action,
		action.Entity,
		action.EntityID,
		action.Details,
		action.IP,
		action.UserAgent,
	)

	if err != nil {
		// Не возвращаем ошибку — логирование не должно ломать основной флоу
		fmt.Printf("Failed to log action: %v\n", err)
	}

	return nil
}

func (s *LogService) GetUserLogs(ctx context.Context, userID string, limit, offset int) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			id,
			user_id,
			action,
			entity,
			entity_id,
			details,
			ip,
			user_agent,
			created_at
		FROM user_actions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	logs := []map[string]interface{}{}
	for rows.Next() {
		var (
			id, userID, action, entity, entityID, details, ip, userAgent string
			createdAt                                                    interface{}
		)

		err := rows.Scan(&id, &userID, &action, &entity, &entityID, &details, &ip, &userAgent, &createdAt)
		if err != nil {
			return nil, err
		}

		logs = append(logs, map[string]interface{}{
			"id":         id,
			"user_id":    userID,
			"action":     action,
			"entity":     entity,
			"entity_id":  entityID,
			"details":    details,
			"ip":         ip,
			"user_agent": userAgent,
			"created_at": createdAt,
		})
	}

	return logs, nil
}

func (s *LogService) GetAllLogs(ctx context.Context, limit, offset int, filters map[string]string) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			ua.id,
			ua.user_id,
			u.email,
			ua.action,
			ua.entity,
			ua.entity_id,
			ua.details,
			ua.ip,
			ua.created_at
		FROM user_actions ua
		LEFT JOIN users u ON ua.user_id = u.id
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	if action := filters["action"]; action != "" {
		query += fmt.Sprintf(" AND ua.action = $%d", argCount)
		args = append(args, action)
		argCount++
	}

	if entity := filters["entity"]; entity != "" {
		query += fmt.Sprintf(" AND ua.entity = $%d", argCount)
		args = append(args, entity)
		argCount++
	}

	if userID := filters["user_id"]; userID != "" {
		query += fmt.Sprintf(" AND ua.user_id = $%d", argCount)
		args = append(args, userID)
		argCount++
	}

	query += fmt.Sprintf(" ORDER BY ua.created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query all logs: %w", err)
	}
	defer rows.Close()

	logs := []map[string]interface{}{}
	for rows.Next() {
		var (
			id, userID, email, action, entity, entityID, details, ip string
			createdAt                                                interface{}
		)

		err := rows.Scan(&id, &userID, &email, &action, &entity, &entityID, &details, &ip, &createdAt)
		if err != nil {
			return nil, err
		}

		logs = append(logs, map[string]interface{}{
			"id":         id,
			"user_id":    userID,
			"email":      email,
			"action":     action,
			"entity":     entity,
			"entity_id":  entityID,
			"details":    details,
			"ip":         ip,
			"created_at": createdAt,
		})
	}

	return logs, nil
}

func (s *LogService) GetStats(ctx context.Context, userID string, days int) (map[string]interface{}, error) {
	query := `
		SELECT 
			action,
			entity,
			COUNT(*) as count
		FROM user_actions
		WHERE user_id = $1
		  AND created_at >= NOW() - INTERVAL '%d days'
		GROUP BY action, entity
		ORDER BY count DESC
	`

	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(query, days), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query stats: %w", err)
	}
	defer rows.Close()

	stats := []map[string]interface{}{}
	for rows.Next() {
		var action, entity string
		var count int

		err := rows.Scan(&action, &entity, &count)
		if err != nil {
			return nil, err
		}

		stats = append(stats, map[string]interface{}{
			"action": action,
			"entity": entity,
			"count":  count,
		})
	}

	return map[string]interface{}{
		"period": fmt.Sprintf("%d days", days),
		"stats":  stats,
	}, nil
}
