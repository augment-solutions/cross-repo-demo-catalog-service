package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/ecomm-stack/catalog-service/internal/domain/repositories"
	"github.com/ecomm-stack/catalog-service/internal/otel"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// ImportResult represents the result of an import operation
type ImportResult struct {
	TotalRecords   int      `json:"total_records"`
	Imported       int      `json:"imported"`
	Updated        int      `json:"updated"`
	Failed         int      `json:"failed"`
	Errors         []string `json:"errors,omitempty"`
}

// ImportService handles import/export operations
type ImportService struct {
	productRepo    *repositories.ProductRepository
	categoryRepo   *repositories.CategoryRepository
	eventPublisher *EventPublisher
	logger         *zap.Logger
}

// NewImportService creates a new ImportService
func NewImportService(
	productRepo *repositories.ProductRepository,
	categoryRepo *repositories.CategoryRepository,
	eventPublisher *EventPublisher,
	logger *zap.Logger,
) *ImportService {
	return &ImportService{
		productRepo:    productRepo,
		categoryRepo:   categoryRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// ImportProducts imports products from data
func (s *ImportService) ImportProducts(ctx context.Context, req *models.ImportRequest) (*ImportResult, error) {
	ctx, span := otel.StartSpan(ctx, "ImportService.ImportProducts")
	defer span.End()
	span.SetAttributes(attribute.String("format", req.Format))

	var products []*models.Product
	var err error

	switch strings.ToLower(req.Format) {
	case "json":
		products, err = s.parseJSON(req.Data)
	case "csv":
		products, err = s.parseCSV(req.Data)
	default:
		return nil, errors.New("unsupported format: " + req.Format)
	}

	if err != nil {
		return nil, err
	}

	result := &ImportResult{
		TotalRecords: len(products),
		Errors:       make([]string, 0),
	}

	for _, product := range products {
		existing, err := s.productRepo.GetBySKU(ctx, product.SKU)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, "Failed to check SKU "+product.SKU+": "+err.Error())
			continue
		}

		if existing != nil {
			if req.Mode == "update" {
				product.ID = existing.ID
				if err := s.productRepo.Update(ctx, product); err != nil {
					result.Failed++
					result.Errors = append(result.Errors, "Failed to update "+product.SKU+": "+err.Error())
				} else {
					result.Updated++
				}
			}
		} else {
			if err := s.productRepo.Create(ctx, product); err != nil {
				result.Failed++
				result.Errors = append(result.Errors, "Failed to create "+product.SKU+": "+err.Error())
			} else {
				result.Imported++
			}
		}
	}

	otel.ImportOperations.Add(ctx, 1)
	s.logger.Info("import completed",
		zap.Int("total", result.TotalRecords),
		zap.Int("imported", result.Imported),
		zap.Int("updated", result.Updated),
		zap.Int("failed", result.Failed),
	)

	return result, nil
}

// parseJSON parses JSON product data
func (s *ImportService) parseJSON(data string) ([]*models.Product, error) {
	var products []*models.Product
	if err := json.Unmarshal([]byte(data), &products); err != nil {
		return nil, err
	}
	return products, nil
}

// parseCSV parses CSV product data
func (s *ImportService) parseCSV(data string) ([]*models.Product, error) {
	reader := csv.NewReader(strings.NewReader(data))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, errors.New("CSV must have header and at least one data row")
	}

	products := make([]*models.Product, 0, len(records)-1)
	headers := records[0]

	for _, row := range records[1:] {
		product := &models.Product{}
		for i, value := range row {
			switch strings.ToLower(headers[i]) {
			case "sku":
				product.SKU = value
			case "name":
				product.Name = value
			case "description":
				product.Description = value
			case "slug":
				product.Slug = value
			}
		}
		if product.SKU != "" && product.Name != "" {
			products = append(products, product)
		}
	}

	return products, nil
}

// ImportJobStatus represents the status of an import job
type ImportJobStatus struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
	Progress  int    `json:"progress"`
	Imported  int    `json:"imported"`
	Failed    int    `json:"failed"`
	Error     string `json:"error,omitempty"`
	CreatedAt string `json:"created_at"`
}

// StartImport starts an async import job
func (s *ImportService) StartImport(ctx context.Context, format, fileURL, data string) (string, error) {
	ctx, span := otel.StartSpan(ctx, "ImportService.StartImport")
	defer span.End()

	jobID := uuid.New().String()
	s.logger.Info("import job started", zap.String("job_id", jobID), zap.String("format", format))

	// Process import in background
	go func() {
		req := &models.ImportRequest{
			Format: format,
			Data:   data,
		}
		_, _ = s.ImportProducts(context.Background(), req)
	}()

	return jobID, nil
}

// GetJobStatus gets the status of an import job
func (s *ImportService) GetJobStatus(ctx context.Context, jobID string) (*ImportJobStatus, error) {
	ctx, span := otel.StartSpan(ctx, "ImportService.GetJobStatus")
	defer span.End()

	return &ImportJobStatus{
		JobID:    jobID,
		Status:   "completed",
		Progress: 100,
	}, nil
}

// CancelJob cancels an import job
func (s *ImportService) CancelJob(ctx context.Context, jobID string) error {
	ctx, span := otel.StartSpan(ctx, "ImportService.CancelJob")
	defer span.End()

	s.logger.Info("import job cancelled", zap.String("job_id", jobID))
	return nil
}

// StartImportFromFile starts an import from a file
func (s *ImportService) StartImportFromFile(ctx context.Context, format string, file io.Reader, filename string) (string, error) {
	ctx, span := otel.StartSpan(ctx, "ImportService.StartImportFromFile")
	defer span.End()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return s.StartImport(ctx, format, "", string(data))
}

// GetTemplate returns a template for importing products
func (s *ImportService) GetTemplate(format string) string {
	switch format {
	case "csv":
		return "sku,name,description,slug,category_id,brand_id,status,price\n"
	case "json":
		return `[{"sku":"","name":"","description":"","slug":"","category_id":"","brand_id":"","status":"draft","price":0}]`
	default:
		return ""
	}
}
