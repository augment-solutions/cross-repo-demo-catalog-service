package services

import (
	"context"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/ecomm-stack/catalog-service/internal/domain/repositories"
	"github.com/ecomm-stack/catalog-service/internal/infrastructure/cache"
	"github.com/ecomm-stack/catalog-service/internal/otel"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// CatalogService handles catalog operations
type CatalogService struct {
	productRepo    *repositories.ProductRepository
	categoryRepo   *repositories.CategoryRepository
	brandRepo      *repositories.BrandRepository
	productCache   *cache.ProductCache
	eventPublisher *EventPublisher
	logger         *zap.Logger
}

// NewCatalogService creates a new CatalogService
func NewCatalogService(
	productRepo *repositories.ProductRepository,
	categoryRepo *repositories.CategoryRepository,
	brandRepo *repositories.BrandRepository,
	productCache *cache.ProductCache,
	eventPublisher *EventPublisher,
	logger *zap.Logger,
) *CatalogService {
	return &CatalogService{
		productRepo:    productRepo,
		categoryRepo:   categoryRepo,
		brandRepo:      brandRepo,
		productCache:   productCache,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// GetProduct retrieves a product with enriched data
func (s *CatalogService) GetProduct(ctx context.Context, id uuid.UUID) (*models.ProductResponse, error) {
	ctx, span := otel.StartSpan(ctx, "CatalogService.GetProduct")
	defer span.End()
	span.SetAttributes(attribute.String("product_id", id.String()))

	// Try cache first
	product, err := s.productCache.Get(ctx, id)
	if err == nil && product != nil {
		otel.CacheHits.Add(ctx, 1)
		return s.enrichProduct(ctx, product)
	}
	otel.CacheMisses.Add(ctx, 1)

	// Fetch from database
	product, err = s.productRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get product", zap.Error(err), zap.String("id", id.String()))
		return nil, err
	}
	if product == nil {
		return nil, nil
	}

	// Cache the result
	if err := s.productCache.Set(ctx, product); err != nil {
		s.logger.Warn("failed to cache product", zap.Error(err))
	}

	otel.ProductsViewed.Add(ctx, 1)
	return s.enrichProduct(ctx, product)
}

// enrichProduct adds category and brand info
func (s *CatalogService) enrichProduct(ctx context.Context, product *models.Product) (*models.ProductResponse, error) {
	response := &models.ProductResponse{Product: models.NewProductDetailsResponse(product)}

	if product.CategoryID != nil {
		category, err := s.categoryRepo.GetByID(ctx, *product.CategoryID)
		if err == nil && category != nil {
			response.Category = category
		}
	}

	if product.BrandID != nil {
		brand, err := s.brandRepo.GetByID(ctx, *product.BrandID)
		if err == nil && brand != nil {
			response.Brand = brand
		}
	}

	return response, nil
}

// GetCategoryTree retrieves the full category tree
func (s *CatalogService) GetCategoryTree(ctx context.Context) ([]*models.CategoryTree, error) {
	ctx, span := otel.StartSpan(ctx, "CatalogService.GetCategoryTree")
	defer span.End()

	return s.categoryRepo.GetTree(ctx)
}

// GetFeaturedProducts retrieves featured products
func (s *CatalogService) GetFeaturedProducts(ctx context.Context, limit int) ([]*models.Product, error) {
	ctx, span := otel.StartSpan(ctx, "CatalogService.GetFeaturedProducts")
	defer span.End()

	opts := repositories.DefaultListOptions()
	opts.PageSize = limit
	opts.Filters["is_featured"] = true
	opts.Filters["status"] = models.ProductStatusActive

	products, _, err := s.productRepo.List(ctx, opts)
	return products, err
}

// ExportProducts exports products based on criteria
func (s *CatalogService) ExportProducts(ctx context.Context, req *models.ExportRequest) ([]*models.Product, error) {
	ctx, span := otel.StartSpan(ctx, "CatalogService.ExportProducts")
	defer span.End()

	opts := repositories.ListOptions{
		Page:     1,
		PageSize: 10000, // Max export limit
		Filters:  make(map[string]interface{}),
	}

	if req.CategoryID != nil {
		opts.Filters["category_id"] = *req.CategoryID
	}
	if req.BrandID != nil {
		opts.Filters["brand_id"] = *req.BrandID
	}
	if req.Status != "" {
		opts.Filters["status"] = req.Status
	}

	products, _, err := s.productRepo.List(ctx, opts)
	if err != nil {
		return nil, err
	}

	otel.ExportOperations.Add(ctx, 1)
	return products, nil
}

// RefreshCache refreshes the product cache
func (s *CatalogService) RefreshCache(ctx context.Context, productID uuid.UUID) error {
	ctx, span := otel.StartSpan(ctx, "CatalogService.RefreshCache")
	defer span.End()

	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}
	if product != nil {
		return s.productCache.Set(ctx, product)
	}
	return s.productCache.Delete(ctx, productID)
}
