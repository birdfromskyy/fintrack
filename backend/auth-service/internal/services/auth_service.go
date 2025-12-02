package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"auth-service/internal/config"
	"auth-service/internal/models"
	"auth-service/pkg/jwt"
)

type AuthService struct {
	db           *sql.DB
	redis        *redis.Client
	emailService *EmailService
	config       *config.Config
}

func NewAuthService(db *sql.DB, redis *redis.Client, emailService *EmailService, cfg *config.Config) *AuthService {
	return &AuthService{
		db:           db,
		redis:        redis,
		emailService: emailService,
		config:       cfg,
	}
}

func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	// Check if user already exists
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	if exists {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create user
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Verified:     false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO users (id, email, password_hash, verified, created_at, updated_at) 
         VALUES ($1, $2, $3, $4, $5, $6)`,
		user.ID, user.Email, user.PasswordHash, user.Verified, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create default account for user
	accountID := uuid.New().String()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO accounts (id, user_id, name, balance, is_default) 
         VALUES ($1, $2, $3, $4, $5)`,
		accountID, user.ID, "Основной счёт", 0.00, true)

	if err != nil {
		return nil, fmt.Errorf("failed to create default account: %w", err)
	}

	// Copy system categories for the new user
	_, err = tx.ExecContext(ctx,
		`INSERT INTO categories (user_id, name, type, icon, color, is_system)
         SELECT $1, name, type, icon, color, false
         FROM categories 
         WHERE is_system = true`,
		user.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to copy default categories: %w", err)
	}

	// Generate verification code
	code := s.generateVerificationCode()
	expiresAt := time.Now().Add(time.Duration(s.config.EmailCodeTTL) * time.Minute)

	_, err = tx.ExecContext(ctx,
		`INSERT INTO email_verifications (user_id, code, expires_at) 
         VALUES ($1, $2, $3)`,
		user.ID, code, expiresAt)

	if err != nil {
		return nil, fmt.Errorf("failed to save verification code: %w", err)
	}

	// Store code in Redis with TTL
	redisKey := fmt.Sprintf("email_verification:%s", user.Email)
	err = s.redis.Set(ctx, redisKey, code, time.Duration(s.config.EmailCodeTTL)*time.Minute).Err()
	if err != nil {
		log.Printf("Failed to store verification code in Redis: %v", err)
	}

	// Send verification email
	if err := s.emailService.SendVerificationCode(user.Email, code); err != nil {
		log.Printf("Failed to send verification email: %v", err)
		// Don't return error, user is created
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// ✅ Логируем создание дефолтного аккаунта (асинхронно)
	go s.logDefaultAccountCreation(user.ID, accountID)

	return user, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, req *models.VerifyEmailRequest) error {
	// Get code from Redis first
	redisKey := fmt.Sprintf("email_verification:%s", req.Email)
	storedCode, err := s.redis.Get(ctx, redisKey).Result()

	if err == nil && storedCode == req.Code {
		// Code is valid in Redis
		_, err = s.db.ExecContext(ctx,
			`UPDATE users SET verified = true WHERE email = $1`,
			req.Email)

		if err != nil {
			return fmt.Errorf("failed to verify user: %w", err)
		}

		// Delete code from Redis
		s.redis.Del(ctx, redisKey)

		// Delete from database
		s.db.ExecContext(ctx,
			`DELETE FROM email_verifications 
             WHERE user_id = (SELECT id FROM users WHERE email = $1)`,
			req.Email)

		return nil
	}

	// Fallback to database check
	var userID string
	var expiresAt time.Time

	err = s.db.QueryRowContext(ctx,
		`SELECT ev.user_id, ev.expires_at 
         FROM email_verifications ev
         JOIN users u ON u.id = ev.user_id
         WHERE u.email = $1 AND ev.code = $2
         ORDER BY ev.created_at DESC
         LIMIT 1`,
		req.Email, req.Code).Scan(&userID, &expiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("invalid verification code")
		}
		return fmt.Errorf("failed to verify code: %w", err)
	}

	if time.Now().After(expiresAt) {
		return errors.New("verification code has expired")
	}

	// Update user as verified
	_, err = s.db.ExecContext(ctx,
		`UPDATE users SET verified = true WHERE id = $1`,
		userID)

	if err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}

	// Delete verification code
	s.db.ExecContext(ctx,
		`DELETE FROM email_verifications WHERE user_id = $1`,
		userID)

	return nil
}

