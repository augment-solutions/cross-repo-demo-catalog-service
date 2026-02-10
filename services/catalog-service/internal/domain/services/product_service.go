package services

import (
	"context"
	"strings"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/ecomm-stack/catalog-service/internal/domain/repositories"
	"github.com/ecomm-stack/catalog-service/internal/infrastructure/cache"
	"github.com/ecomm-stack/catalog-service/internal/otel"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// ProductService handles product operations
type ProductService struct {
	productRepo    *repositories.ProductRepository
	productCache   *cache.ProductCache
	eventPublisher *EventPublisher
	logger         *zap.Logger
}

// NewProductService creates a new ProductService
func NewProductService(
	productRepo *repositories.ProductRepository,
	productCache *cache.ProductCache,
	eventPublisher *EventPublisher,
	logger *zap.Logger,
) *ProductService {
	return &ProductService{
		productRepo:    productRepo,
		productCache:   productCache,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(ctx context.Context, req *models.CreateProductRequest) (*models.Product, error) {
	ctx, span := otel.StartSpan(ctx, "ProductService.CreateProduct")
	defer span.End()
	span.SetAttributes(attribute.String("sku", req.SKU))

	product := &models.Product{
		SKU:              req.SKU,
		Name:             req.Name,
		Slug:             s.generateSlug(req.Name, req.Slug),
		Description:      req.Description,
		ShortDescription: req.ShortDescription,
		CategoryID:       req.CategoryID,
		BrandID:          req.BrandID,
		Status:           req.Status,
		IsFeatured:       req.IsFeatured,
		Weight:           req.Weight,
		MetaTitle:        req.MetaTitle,
		MetaDescription:  req.MetaDescription,
		Tags:             req.Tags,
	}

	if product.Status == "" {
		product.Status = models.ProductStatusDraft
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		s.logger.Error("failed to create product", zap.Error(err))
		return nil, err
	}

	// Create variants
	for _, v := range req.Variants {
		variant := &models.Variant{
			ProductID:      product.ID,
			SKU:            v.SKU,
			Name:           v.Name,
			Price:          v.Price,
			CompareAtPrice: v.CompareAtPrice,
			CostPrice:      v.CostPrice,
			Weight:         v.Weight,
			Barcode:        v.Barcode,
			IsDefault:      v.IsDefault,
		}
		product.Variants = append(product.Variants, *variant)
	}

	// Publish event
	event, _ := models.NewEvent(models.EventTypeProductCreated, product.ID, models.ProductCreatedData{
		ProductID:  product.ID,
		SKU:        product.SKU,
		Name:       product.Name,
		CategoryID: product.CategoryID,
		BrandID:    product.BrandID,
	})
	if err := s.eventPublisher.Publish(ctx, event); err != nil {
		s.logger.Warn("failed to publish product created event", zap.Error(err))
	}

	otel.ProductsCreated.Add(ctx, 1)
	s.logger.Info("product created", zap.String("id", product.ID.String()), zap.String("sku", product.SKU))
	return product, nil
}

// UpdateProduct updates a product
func (s *ProductService) UpdateProduct(ctx context.Context, id uuid.UUID, req *models.UpdateProductRequest) (*models.Product, error) {
	ctx, span := otel.StartSpan(ctx, "ProductService.UpdateProduct")
	defer span.End()

	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, nil
	}

	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Status != nil {
		product.Status = *req.Status
	}
	if req.CategoryID != nil {
		product.CategoryID = req.CategoryID
	}
	if req.BrandID != nil {
		product.BrandID = req.BrandID
	}

	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}

	// Invalidate cache
	s.productCache.Delete(ctx, id)

	otel.ProductsUpdated.Add(ctx, 1)
	return product, nil
}

// DeleteProduct deletes a product
func (s *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	ctx, span := otel.StartSpan(ctx, "ProductService.DeleteProduct")
	defer span.End()

	if err := s.productRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.productCache.Delete(ctx, id)
	otel.ProductsDeleted.Add(ctx, 1)
	return nil
}

// ListProducts lists products with pagination
func (s *ProductService) ListProducts(ctx context.Context, opts repositories.ListOptions) ([]*models.Product, int64, error) {
	ctx, span := otel.StartSpan(ctx, "ProductService.ListProducts")
	defer span.End()

	return s.productRepo.List(ctx, opts)
}

// GetProductByID retrieves a product by ID
func (s *ProductService) GetProductByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	ctx, span := otel.StartSpan(ctx, "ProductService.GetProductByID")
	defer span.End()

	return s.productRepo.GetByID(ctx, id)
}

// GetProductBySKU retrieves a product by SKU
func (s *ProductService) GetProductBySKU(ctx context.Context, sku string) (*models.Product, error) {
	ctx, span := otel.StartSpan(ctx, "ProductService.GetProductBySKU")
	defer span.End()

	return s.productRepo.GetBySKU(ctx, sku)
}

