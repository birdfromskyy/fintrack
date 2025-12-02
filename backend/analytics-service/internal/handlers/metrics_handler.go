package handlers

import (
	"net/http"

	"analytics-service/internal/services"

	"github.com/gin-gonic/gin"
)

type MetricsHandler struct {
	metricsService *services.MetricsService
}

func NewMetricsHandler(metricsService *services.MetricsService) *MetricsHandler {
	return &MetricsHandler{
		metricsService: metricsService,
	}
}

func (h *MetricsHandler) GetDashboardMetrics(c *gin.Context) {
	metrics, err := h.metricsService.GetDashboardMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get dashboard metrics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *MetricsHandler) GetUserMetrics(c *gin.Context) {
	period := c.Query("period")
	if period == "" {
		period = "month"
	}

	metrics, err := h.metricsService.GetUserMetrics(c.Request.Context(), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user metrics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
		"period":  period,
	})
}

func (h *MetricsHandler) GetTransactionMetrics(c *gin.Context) {
	period := c.Query("period")
	if period == "" {
		period = "month"
	}

	metrics, err := h.metricsService.GetTransactionMetrics(c.Request.Context(), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get transaction metrics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *MetricsHandler) GetSystemMetrics(c *gin.Context) {
	// This endpoint is for internal monitoring
	// You might want to add authentication or IP restriction

	ctx := c.Request.Context()

	// Record request metric
	h.metricsService.RecordSystemMetric(ctx, "api_requests", 1, "endpoint:system_metrics")

	metrics, err := h.metricsService.GetDashboardMetrics(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get system metrics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"metrics": metrics,
	})
}
