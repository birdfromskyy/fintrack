package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-service/internal/config"
	"api-service/internal/database"
	"api-service/internal/handlers"
	"api-service/internal/middleware"
	"api-service/internal/services"

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

	// Connect to PostgreSQL
	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	defer db.Close()

	// Connect to Redis
	redisClient := database.ConnectRedis(cfg)
	defer redisClient.Close()

	// Initialize services
	logService := services.NewLogService(db)
	transactionService := services.NewTransactionService(db, logService)
	accountService := services.NewAccountService(db, logService)
	categoryService := services.NewCategoryService(db, logService)
	statsService := services.NewStatsService(db)

	// Initialize handlers
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	accountHandler := handlers.NewAccountHandler(accountService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	statsHandler := handlers.NewStatsHandler(statsService)
	logHandler := handlers.NewLogHandler(logService)

	// Setup Gin router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "api-service"})
	})

	// ✅ ДОБАВЬ ЭТО: Internal routes (для других микросервисов)
	internal := router.Group("/api/v1/internal")
	{
		internal.POST("/logs", logHandler.LogInternalAction)
	}

	// Protected API routes
	api := router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// Transaction routes
		api.POST("/transactions", transactionHandler.CreateTransaction)
		api.GET("/transactions", transactionHandler.GetTransactions)
		api.GET("/transactions/:id", transactionHandler.GetTransaction)
		api.PUT("/transactions/:id", transactionHandler.UpdateTransaction)
		api.DELETE("/transactions/:id", transactionHandler.DeleteTransaction)

		// Account routes
		api.POST("/accounts", accountHandler.CreateAccount)
		api.GET("/accounts", accountHandler.GetAccounts)
		api.GET("/accounts/:id", accountHandler.GetAccount)
		api.PUT("/accounts/:id", accountHandler.UpdateAccount)
		api.DELETE("/accounts/:id", accountHandler.DeleteAccount)
		api.POST("/accounts/:id/set-default", accountHandler.SetDefaultAccount)

		// Category routes
		api.POST("/categories", categoryHandler.CreateCategory)
		api.GET("/categories", categoryHandler.GetCategories)
		api.GET("/categories/:id", categoryHandler.GetCategory)
		api.PUT("/categories/:id", categoryHandler.UpdateCategory)
		api.DELETE("/categories/:id", categoryHandler.DeleteCategory)

		// Statistics routes
		api.GET("/stats/summary", statsHandler.GetSummary)
		api.GET("/stats/monthly", statsHandler.GetMonthlyStats)
		api.GET("/stats/category", statsHandler.GetCategoryStats)
		api.GET("/stats/balance-history", statsHandler.GetBalanceHistory)

		// Log routes
		api.GET("/logs", logHandler.GetMyLogs)        // Мои логи
		api.GET("/logs/stats", logHandler.GetMyStats) // Моя статистика
		api.GET("/logs/all", logHandler.GetAllLogs)   // Все логи (для админа)
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

	log.Printf("API Service started on port %s", cfg.ServicePort)

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
