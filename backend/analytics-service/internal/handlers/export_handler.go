package handlers

import (
	"fmt"
	"net/http"
	"time"

	"analytics-service/internal/models"
	"analytics-service/internal/services"

	"github.com/gin-gonic/gin"
)

type ExportHandler struct {
	exportService *services.ExportService
}

func NewExportHandler(exportService *services.ExportService) *ExportHandler {
	return &ExportHandler{
		exportService: exportService,
	}
}

func (h *ExportHandler) ExportTransactions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	req := &models.ExportRequest{
		UserID: userID.(string),
		Format: c.Query("format"),
	}

	if req.Format == "" {
		req.Format = "csv"
	}

	// Parse dates
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			req.DateFrom = t
		}
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			req.DateTo = t
		}
	}

	if accountID := c.Query("account_id"); accountID != "" {
		req.AccountID = accountID
	}

	if txType := c.Query("type"); txType != "" {
		req.Type = txType
	}

	// Currently only CSV is supported
	if req.Format != "csv" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported format. Only CSV is supported",
		})
		return
	}

	data, err := h.exportService.ExportTransactionsCSV(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to export transactions",
			"details": err.Error(),
		})
		return
	}

	filename := fmt.Sprintf("transactions_%s_%s.csv", userID.(string), time.Now().Format("20060102_150405"))

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "text/csv", data)
}

func (h *ExportHandler) GenerateReport(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	period := c.Query("period")
	if period == "" {
		period = "month"
	}

	report, err := h.exportService.GenerateReport(c.Request.Context(), userID.(string), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate report",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"report": report,
	})
}

func (h *ExportHandler) ExportSummary(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse date range
	startDate := time.Now().AddDate(0, -1, 0) // Default: last month
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

	data, err := h.exportService.ExportSummaryCSV(c.Request.Context(), userID.(string), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to export summary",
			"details": err.Error(),
		})
		return
	}

	filename := fmt.Sprintf("summary_%s_%s_%s.csv",
		userID.(string),
		startDate.Format("20060102"),
		endDate.Format("20060102"))

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "text/csv", data)
}
