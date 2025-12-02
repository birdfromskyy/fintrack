package handlers

import (
	"net/http"
	"strconv"

	"api-service/internal/services"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsService *services.StatsService
}

func NewStatsHandler(statsService *services.StatsService) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
	}
}

func (h *StatsHandler) GetSummary(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Disable caching for real-time data
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	summary, err := h.statsService.GetSummary(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": summary,
	})
}

func (h *StatsHandler) GetMonthlyStats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	months := 12 // Default to 12 months
	if m := c.Query("months"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed > 0 && parsed <= 24 {
			months = parsed
		}
	}

	stats, err := h.statsService.GetMonthlyStats(c.Request.Context(), userID.(string), months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":  stats,
		"months": months,
	})
}

func (h *StatsHandler) GetCategoryStats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	period := c.Query("period")
	if period == "" {
		period = "month"
	}

	transactionType := c.Query("type")
	if transactionType == "" {
		transactionType = "expense"
	}

	if transactionType != "income" && transactionType != "expense" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction type"})
		return
	}

	breakdown, err := h.statsService.GetCategoryBreakdown(c.Request.Context(), userID.(string), transactionType, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, breakdown)
}

func (h *StatsHandler) GetBalanceHistory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	days := 30 // Default to 30 days
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}

	history, err := h.statsService.GetBalanceHistory(c.Request.Context(), userID.(string), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"days":    days,
	})
}
