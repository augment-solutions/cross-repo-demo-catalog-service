package repositories

import (
	"context"
	"errors"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CategoryRepository implements CategoryRepositoryInterface
type CategoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new CategoryRepository
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Create creates a new category
func (r *CategoryRepository) Create(ctx context.Context, category *models.Category) error {
	if category.ParentID != nil {
		var parent models.Category
		if err := r.db.WithContext(ctx).First(&parent, "id = ?", *category.ParentID).Error; err != nil {
			return err
		}
		category.Level = parent.Level + 1
		category.Path = parent.GetFullPath()
	}
	return r.db.WithContext(ctx).Create(category).Error
}

// GetByID retrieves a category by ID
func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	var category models.Category
	err := r.db.WithContext(ctx).
		Preload("Children").
		First(&category, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &category, err
}

// GetBySlug retrieves a category by slug
func (r *CategoryRepository) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	var category models.Category
	err := r.db.WithContext(ctx).
		Preload("Children").
		First(&category, "slug = ?", slug).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &category, err
}

// Update updates a category
func (r *CategoryRepository) Update(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

// Delete soft-deletes a category
func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Category{}, "id = ?", id).Error
}

// List retrieves categories with pagination
func (r *CategoryRepository) List(ctx context.Context, opts ListOptions) ([]*models.Category, int64, error) {
	var categories []*models.Category
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Category{})

	for key, value := range opts.Filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	q := query.
		Offset(opts.Offset()).
		Limit(opts.Limit())

	if opts.SortBy != "" {
		q = q.Order(opts.SortBy + " " + opts.SortOrder)
	}

	err := q.Find(&categories).Error
	if err != nil {
		return nil, 0, err
	}

	// Calculate product counts for each category
	for _, category := range categories {
		var count int64
		if countErr := r.db.WithContext(ctx).Model(&models.Product{}).
			Where("category_id = ?", category.ID).
			Count(&count).Error; countErr == nil {
			category.ProductCount = int(count)
		}
	}

	return categories, total, nil
}

// ListChildren retrieves child categories
func (r *CategoryRepository) ListChildren(ctx context.Context, parentID uuid.UUID) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("sort_order ASC").
		Find(&categories).Error
	return categories, err
}

// GetTree retrieves the full category tree
func (r *CategoryRepository) GetTree(ctx context.Context) ([]*models.CategoryTree, error) {
	var categories []models.Category
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("level ASC, sort_order ASC").
		Find(&categories).Error
	if err != nil {
		return nil, err
	}

	// Calculate product counts for each category
	for i := range categories {
		var count int64
		if countErr := r.db.WithContext(ctx).Model(&models.Product{}).
			Where("category_id = ?", categories[i].ID).
			Count(&count).Error; countErr == nil {
			categories[i].ProductCount = int(count)
		}
	}

	return models.BuildCategoryTree(categories), nil
}

// UpdateProductCount updates the product count for a category
func (r *CategoryRepository) UpdateProductCount(ctx context.Context, categoryID uuid.UUID) error {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Product{}).
		Where("category_id = ?", categoryID).
		Count(&count).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&models.Category{}).
		Where("id = ?", categoryID).
		Update("product_count", count).Error
}

