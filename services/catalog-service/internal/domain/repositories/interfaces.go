package repositories

import (
	"context"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/google/uuid"
)

// ProductRepository defines the interface for product persistence
type ProductRepositoryInterface interface {
	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	GetBySKU(ctx context.Context, sku string) (*models.Product, error)
	GetBySlug(ctx context.Context, slug string) (*models.Product, error)
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, opts ListOptions) ([]*models.Product, int64, error)
	ListByCategory(ctx context.Context, categoryID uuid.UUID, opts ListOptions) ([]*models.Product, int64, error)
	ListByBrand(ctx context.Context, brandID uuid.UUID, opts ListOptions) ([]*models.Product, int64, error)
	Search(ctx context.Context, query string, opts ListOptions) ([]*models.Product, int64, error)
	BulkCreate(ctx context.Context, products []*models.Product) error
	BulkUpdate(ctx context.Context, products []*models.Product) error
}

// CategoryRepository defines the interface for category persistence
type CategoryRepositoryInterface interface {
	Create(ctx context.Context, category *models.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error)
	GetBySlug(ctx context.Context, slug string) (*models.Category, error)
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, opts ListOptions) ([]*models.Category, int64, error)
	ListChildren(ctx context.Context, parentID uuid.UUID) ([]*models.Category, error)
	GetTree(ctx context.Context) ([]*models.CategoryTree, error)
	UpdateProductCount(ctx context.Context, categoryID uuid.UUID) error
}

// BrandRepository defines the interface for brand persistence
type BrandRepositoryInterface interface {
	Create(ctx context.Context, brand *models.Brand) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Brand, error)
	GetBySlug(ctx context.Context, slug string) (*models.Brand, error)
	Update(ctx context.Context, brand *models.Brand) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, opts ListOptions) ([]*models.Brand, int64, error)
	UpdateProductCount(ctx context.Context, brandID uuid.UUID) error
}

// MediaRepository defines the interface for media persistence
type MediaRepositoryInterface interface {
	Create(ctx context.Context, media *models.Media) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Media, error)
	Update(ctx context.Context, media *models.Media) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByProduct(ctx context.Context, productID uuid.UUID) ([]*models.Media, error)
	SetPrimary(ctx context.Context, productID, mediaID uuid.UUID) error
	UpdateSortOrder(ctx context.Context, productID uuid.UUID, mediaIDs []uuid.UUID) error
}

// ListOptions contains pagination and filtering options
type ListOptions struct {
	Page       int
	PageSize   int
	SortBy     string
	SortOrder  string
	Filters    map[string]interface{}
	IncludeDeleted bool
}

// DefaultListOptions returns default list options
func DefaultListOptions() ListOptions {
	return ListOptions{
		Page:      1,
		PageSize:  20,
		SortBy:    "created_at",
		SortOrder: "desc",
		Filters:   make(map[string]interface{}),
	}
}

// Offset calculates the offset for pagination
func (o ListOptions) Offset() int {
	if o.Page < 1 {
		o.Page = 1
	}
	return (o.Page - 1) * o.PageSize
}

// Limit returns the limit for pagination
func (o ListOptions) Limit() int {
	if o.PageSize <= 0 {
		return 20
	}
	if o.PageSize > 100 {
		return 100
	}
	return o.PageSize
}

