package handlers

import (
	"net/http"
	"time"

	"api-service/internal/models"
	"api-service/internal/services"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService *services.CategoryService
}

func NewCategoryHandler(categoryService *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	category, err := h.categoryService.CreateCategory(c.Request.Context(), userID.(string), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "category with this name already exists" {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Category created successfully",
		"category": category,
	})
}

func (h *CategoryHandler) GetCategories(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	categoryType := c.Query("type")

	var categories []*models.Category
	var err error

	if categoryType != "" {
		if categoryType != "income" && categoryType != "expense" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category type"})
			return
		}
		categories, err = h.categoryService.GetCategoriesByType(c.Request.Context(), userID.(string), categoryType)
	} else {
		categories, err = h.categoryService.GetCategories(c.Request.Context(), userID.(string))
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Group categories by type
	grouped := map[string][]*models.Category{
		"income":  []*models.Category{},
		"expense": []*models.Category{},
	}

	for _, category := range categories {
		grouped[category.Type] = append(grouped[category.Type], category)
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
		"grouped":    grouped,
		"count":      len(categories),
	})
}

func (h *CategoryHandler) GetCategory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	categoryID := c.Param("id")
	if categoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID is required"})
		return
	}

	category, err := h.categoryService.GetCategory(c.Request.Context(), userID.(string), categoryID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "category not found" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// Get category stats for the last month
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	stats, _ := h.categoryService.GetCategoryStats(c.Request.Context(), userID.(string), startDate, endDate)

	var categoryStats *models.CategoryStats
	for _, stat := range stats {
		if stat.CategoryID == categoryID {
			categoryStats = stat
			break
		}
	}

	response := gin.H{
		"category": category,
	}

	if categoryStats != nil {
		response["stats"] = categoryStats
	}

	c.JSON(http.StatusOK, response)
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	categoryID := c.Param("id")
	if categoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID is required"})
		return
	}

	var req models.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	category, err := h.categoryService.UpdateCategory(c.Request.Context(), userID.(string), categoryID, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "category not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "cannot modify system category" {
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Category updated successfully",
		"category": category,
	})
}

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	categoryID := c.Param("id")
	if categoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID is required"})
		return
	}

	err := h.categoryService.DeleteCategory(c.Request.Context(), userID.(string), categoryID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "category not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "cannot delete system category" {
			statusCode = http.StatusForbidden
		} else if err.Error() == "cannot delete category with existing transactions" {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category deleted successfully",
	})
}
