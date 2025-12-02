package services

import (
	"analytics-service/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type LogService struct {
	clickhouseDB driver.Conn
}

func NewLogService(clickhouseDB driver.Conn) *LogService {
	return &LogService{
		clickhouseDB: clickhouseDB,
	}
}

func (s *LogService) LogUserAction(ctx context.Context, action *models.UserAction) error {
	if s.clickhouseDB == nil {
		return fmt.Errorf("ClickHouse connection is nil")
	}

	query := `
        INSERT INTO user_actions (user_id, action, entity, entity_id, details, ip, user_agent)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `

	err := s.clickhouseDB.Exec(ctx, query,
		action.UserID,
		action.Action,
		action.Entity,
		action.EntityID,
		action.Details,
		action.IP,
		action.UserAgent,
	)

	if err != nil {
		return fmt.Errorf("failed to insert user action: %w", err)
	}

	return nil
}

func (s *LogService) GetUserLogs(ctx context.Context, userID string, limit int, offset int) ([]*models.UserAction, error) {
	if s.clickhouseDB == nil {
		return nil, fmt.Errorf("ClickHouse connection is nil")
	}

	query := `
        SELECT 
            toString(id) as id,
            user_id,
            action,
            entity,
            entity_id,
            details,
            ip,
            user_agent,
            timestamp
        FROM user_actions
        WHERE user_id = ?
        ORDER BY timestamp DESC
        LIMIT ? OFFSET ?
    `

	rows, err := s.clickhouseDB.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query user actions: %w", err)
	}
	defer rows.Close()

	var logs []*models.UserAction
	for rows.Next() {
		var log models.UserAction
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.Entity,
			&log.EntityID,
			&log.Details,
			&log.IP,
			&log.UserAgent,
			&log.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		logs = append(logs, &log)
	}

	return logs, nil
}

func (s *LogService) GetActionStats(ctx context.Context, userID string, days int) (map[string]interface{}, error) {
	if s.clickhouseDB == nil {
		return nil, fmt.Errorf("ClickHouse connection is nil")
	}

	startDate := time.Now().AddDate(0, 0, -days)

	query := `
        SELECT 
            action,
            entity,
            count() as count,
            max(timestamp) as last_action
        FROM user_actions
        WHERE user_id = ? AND date >= ?
        GROUP BY action, entity
        ORDER BY count DESC
    `

	rows, err := s.clickhouseDB.Query(ctx, query, userID, startDate.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to query stats: %w", err)
	}
	defer rows.Close()

	stats := make([]map[string]interface{}, 0)
	for rows.Next() {
		var action, entity string
		var count uint64
		var lastAction time.Time

		err := rows.Scan(&action, &entity, &count, &lastAction)
		if err != nil {
			return nil, err
		}

		stats = append(stats, map[string]interface{}{
			"action":      action,
			"entity":      entity,
			"count":       count,
			"last_action": lastAction,
		})
	}

	return map[string]interface{}{
		"period": fmt.Sprintf("%d days", days),
		"stats":  stats,
	}, nil
}

func (s *LogService) StartCleanupWorker(ctx context.Context) {
	// TTL уже настроен в таблице — автоочистка работает сама
	// Можно добавить дополнительную логику при необходимости
}