func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	// Get user by email
	var user models.User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, verified, created_at, updated_at 
         FROM users WHERE email = $1`,
		req.Email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Verified, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid email or password")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check if user is verified
	if !user.Verified {
		return nil, errors.New("email not verified")
	}

	// Generate JWT token
	token, err := jwt.GenerateToken(user.ID, user.Email, s.config.JWTSecret, s.config.JWTExpiryHours)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Store session in Redis
	sessionKey := fmt.Sprintf("session:%s", user.ID)
	sessionData := fmt.Sprintf(`{"user_id":"%s","email":"%s","login_time":"%s"}`,
		user.ID, user.Email, time.Now().Format(time.RFC3339))

	err = s.redis.Set(ctx, sessionKey, sessionData,
		time.Duration(s.config.JWTExpiryHours)*time.Hour).Err()
	if err != nil {
		log.Printf("Failed to store session in Redis: %v", err)
	}

	return &models.AuthResponse{
		Token: token,
		User:  &user,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID string) error {
	// Remove session from Redis
	sessionKey := fmt.Sprintf("session:%s", userID)
	err := s.redis.Del(ctx, sessionKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (s *AuthService) ResendVerificationCode(ctx context.Context, email string) error {
	// Check if user exists and not verified
	var userID string
	var verified bool

	err := s.db.QueryRowContext(ctx,
		`SELECT id, verified FROM users WHERE email = $1`,
		email).Scan(&userID, &verified)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	if verified {
		return errors.New("email already verified")
	}

	// Delete old verification codes
	_, err = s.db.ExecContext(ctx,
		`DELETE FROM email_verifications WHERE user_id = $1`,
		userID)

	if err != nil {
		log.Printf("Failed to delete old verification codes: %v", err)
	}

	// Generate new verification code
	code := s.generateVerificationCode()
	expiresAt := time.Now().Add(time.Duration(s.config.EmailCodeTTL) * time.Minute)

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO email_verifications (user_id, code, expires_at) 
         VALUES ($1, $2, $3)`,
		userID, code, expiresAt)

	if err != nil {
		return fmt.Errorf("failed to save verification code: %w", err)
	}

	// Store in Redis
	redisKey := fmt.Sprintf("email_verification:%s", email)
	err = s.redis.Set(ctx, redisKey, code, time.Duration(s.config.EmailCodeTTL)*time.Minute).Err()
	if err != nil {
		log.Printf("Failed to store verification code in Redis: %v", err)
	}

	// Send email
	if err := s.emailService.SendVerificationCode(email, code); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, verified, created_at, updated_at 
         FROM users WHERE id = $1`,
		userID).Scan(&user.ID, &user.Email, &user.Verified, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *AuthService) generateVerificationCode() string {
	rand.Seed(time.Now().UnixNano())
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	return code
}

func (s *AuthService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	// Get user
	var passwordHash string
	err := s.db.QueryRowContext(ctx,
		`SELECT password_hash FROM users WHERE id = $1`,
		userID).Scan(&passwordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(currentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	_, err = s.db.ExecContext(ctx,
		`UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		string(newHashedPassword), userID)

	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate all sessions
	sessionKey := fmt.Sprintf("session:%s", userID)
	s.redis.Del(ctx, sessionKey)

	return nil
}

func (s *AuthService) logDefaultAccountCreation(userID, accountID string) {
	// Отправляем HTTP запрос в API Service для логирования
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	logData := map[string]interface{}{
		"action": "created",
		"data": map[string]interface{}{
			"id":         accountID,
			"name":       "Основной счёт",
			"balance":    0.00,
			"is_default": true,
			"source":     "registration",
		},
	}

	jsonData, err := json.Marshal(logData)
	if err != nil {
		log.Printf("Failed to marshal log data: %v", err)
		return
	}

	// URL API Service (из env или hardcoded для внутренней сети Docker)
	apiURL := "http://api-service:8082/api/v1/internal/logs"

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to create log request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Service", "auth-service") // Для проверки что запрос изнутри

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send log to api-service: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("API service returned non-OK status: %d", resp.StatusCode)
	}
}
