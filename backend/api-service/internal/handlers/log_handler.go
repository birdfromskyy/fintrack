package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"api-service/internal/services"

	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	logService *services.LogService
}

func NewLogHandler(logService *services.LogService) *LogHandler {
	return &LogHandler{logService: logService}
}

// GetMyLogs - получить свои логи (для авторизованного пользователя)
func (h *LogHandler) GetMyLogs(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	logs, err := h.logService.GetUserLogs(c.Request.Context(), userID.(string), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Парсим JSON details для читаемости
	for i, log := range logs {
		if details, ok := log["details"].(string); ok && details != "" {
			var parsedDetails map[string]interface{}
			if err := json.Unmarshal([]byte(details), &parsedDetails); err == nil {
				logs[i]["details_parsed"] = parsedDetails
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"count":  len(logs),
		"limit":  limit,
		"offset": offset,
	})
}

// GetMyStats - получить статистику своих действий
func (h *LogHandler) GetMyStats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	days := 30
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	stats, err := h.logService.GetStats(c.Request.Context(), userID.(string), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetAllLogs - получить все логи (для админа)
func (h *LogHandler) GetAllLogs(c *gin.Context) {
	// TODO: Добавь проверку что пользователь — админ
	// userID, _ := c.Get("userID")
	// if !isAdmin(userID) { return error }

	limit := 100
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	filters := map[string]string{
		"action":  c.Query("action"),
		"entity":  c.Query("entity"),
		"user_id": c.Query("user_id"),
	}

	logs, err := h.logService.GetAllLogs(c.Request.Context(), limit, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"count":  len(logs),
		"limit":  limit,
		"offset": offset,
	})
}

// LogInternalAction - для логирования из других микросервисов (auth, analytics)
func (h *LogHandler) LogInternalAction(c *gin.Context) {
	// Проверяем что запрос от внутреннего сервиса
	if c.GetHeader("X-Internal-Service") == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	var req struct {
		UserID   string                 `json:"user_id"`
		Action   string                 `json:"action"`
		Entity   string                 `json:"entity"`
		EntityID string                 `json:"entity_id"`
		Data     map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Преобразуем data в JSON string
	detailsJSON, _ := json.Marshal(req.Data)

	err := h.logService.Log(c.Request.Context(), &services.UserAction{
		UserID:   req.UserID,
		Action:   req.Action,
		Entity:   req.Entity,
		EntityID: req.EntityID,
		Details:  string(detailsJSON),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log action"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged successfully"})
}
