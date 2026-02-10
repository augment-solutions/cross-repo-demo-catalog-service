package handlers

import (
	"net/http"
	"strconv"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/ecomm-stack/catalog-service/internal/domain/repositories"
	"github.com/ecomm-stack/catalog-service/internal/domain/services"
	"github.com/ecomm-stack/catalog-service/internal/otel"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// CategoryHandler handles category-related requests
type CategoryHandler struct {
	categoryService *services.CategoryService
	logger          *zap.Logger
}

// NewCategoryHandler creates a new CategoryHandler
func NewCategoryHandler(categoryService *services.CategoryService, logger *zap.Logger) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
		logger:          logger,
	}
}

// Create creates a new category
func (h *CategoryHandler) Create(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "CategoryHandler.Create")
	defer span.End()

	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.categoryService.CreateCategory(ctx, &req)
	if err != nil {
		h.logger.Error("failed to create category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	span.SetAttributes(attribute.String("category_id", category.ID.String()))
	c.JSON(http.StatusCreated, category)
}

// Get retrieves a category by ID
func (h *CategoryHandler) Get(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "CategoryHandler.Get")
	defer span.End()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	span.SetAttributes(attribute.String("category_id", id.String()))
	category, err := h.categoryService.GetCategory(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// List lists categories with pagination
func (h *CategoryHandler) List(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "CategoryHandler.List")
	defer span.End()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	opts := repositories.ListOptions{
		Page:     page,
		PageSize: pageSize,
	}
	categories, total, err := h.categoryService.ListCategories(ctx, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      categories,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetTree retrieves the category tree
func (h *CategoryHandler) GetTree(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "CategoryHandler.GetTree")
	defer span.End()

	tree, err := h.categoryService.GetTree(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tree)
}

// Update updates a category
func (h *CategoryHandler) Update(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "CategoryHandler.Update")
	defer span.End()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	var req models.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.categoryService.UpdateCategory(ctx, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// Delete soft deletes a category
func (h *CategoryHandler) Delete(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "CategoryHandler.Delete")
	defer span.End()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	if err := h.categoryService.DeleteCategory(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

