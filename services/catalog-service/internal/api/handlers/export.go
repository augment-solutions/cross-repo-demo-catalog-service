package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ecomm-stack/catalog-service/internal/domain/services"
	"github.com/ecomm-stack/catalog-service/internal/otel"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// ExportHandler handles product export requests
type ExportHandler struct {
	productService *services.ProductService
	logger         *zap.Logger
}

// NewExportHandler creates a new ExportHandler
func NewExportHandler(productService *services.ProductService, logger *zap.Logger) *ExportHandler {
	return &ExportHandler{
		productService: productService,
		logger:         logger,
	}
}

// ExportRequest represents an export request
type ExportRequest struct {
	Format     string   `form:"format" binding:"required,oneof=csv json xml"`
	CategoryID string   `form:"category_id"`
	BrandID    string   `form:"brand_id"`
	Status     string   `form:"status"`
	Fields     []string `form:"fields"`
}

// Export exports products
func (h *ExportHandler) Export(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "ExportHandler.Export")
	defer span.End()

	var req ExportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	span.SetAttributes(attribute.String("format", req.Format))

	// Get products based on filters
	products, _, err := h.productService.ListWithFilters(ctx, services.ProductFilters{
		CategoryID: req.CategoryID,
		BrandID:    req.BrandID,
		Status:     req.Status,
	}, 1, 10000)
	if err != nil {
		h.logger.Error("failed to get products for export", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("products_export_%s.%s", time.Now().Format("20060102_150405"), req.Format)
	c.Header("Content-Disposition", "attachment; filename="+filename)

	switch req.Format {
	case "csv":
		c.Header("Content-Type", "text/csv")
		h.exportCSV(c, products)
	case "json":
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, products)
	case "xml":
		c.Header("Content-Type", "application/xml")
		c.XML(http.StatusOK, products)
	}
}

// ExportAsync starts an async export job
func (h *ExportHandler) ExportAsync(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "ExportHandler.ExportAsync")
	defer span.End()

	var req ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	span.SetAttributes(attribute.String("format", req.Format))

	// Start async export job
	jobID, err := h.productService.StartExportJob(ctx, services.ExportOptions{
		Format:     req.Format,
		CategoryID: req.CategoryID,
		BrandID:    req.BrandID,
		Status:     req.Status,
		Fields:     req.Fields,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"job_id":  jobID,
		"message": "export job started",
	})
}

// GetExportStatus gets the status of an export job
func (h *ExportHandler) GetExportStatus(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "ExportHandler.GetExportStatus")
	defer span.End()

	jobID := c.Param("jobId")
	span.SetAttributes(attribute.String("job_id", jobID))

	status, err := h.productService.GetExportJobStatus(ctx, jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// DownloadExport downloads the export file
func (h *ExportHandler) DownloadExport(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "ExportHandler.DownloadExport")
	defer span.End()

	jobID := c.Param("jobId")
	span.SetAttributes(attribute.String("job_id", jobID))

	data, filename, err := h.productService.GetExportFile(ctx, jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "export not found"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/octet-stream", data)
}

func (h *ExportHandler) exportCSV(c *gin.Context, products interface{}) {
	// CSV export implementation
	c.String(http.StatusOK, "sku,name,description,price,status\n")
}

