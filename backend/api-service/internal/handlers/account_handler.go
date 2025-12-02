package handlers

import (
	"net/http"

	"api-service/internal/models"
	"api-service/internal/services"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	accountService *services.AccountService
}

func NewAccountHandler(accountService *services.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

func (h *AccountHandler) CreateAccount(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	account, err := h.accountService.CreateAccount(c.Request.Context(), userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Account created successfully",
		"account": account,
	})
}

func (h *AccountHandler) GetAccounts(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	accounts, err := h.accountService.GetAccounts(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get stats for each account
	accountsWithStats := make([]map[string]interface{}, len(accounts))
	for i, account := range accounts {
		stats, _ := h.accountService.GetAccountStats(c.Request.Context(), userID.(string), account.ID)

		accountsWithStats[i] = map[string]interface{}{
			"id":         account.ID,
			"name":       account.Name,
			"balance":    account.Balance,
			"is_default": account.IsDefault,
			"created_at": account.CreatedAt,
			"updated_at": account.UpdatedAt,
		}

		if stats != nil {
			accountsWithStats[i]["stats"] = stats
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": accountsWithStats,
		"count":    len(accounts),
	})
}

func (h *AccountHandler) GetAccount(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	accountID := c.Param("id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account ID is required"})
		return
	}

	account, err := h.accountService.GetAccount(c.Request.Context(), userID.(string), accountID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "account not found" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	stats, _ := h.accountService.GetAccountStats(c.Request.Context(), userID.(string), accountID)

	response := gin.H{
		"account": account,
	}

	if stats != nil {
		response["stats"] = stats
	}

	c.JSON(http.StatusOK, response)
}

func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	accountID := c.Param("id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account ID is required"})
		return
	}

	var req models.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	account, err := h.accountService.UpdateAccount(c.Request.Context(), userID.(string), accountID, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "account not found" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account updated successfully",
		"account": account,
	})
}

func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	accountID := c.Param("id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account ID is required"})
		return
	}

	err := h.accountService.DeleteAccount(c.Request.Context(), userID.(string), accountID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "account not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "cannot delete the only account" {
			statusCode = http.StatusBadRequest
		} else if err.Error() == "cannot delete account with existing transactions" {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account deleted successfully",
	})
}

func (h *AccountHandler) SetDefaultAccount(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	accountID := c.Param("id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account ID is required"})
		return
	}

	err := h.accountService.SetDefaultAccount(c.Request.Context(), userID.(string), accountID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "account not found" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Default account set successfully",
	})
}
