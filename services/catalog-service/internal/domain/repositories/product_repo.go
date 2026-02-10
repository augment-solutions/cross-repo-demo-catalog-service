package repositories

import (
	"context"
	"errors"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductRepository implements ProductRepositoryInterface
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new ProductRepository
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create creates a new product
func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

// GetByID retrieves a product by ID
func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).
		Preload("Variants").
		Preload("Media").
		Preload("Attributes").
		First(&product, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &product, err
}

// GetBySKU retrieves a product by SKU
func (r *ProductRepository) GetBySKU(ctx context.Context, sku string) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).
		Preload("Variants").
		First(&product, "sku = ?", sku).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &product, err
}

// GetBySlug retrieves a product by slug
func (r *ProductRepository) GetBySlug(ctx context.Context, slug string) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).
		Preload("Variants").
		Preload("Media").
		First(&product, "slug = ?", slug).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &product, err
}

// Update updates a product
func (r *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

// Delete soft-deletes a product
func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Product{}, "id = ?", id).Error
}

// List retrieves products with pagination
func (r *ProductRepository) List(ctx context.Context, opts ListOptions) ([]*models.Product, int64, error) {
	var products []*models.Product
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Product{})

	for key, value := range opts.Filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply default sort if not specified
	sortBy := opts.SortBy
	sortOrder := opts.SortOrder
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	err := query.
		Preload("Variants").
		Preload("Media").
		Offset(opts.Offset()).
		Limit(opts.Limit()).
		Order(sortBy + " " + sortOrder).
		Find(&products).Error

	return products, total, err
}

// ListByCategory retrieves products by category
func (r *ProductRepository) ListByCategory(ctx context.Context, categoryID uuid.UUID, opts ListOptions) ([]*models.Product, int64, error) {
	opts.Filters["category_id"] = categoryID
	return r.List(ctx, opts)
}

// ListByBrand retrieves products by brand
func (r *ProductRepository) ListByBrand(ctx context.Context, brandID uuid.UUID, opts ListOptions) ([]*models.Product, int64, error) {
	opts.Filters["brand_id"] = brandID
	return r.List(ctx, opts)
}

// Search searches products by query
func (r *ProductRepository) Search(ctx context.Context, query string, opts ListOptions) ([]*models.Product, int64, error) {
	var products []*models.Product
	var total int64

	searchQuery := "%" + query + "%"
	dbQuery := r.db.WithContext(ctx).Model(&models.Product{}).
		Where("name ILIKE ? OR description ILIKE ? OR sku ILIKE ?", searchQuery, searchQuery, searchQuery)

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := dbQuery.
		Preload("Variants").
		Offset(opts.Offset()).
		Limit(opts.Limit()).
		Find(&products).Error

	return products, total, err
}

// BulkCreate creates multiple products
func (r *ProductRepository) BulkCreate(ctx context.Context, products []*models.Product) error {
	return r.db.WithContext(ctx).CreateInBatches(products, 100).Error
}

// BulkUpdate updates multiple products
func (r *ProductRepository) BulkUpdate(ctx context.Context, products []*models.Product) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, product := range products {
			if err := tx.Save(product).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

