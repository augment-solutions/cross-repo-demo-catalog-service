package repositories

import (
	"context"
	"errors"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MediaRepository implements MediaRepositoryInterface
type MediaRepository struct {
	db *gorm.DB
}

// NewMediaRepository creates a new MediaRepository
func NewMediaRepository(db *gorm.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

// Create creates a new media
func (r *MediaRepository) Create(ctx context.Context, media *models.Media) error {
	return r.db.WithContext(ctx).Create(media).Error
}

// GetByID retrieves a media by ID
func (r *MediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Media, error) {
	var media models.Media
	err := r.db.WithContext(ctx).First(&media, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &media, err
}

// Update updates a media
func (r *MediaRepository) Update(ctx context.Context, media *models.Media) error {
	return r.db.WithContext(ctx).Save(media).Error
}

// Delete deletes a media
func (r *MediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Media{}, "id = ?", id).Error
}

// ListByProduct retrieves all media for a product
func (r *MediaRepository) ListByProduct(ctx context.Context, productID uuid.UUID) ([]*models.Media, error) {
	var media []*models.Media
	err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("sort_order ASC").
		Find(&media).Error
	return media, err
}

// SetPrimary sets a media as primary for a product
func (r *MediaRepository) SetPrimary(ctx context.Context, productID, mediaID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Unset all primary
		if err := tx.Model(&models.Media{}).
			Where("product_id = ?", productID).
			Update("is_primary", false).Error; err != nil {
			return err
		}
		// Set the new primary
		return tx.Model(&models.Media{}).
			Where("id = ?", mediaID).
			Update("is_primary", true).Error
	})
}

// UpdateSortOrder updates the sort order for media
func (r *MediaRepository) UpdateSortOrder(ctx context.Context, productID uuid.UUID, mediaIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range mediaIDs {
			if err := tx.Model(&models.Media{}).
				Where("id = ? AND product_id = ?", id, productID).
				Update("sort_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetPrimaryByProduct retrieves the primary media for a product
func (r *MediaRepository) GetPrimaryByProduct(ctx context.Context, productID uuid.UUID) (*models.Media, error) {
	var media models.Media
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND is_primary = ?", productID, true).
		First(&media).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &media, err
}

// CountByProduct counts media for a product
func (r *MediaRepository) CountByProduct(ctx context.Context, productID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Media{}).
		Where("product_id = ?", productID).
		Count(&count).Error
	return count, err
}

