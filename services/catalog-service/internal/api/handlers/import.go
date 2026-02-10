package handlers

import (
	"net/http"

	"github.com/ecomm-stack/catalog-service/internal/domain/services"
	"github.com/ecomm-stack/catalog-service/internal/otel"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// ImportHandler handles product import requests
type ImportHandler struct {
	importService *services.ImportService
	logger        *zap.Logger
}

// NewImportHandler creates a new ImportHandler
func NewImportHandler(importService *services.ImportService, logger *zap.Logger) *ImportHandler {
	return &ImportHandler{
		importService: importService,
		logger:        logger,
	}
}

// ImportRequest represents an import request
type ImportRequest struct {
	Format  string `json:"format" binding:"required,oneof=csv json xml"`
	FileURL string `json:"file_url,omitempty"`
	Data    string `json:"data,omitempty"`
}

// Import imports products from a file or data
func (h *ImportHandler) Import(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "ImportHandler.Import")
	defer span.End()

	var req ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	span.SetAttributes(attribute.String("format", req.Format))

	jobID, err := h.importService.StartImport(ctx, req.Format, req.FileURL, req.Data)
	if err != nil {
		h.logger.Error("failed to start import", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"job_id":  jobID,
		"message": "import job started",
	})
}

// GetImportStatus gets the status of an import job
func (h *ImportHandler) GetImportStatus(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "ImportHandler.GetImportStatus")
	defer span.End()

	jobID := c.Param("jobId")
	span.SetAttributes(attribute.String("job_id", jobID))

	status, err := h.importService.GetJobStatus(ctx, jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// CancelImport cancels an import job
func (h *ImportHandler) CancelImport(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "ImportHandler.CancelImport")
	defer span.End()

	jobID := c.Param("jobId")
	span.SetAttributes(attribute.String("job_id", jobID))

	if err := h.importService.CancelJob(ctx, jobID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "import job cancelled"})
}

// UploadImport handles file upload for import
func (h *ImportHandler) UploadImport(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "ImportHandler.UploadImport")
	defer span.End()

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	defer file.Close()

	format := c.PostForm("format")
	if format == "" {
		format = "csv"
	}

	span.SetAttributes(
		attribute.String("format", format),
		attribute.String("filename", header.Filename),
	)

	jobID, err := h.importService.StartImportFromFile(ctx, format, file, header.Filename)
	if err != nil {
		h.logger.Error("failed to start import", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"job_id":  jobID,
		"message": "import job started",
	})
}

// GetImportTemplate returns a template for importing products
func (h *ImportHandler) GetImportTemplate(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")

	template := h.importService.GetTemplate(format)
	c.Header("Content-Disposition", "attachment; filename=product_import_template."+format)
	c.Data(http.StatusOK, "application/octet-stream", []byte(template))
}

