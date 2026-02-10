package handlers

import (
	"net/http"
	"strconv"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/ecomm-stack/catalog-service/internal/otel"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AttributeHandler handles attribute-related requests
type AttributeHandler struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAttributeHandler creates a new AttributeHandler
func NewAttributeHandler(db *gorm.DB, logger *zap.Logger) *AttributeHandler {
	return &AttributeHandler{
		db:     db,
		logger: logger,
	}
}

// Create creates a new attribute
func (h *AttributeHandler) Create(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "AttributeHandler.Create")
	defer span.End()

	var attr models.Attribute
	if err := c.ShouldBindJSON(&attr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	attr.ID = uuid.New()
	if err := h.db.WithContext(ctx).Create(&attr).Error; err != nil {
		h.logger.Error("failed to create attribute", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	span.SetAttributes(attribute.String("attribute_id", attr.ID.String()))
	c.JSON(http.StatusCreated, attr)
}

// Get retrieves an attribute by ID
func (h *AttributeHandler) Get(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "AttributeHandler.Get")
	defer span.End()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attribute ID"})
		return
	}

	span.SetAttributes(attribute.String("attribute_id", id.String()))

	var attr models.Attribute
	if err := h.db.WithContext(ctx).Preload("Options").First(&attr, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "attribute not found"})
		return
	}

	c.JSON(http.StatusOK, attr)
}

// List lists attributes with pagination
func (h *AttributeHandler) List(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "AttributeHandler.List")
	defer span.End()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	var attrs []models.Attribute
	var total int64

	h.db.WithContext(ctx).Model(&models.Attribute{}).Count(&total)
	if err := h.db.WithContext(ctx).Preload("Options").Limit(pageSize).Offset(offset).Find(&attrs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      attrs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Update updates an attribute
func (h *AttributeHandler) Update(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "AttributeHandler.Update")
	defer span.End()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attribute ID"})
		return
	}

	var attr models.Attribute
	if err := c.ShouldBindJSON(&attr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	attr.ID = id
	if err := h.db.WithContext(ctx).Save(&attr).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, attr)
}

// Delete soft deletes an attribute
func (h *AttributeHandler) Delete(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "AttributeHandler.Delete")
	defer span.End()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attribute ID"})
		return
	}

	if err := h.db.WithContext(ctx).Delete(&models.Attribute{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// AddOption adds an option to an attribute
func (h *AttributeHandler) AddOption(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "AttributeHandler.AddOption")
	defer span.End()

	attrID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attribute ID"})
		return
	}

	var opt models.AttributeOption
	if err := c.ShouldBindJSON(&opt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	opt.ID = uuid.New()
	opt.AttributeID = attrID

	if err := h.db.WithContext(ctx).Create(&opt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, opt)
}

