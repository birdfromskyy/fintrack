package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"analytics-service/internal/config"
	"analytics-service/internal/database"
	"analytics-service/internal/handlers"
	"analytics-service/internal/middleware"
	"analytics-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Connect to PostgreSQL only (skip ClickHouse for now)
	postgresDB, err := database.ConnectPostgres(cfg)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	defer postgresDB.Close()

	// Initialize services (without ClickHouse)
	analyticsService := services.NewAnalyticsService(postgresDB, nil)
	logService := services.NewLogService(nil)
	exportService := services.NewExportService(postgresDB, nil)
	metricsService := services.NewMetricsService(nil)

	// Initialize handlers
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	logHandler := handlers.NewLogHandler(logService)
	exportHandler := handlers.NewExportHandler(exportService)
	metricsHandler := handlers.NewMetricsHandler(metricsService)

	// Setup Gin router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "analytics-service"})
	})

	// Public routes for Grafana
	public := router.Group("/api/v1/metrics")
	{
		public.GET("/dashboard", metricsHandler.GetDashboardMetrics)
		public.GET("/users", metricsHandler.GetUserMetrics)
		public.GET("/transactions", metricsHandler.GetTransactionMetrics)
		public.GET("/system", metricsHandler.GetSystemMetrics)
	}

	// Protected API routes
	api := router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// Analytics routes
		api.GET("/analytics/overview", analyticsHandler.GetOverview)
		api.GET("/analytics/trends", analyticsHandler.GetTrends)
		api.GET("/analytics/forecast", analyticsHandler.GetForecast)
		api.GET("/analytics/insights", analyticsHandler.GetInsights)
		api.GET("/analytics/cashflow", analyticsHandler.GetCashflow)

		// Export routes
		api.GET("/export/transactions", exportHandler.ExportTransactions)
		api.GET("/export/report", exportHandler.GenerateReport)
		api.GET("/export/summary", exportHandler.ExportSummary)

		// Log routes (for tracking user actions)
		api.POST("/logs/action", logHandler.LogAction)
		api.GET("/logs/user", logHandler.GetUserLogs)
	}

	// Server setup
	srv := &http.Server{
		Addr:    ":" + cfg.ServicePort,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Printf("Analytics Service started on port %s (PostgreSQL only mode)", cfg.ServicePort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
