package handlers

import (
	"net/http"
	"strconv"

	"analytics-service/internal/models"
	"analytics-service/internal/services"

	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	logService *services.LogService
}

func NewLogHandler(logService *services.LogService) *LogHandler {
	return &LogHandler{
		logService: logService,
	}
}

func (h *LogHandler) LogAction(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var action models.UserAction
	if err := c.ShouldBindJSON(&action); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	action.UserID = userID.(string)
	action.IP = c.ClientIP()
	action.UserAgent = c.GetHeader("User-Agent")

	err := h.logService.LogUserAction(c.Request.Context(), &action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to log action",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Action logged successfully",
	})
}

func (h *LogHandler) GetUserLogs(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	limit := 100
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user logs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"count":  len(logs),
		"limit":  limit,
		"offset": offset,
	})
}

func (h *LogHandler) GetActionStats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	days := 30
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}

	stats, err := h.logService.GetActionStats(c.Request.Context(), userID.(string), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get action stats",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}
