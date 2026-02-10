package repositories

import (
	"context"
	"errors"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BrandRepository implements BrandRepositoryInterface
type BrandRepository struct {
	db *gorm.DB
}

// NewBrandRepository creates a new BrandRepository
func NewBrandRepository(db *gorm.DB) *BrandRepository {
	return &BrandRepository{db: db}
}

// Create creates a new brand
func (r *BrandRepository) Create(ctx context.Context, brand *models.Brand) error {
	return r.db.WithContext(ctx).Create(brand).Error
}

// GetByID retrieves a brand by ID
func (r *BrandRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Brand, error) {
	var brand models.Brand
	err := r.db.WithContext(ctx).First(&brand, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &brand, err
}

// GetBySlug retrieves a brand by slug
func (r *BrandRepository) GetBySlug(ctx context.Context, slug string) (*models.Brand, error) {
	var brand models.Brand
	err := r.db.WithContext(ctx).First(&brand, "slug = ?", slug).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &brand, err
}

// Update updates a brand
func (r *BrandRepository) Update(ctx context.Context, brand *models.Brand) error {
	return r.db.WithContext(ctx).Save(brand).Error
}

// Delete soft-deletes a brand
func (r *BrandRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Brand{}, "id = ?", id).Error
}

// List retrieves brands with pagination
func (r *BrandRepository) List(ctx context.Context, opts ListOptions) ([]*models.Brand, int64, error) {
	var brands []*models.Brand
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Brand{})

	for key, value := range opts.Filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Offset(opts.Offset()).
		Limit(opts.Limit()).
		Order(opts.SortBy + " " + opts.SortOrder).
		Find(&brands).Error

	return brands, total, err
}

// UpdateProductCount updates the product count for a brand
func (r *BrandRepository) UpdateProductCount(ctx context.Context, brandID uuid.UUID) error {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Product{}).
		Where("brand_id = ?", brandID).
		Count(&count).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&models.Brand{}).
		Where("id = ?", brandID).
		Update("product_count", count).Error
}

// ListActive retrieves all active brands
func (r *BrandRepository) ListActive(ctx context.Context) ([]*models.Brand, error) {
	var brands []*models.Brand
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC, name ASC").
		Find(&brands).Error
	return brands, err
}

// ListFeatured retrieves featured brands
func (r *BrandRepository) ListFeatured(ctx context.Context, limit int) ([]*models.Brand, error) {
	var brands []*models.Brand
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND is_featured = ?", true, true).
		Order("sort_order ASC").
		Limit(limit).
		Find(&brands).Error
	return brands, err
}

