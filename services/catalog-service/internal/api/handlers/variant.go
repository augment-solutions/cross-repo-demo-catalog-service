package handlers

import (
	"net/http"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/ecomm-stack/catalog-service/internal/domain/services"
	"github.com/ecomm-stack/catalog-service/internal/otel"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// VariantHandler handles variant-related requests
type VariantHandler struct {
	productService *services.ProductService
	logger         *zap.Logger
}

// NewVariantHandler creates a new VariantHandler
func NewVariantHandler(productService *services.ProductService, logger *zap.Logger) *VariantHandler {
	return &VariantHandler{
		productService: productService,
		logger:         logger,
	}
}

// Create creates a new variant for a product
func (h *VariantHandler) Create(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "VariantHandler.Create")
	defer span.End()

	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	var variant models.Variant
	if err := c.ShouldBindJSON(&variant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	variant.ID = uuid.New()
	variant.ProductID = productID

	if err := h.productService.AddVariant(ctx, productID, &variant); err != nil {
		h.logger.Error("failed to create variant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	span.SetAttributes(attribute.String("variant_id", variant.ID.String()))
	c.JSON(http.StatusCreated, variant)
}

// Get retrieves a variant by ID
func (h *VariantHandler) Get(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "VariantHandler.Get")
	defer span.End()

	variantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid variant ID"})
		return
	}

	span.SetAttributes(attribute.String("variant_id", variantID.String()))
	variant, err := h.productService.GetVariant(ctx, variantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "variant not found"})
		return
	}

	c.JSON(http.StatusOK, variant)
}

// List lists variants for a product
func (h *VariantHandler) List(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "VariantHandler.List")
	defer span.End()

	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	span.SetAttributes(attribute.String("product_id", productID.String()))
	variants, err := h.productService.GetVariants(ctx, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": variants})
}

// Update updates a variant
func (h *VariantHandler) Update(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "VariantHandler.Update")
	defer span.End()

	variantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid variant ID"})
		return
	}

	var variant models.Variant
	if err := c.ShouldBindJSON(&variant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	variant.ID = variantID
	if err := h.productService.UpdateVariant(ctx, &variant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, variant)
}

// Delete deletes a variant
func (h *VariantHandler) Delete(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "VariantHandler.Delete")
	defer span.End()

	variantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid variant ID"})
		return
	}

	if err := h.productService.DeleteVariant(ctx, variantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

