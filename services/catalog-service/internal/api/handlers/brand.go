package handlers

import (
	"net/http"
	"strconv"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/ecomm-stack/catalog-service/internal/domain/repositories"
	"github.com/ecomm-stack/catalog-service/internal/otel"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// BrandHandler handles brand-related requests
type BrandHandler struct {
	brandRepo repositories.BrandRepository
	logger    *zap.Logger
}

// NewBrandHandler creates a new BrandHandler
func NewBrandHandler(brandRepo repositories.BrandRepository, logger *zap.Logger) *BrandHandler {
	return &BrandHandler{
		brandRepo: brandRepo,
		logger:    logger,
	}
}

// Create creates a new brand
func (h *BrandHandler) Create(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "BrandHandler.Create")
	defer span.End()

	var brand models.Brand
	if err := c.ShouldBindJSON(&brand); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	brand.ID = uuid.New()
	if err := h.brandRepo.Create(ctx, &brand); err != nil {
		h.logger.Error("failed to create brand", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	span.SetAttributes(attribute.String("brand_id", brand.ID.String()))
	c.JSON(http.StatusCreated, brand)
}

// Get retrieves a brand by ID
func (h *BrandHandler) Get(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "BrandHandler.Get")
	defer span.End()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid brand ID"})
		return
	}

	span.SetAttributes(attribute.String("brand_id", id.String()))
	brand, err := h.brandRepo.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "brand not found"})
		return
	}

	c.JSON(http.StatusOK, brand)
}

// List lists brands with pagination
func (h *BrandHandler) List(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "BrandHandler.List")
	defer span.End()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	opts := repositories.ListOptions{
		Page:     page,
		PageSize: pageSize,
	}

	brands, total, err := h.brandRepo.List(ctx, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      brands,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Update updates a brand
func (h *BrandHandler) Update(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "BrandHandler.Update")
	defer span.End()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid brand ID"})
		return
	}

	var brand models.Brand
	if err := c.ShouldBindJSON(&brand); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	brand.ID = id
	if err := h.brandRepo.Update(ctx, &brand); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, brand)
}

// Delete soft deletes a brand
func (h *BrandHandler) Delete(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "BrandHandler.Delete")
	defer span.End()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid brand ID"})
		return
	}

	if err := h.brandRepo.Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