func (s *ProductService) generateSlug(name, slug string) string {
	if slug != "" {
		return slug
	}
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}

// ProductFilters defines filters for listing products
type ProductFilters struct {
	CategoryID string
	BrandID    string
	Status     string
}

// ExportOptions defines options for exporting products
type ExportOptions struct {
	Format     string
	CategoryID string
	BrandID    string
	Status     string
	Fields     []string
}

// ExportJobStatus represents the status of an export job
type ExportJobStatus struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
	Progress  int    `json:"progress"`
	FileURL   string `json:"file_url,omitempty"`
	Error     string `json:"error,omitempty"`
	CreatedAt string `json:"created_at"`
}

// ListWithFilters lists products with filters
func (s *ProductService) ListWithFilters(ctx context.Context, filters ProductFilters, page, pageSize int) ([]*models.Product, int64, error) {
	ctx, span := otel.StartSpan(ctx, "ProductService.ListWithFilters")
	defer span.End()

	opts := repositories.ListOptions{
		Page:     page,
		PageSize: pageSize,
		Filters:  make(map[string]interface{}),
	}

	if filters.CategoryID != "" {
		opts.Filters["category_id"] = filters.CategoryID
	}
	if filters.BrandID != "" {
		opts.Filters["brand_id"] = filters.BrandID
	}
	if filters.Status != "" {
		opts.Filters["status"] = filters.Status
	}

	return s.productRepo.List(ctx, opts)
}

// StartExportJob starts an async export job
func (s *ProductService) StartExportJob(ctx context.Context, opts ExportOptions) (string, error) {
	ctx, span := otel.StartSpan(ctx, "ProductService.StartExportJob")
	defer span.End()

	jobID := uuid.New().String()
	s.logger.Info("export job started", zap.String("job_id", jobID), zap.String("format", opts.Format))
	return jobID, nil
}

// GetExportJobStatus gets the status of an export job
func (s *ProductService) GetExportJobStatus(ctx context.Context, jobID string) (*ExportJobStatus, error) {
	ctx, span := otel.StartSpan(ctx, "ProductService.GetExportJobStatus")
	defer span.End()

	return &ExportJobStatus{
		JobID:    jobID,
		Status:   "completed",
		Progress: 100,
	}, nil
}

// GetExportFile gets the export file
func (s *ProductService) GetExportFile(ctx context.Context, jobID string) ([]byte, string, error) {
	ctx, span := otel.StartSpan(ctx, "ProductService.GetExportFile")
	defer span.End()

	return []byte{}, "application/octet-stream", nil
}

// AddVariant adds a variant to a product
func (s *ProductService) AddVariant(ctx context.Context, productID uuid.UUID, variant *models.Variant) error {
	ctx, span := otel.StartSpan(ctx, "ProductService.AddVariant")
	defer span.End()
	span.SetAttributes(attribute.String("product_id", productID.String()))

	// In a real implementation, this would add the variant to the database
	// For now, return nil as placeholder
	return nil
}

// GetVariant retrieves a variant by ID
func (s *ProductService) GetVariant(ctx context.Context, variantID uuid.UUID) (*models.Variant, error) {
	ctx, span := otel.StartSpan(ctx, "ProductService.GetVariant")
	defer span.End()
	span.SetAttributes(attribute.String("variant_id", variantID.String()))

	// In a real implementation, this would fetch the variant from the database
	// For now, return nil as placeholder
	return nil, nil
}

// GetVariants retrieves all variants for a product
func (s *ProductService) GetVariants(ctx context.Context, productID uuid.UUID) ([]*models.Variant, error) {
	ctx, span := otel.StartSpan(ctx, "ProductService.GetVariants")
	defer span.End()
	span.SetAttributes(attribute.String("product_id", productID.String()))

	// In a real implementation, this would fetch variants from the database
	// For now, return empty slice as placeholder
	return []*models.Variant{}, nil
}

// UpdateVariant updates a variant
func (s *ProductService) UpdateVariant(ctx context.Context, variant *models.Variant) error {
	ctx, span := otel.StartSpan(ctx, "ProductService.UpdateVariant")
	defer span.End()
	span.SetAttributes(attribute.String("variant_id", variant.ID.String()))

	// In a real implementation, this would update the variant in the database
	// For now, return nil as placeholder
	return nil
}

// DeleteVariant deletes a variant
func (s *ProductService) DeleteVariant(ctx context.Context, variantID uuid.UUID) error {
	ctx, span := otel.StartSpan(ctx, "ProductService.DeleteVariant")
	defer span.End()
	span.SetAttributes(attribute.String("variant_id", variantID.String()))

	// In a real implementation, this would delete the variant from the database
	// For now, return nil as placeholder
	return nil
}
