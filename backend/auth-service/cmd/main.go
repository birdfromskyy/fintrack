package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/database"
	"auth-service/internal/handlers"
	"auth-service/internal/middleware"
	"auth-service/internal/services"

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
	emailService := services.NewEmailService(cfg)
	authService := services.NewAuthService(db, redisClient, emailService, cfg)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Setup Gin router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "auth-service"})
	})

	// Auth routes
	api := router.Group("/api/v1/auth")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/verify-email", authHandler.VerifyEmail)
		api.POST("/login", authHandler.Login)
		api.POST("/logout", middleware.AuthMiddleware(cfg.JWTSecret), authHandler.Logout)
		api.POST("/resend-code", authHandler.ResendVerificationCode)
		api.POST("/change-password", middleware.AuthMiddleware(cfg.JWTSecret), authHandler.ChangePassword)
		api.GET("/me", middleware.AuthMiddleware(cfg.JWTSecret), authHandler.GetCurrentUser)
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

	log.Printf("Auth Service started on port %s", cfg.ServicePort)

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
