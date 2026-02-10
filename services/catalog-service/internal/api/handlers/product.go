package handlers

import (
	"net/http"
	"strconv"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/ecomm-stack/catalog-service/internal/domain/services"
	"github.com/ecomm-stack/catalog-service/internal/otel"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// ProductHandler handles product-related requests
type ProductHandler struct {
	productService *services.ProductService
	logger         *zap.Logger
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler(productService *services.ProductService, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		logger:         logger,
	}
}

// Create creates a new product
func (h *ProductHandler) Create(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "ProductHandler.Create")
	defer span.End()

	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.productService.CreateProduct(ctx, &req)
	if err != nil {
		h.logger.Error("failed to create product", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	otel.ProductsCreated.Add(ctx, 1)
	span.SetAttributes(attribute.String("product_id", product.ID.String()))
	c.JSON(http.StatusCreated, product)
}

// Get retrieves a product by ID
func (h *ProductHandler) Get(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "ProductHandler.Get")
	defer span.End()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	span.SetAttributes(attribute.String("product_id", id.String()))
	product, err := h.productService.GetProductByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// List lists products with pagination
func (h *ProductHandler) List(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "ProductHandler.List")
	defer span.End()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	// Also support "limit" as an alias for page_size
	if limit := c.Query("limit"); limit != "" {
		pageSize, _ = strconv.Atoi(limit)
	}

	opts := services.ProductFilters{
		CategoryID: c.Query("category_id"),
		BrandID:    c.Query("brand_id"),
		Status:     c.Query("status"),
	}
	// Also support "category" as an alias for category_id (frontend uses this)
	if category := c.Query("category"); category != "" && opts.CategoryID == "" {
		opts.CategoryID = category
	}
	products, total, err := h.productService.ListWithFilters(ctx, opts, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"items":      products,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"totalPages": totalPages,
	})
}

// Update updates a product
func (h *ProductHandler) Update(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "ProductHandler.Update")
	defer span.End()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.productService.UpdateProduct(ctx, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// Delete soft deletes a product
func (h *ProductHandler) Delete(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "ProductHandler.Delete")
	defer span.End()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	if err := h.productService.DeleteProduct(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetBySKU retrieves a product by SKU
func (h *ProductHandler) GetBySKU(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "ProductHandler.GetBySKU")
	defer span.End()

	sku := c.Param("sku")
	span.SetAttributes(attribute.String("sku", sku))

	product, err := h.productService.GetProductBySKU(ctx, sku)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}
