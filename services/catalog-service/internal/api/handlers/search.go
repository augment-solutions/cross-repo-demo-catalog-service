package handlers

import (
	"net/http"
	"strconv"

	"github.com/ecomm-stack/catalog-service/internal/domain/services"
	"github.com/ecomm-stack/catalog-service/internal/otel"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// SearchHandler handles search-related requests
type SearchHandler struct {
	searchService *services.SearchService
	logger        *zap.Logger
}

// NewSearchHandler creates a new SearchHandler
func NewSearchHandler(searchService *services.SearchService, logger *zap.Logger) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
		logger:        logger,
	}
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query      string   `form:"q"`
	CategoryID string   `form:"category_id"`
	BrandID    string   `form:"brand_id"`
	MinPrice   float64  `form:"min_price"`
	MaxPrice   float64  `form:"max_price"`
	Tags       []string `form:"tags"`
	SortBy     string   `form:"sort_by"`
	SortOrder  string   `form:"sort_order"`
	Page       int      `form:"page"`
	PageSize   int      `form:"page_size"`
}

// Search performs a product search
func (h *SearchHandler) Search(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "SearchHandler.Search")
	defer span.End()

	var req SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	span.SetAttributes(
		attribute.String("query", req.Query),
		attribute.Int("page", req.Page),
		attribute.Int("page_size", req.PageSize),
	)

	opts := &services.SearchOptions{
		Query:      req.Query,
		CategoryID: req.CategoryID,
		BrandID:    req.BrandID,
		MinPrice:   req.MinPrice,
		MaxPrice:   req.MaxPrice,
		Tags:       req.Tags,
		SortBy:     req.SortBy,
		SortOrder:  req.SortOrder,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	result, err := h.searchService.Search(ctx, opts)
	if err != nil {
		h.logger.Error("search failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Suggest provides search suggestions
func (h *SearchHandler) Suggest(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "SearchHandler.Suggest")
	defer span.End()

	query := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	span.SetAttributes(attribute.String("query", query))

	suggestions, err := h.searchService.Suggest(ctx, query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"suggestions": suggestions})
}

// Facets returns available facets for filtering
func (h *SearchHandler) Facets(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "SearchHandler.Facets")
	defer span.End()

	categoryID := c.Query("category_id")
	facets, err := h.searchService.GetFacets(ctx, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, facets)
}

// Reindex triggers a full reindex
func (h *SearchHandler) Reindex(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "SearchHandler.Reindex")
	defer span.End()

	go func() {
		if err := h.searchService.ReindexAll(ctx); err != nil {
			h.logger.Error("reindex failed", zap.Error(err))
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{"message": "reindex started"})
}

