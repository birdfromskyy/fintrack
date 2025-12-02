package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"analytics-service/internal/services"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
}

func NewAnalyticsHandler(analyticsService *services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

func (h *AnalyticsHandler) GetOverview(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	period := c.Query("period")
	if period == "" {
		period = "month"
	}

	overview, err := h.analyticsService.GetOverview(c.Request.Context(), userID.(string), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get overview",
			"details": err.Error(),
		})
		return
	}

	// Логирование для отладки
	log.Printf("Sending overview response: Income=%.2f, Expense=%.2f, Categories=%d",
		overview.TotalIncome, overview.TotalExpense, len(overview.TopCategories))

	// Отправляем данные напрямую, без вложенности
	c.JSON(http.StatusOK, overview)
}

func (h *AnalyticsHandler) GetTrends(c *gin.Context) {
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

	trends, err := h.analyticsService.GetTrends(c.Request.Context(), userID.(string), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get trends",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Sending trends response: %d points", len(trends))

	// Отправляем массив напрямую
	c.JSON(http.StatusOK, trends)
}

func (h *AnalyticsHandler) GetForecast(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	months := 3
	if m := c.Query("months"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed > 0 {
			months = parsed
		}
	}

	forecast, err := h.analyticsService.GetForecast(c.Request.Context(), userID.(string), months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get forecast",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Sending forecast response: Income=%.2f, Expense=%.2f",
		forecast.PredictedIncome, forecast.PredictedExpense)

	// Отправляем объект напрямую
	c.JSON(http.StatusOK, forecast)
}

func (h *AnalyticsHandler) GetInsights(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	insights, err := h.analyticsService.GetInsights(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get insights",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Sending insights response: %d insights", len(insights))

	// Отправляем массив напрямую
	c.JSON(http.StatusOK, insights)
}

func (h *AnalyticsHandler) GetCashflow(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse date range
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	if start := c.Query("start_date"); start != "" {
		if parsed, err := time.Parse("2006-01-02", start); err == nil {
			startDate = parsed
		}
	}

	if end := c.Query("end_date"); end != "" {
		if parsed, err := time.Parse("2006-01-02", end); err == nil {
			endDate = parsed
		}
	}

	cashflow, err := h.analyticsService.GetCashflow(c.Request.Context(), userID.(string), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get cashflow",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, cashflow)
}
