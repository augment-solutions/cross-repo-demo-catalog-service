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

// MediaHandler handles media-related requests
type MediaHandler struct {
	mediaRepo repositories.MediaRepository
	logger    *zap.Logger
}

// NewMediaHandler creates a new MediaHandler
func NewMediaHandler(mediaRepo repositories.MediaRepository, logger *zap.Logger) *MediaHandler {
	return &MediaHandler{
		mediaRepo: mediaRepo,
		logger:    logger,
	}
}

// Create creates a new media entry
func (h *MediaHandler) Create(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "MediaHandler.Create")
	defer span.End()

	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	var media models.Media
	if err := c.ShouldBindJSON(&media); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	media.ID = uuid.New()
	media.ProductID = productID

	if err := h.mediaRepo.Create(ctx, &media); err != nil {
		h.logger.Error("failed to create media", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	span.SetAttributes(attribute.String("media_id", media.ID.String()))
	c.JSON(http.StatusCreated, media)
}

// Get retrieves a media by ID
func (h *MediaHandler) Get(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "MediaHandler.Get")
	defer span.End()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	span.SetAttributes(attribute.String("media_id", id.String()))
	media, err := h.mediaRepo.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "media not found"})
		return
	}

	c.JSON(http.StatusOK, media)
}

// List lists media for a product
func (h *MediaHandler) List(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "MediaHandler.List")
	defer span.End()

	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	span.SetAttributes(attribute.String("product_id", productID.String()))
	mediaList, err := h.mediaRepo.ListByProduct(ctx, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mediaList})
}

// Update updates a media entry
func (h *MediaHandler) Update(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "MediaHandler.Update")
	defer span.End()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	var media models.Media
	if err := c.ShouldBindJSON(&media); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	media.ID = id
	if err := h.mediaRepo.Update(ctx, &media); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, media)
}

// Delete deletes a media entry
func (h *MediaHandler) Delete(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "MediaHandler.Delete")
	defer span.End()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	if err := h.mediaRepo.Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// SetPrimary sets a media as primary
func (h *MediaHandler) SetPrimary(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "MediaHandler.SetPrimary")
	defer span.End()

	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	mediaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	if err := h.mediaRepo.SetPrimary(ctx, productID, mediaID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "primary media set"})
}

// Reorder reorders media for a product
func (h *MediaHandler) Reorder(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "MediaHandler.Reorder")
	defer span.End()

	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	var req struct {
		MediaIDs []uuid.UUID `json:"media_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.mediaRepo.UpdateSortOrder(ctx, productID, req.MediaIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "media reordered"})
}

// unused but required to satisfy interface
var _ = strconv.Atoi

