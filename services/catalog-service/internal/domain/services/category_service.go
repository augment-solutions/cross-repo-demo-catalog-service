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

// CategoryService handles category operations
type CategoryService struct {
	categoryRepo   *repositories.CategoryRepository
	categoryCache  *cache.CategoryCache
	eventPublisher *EventPublisher
	logger         *zap.Logger
}

// NewCategoryService creates a new CategoryService
func NewCategoryService(
	categoryRepo *repositories.CategoryRepository,
	categoryCache *cache.CategoryCache,
	eventPublisher *EventPublisher,
	logger *zap.Logger,
) *CategoryService {
	return &CategoryService{
		categoryRepo:   categoryRepo,
		categoryCache:  categoryCache,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// CreateCategory creates a new category
func (s *CategoryService) CreateCategory(ctx context.Context, req *models.CreateCategoryRequest) (*models.Category, error) {
	ctx, span := otel.StartSpan(ctx, "CategoryService.CreateCategory")
	defer span.End()
	span.SetAttributes(attribute.String("name", req.Name))

	category := &models.Category{
		Name:            req.Name,
		Slug:            s.generateSlug(req.Name, req.Slug),
		Description:     req.Description,
		ParentID:        req.ParentID,
		ImageURL:        req.ImageURL,
		IsActive:        req.IsActive,
		IsFeatured:      req.IsFeatured,
		MetaTitle:       req.MetaTitle,
		MetaDescription: req.MetaDescription,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		s.logger.Error("failed to create category", zap.Error(err))
		return nil, err
	}

	// Publish event
	event, _ := models.NewEvent(models.EventTypeCategoryCreated, category.ID, models.CategoryCreatedData{
		CategoryID: category.ID,
		Name:       category.Name,
		Slug:       category.Slug,
		ParentID:   category.ParentID,
	})
	if err := s.eventPublisher.Publish(ctx, event); err != nil {
		s.logger.Warn("failed to publish category created event", zap.Error(err))
	}

	// Invalidate category tree cache
	s.categoryCache.InvalidateTree(ctx)

	otel.CategoriesCreated.Add(ctx, 1)
	s.logger.Info("category created", zap.String("id", category.ID.String()))
	return category, nil
}

// GetCategory retrieves a category by ID
func (s *CategoryService) GetCategory(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	ctx, span := otel.StartSpan(ctx, "CategoryService.GetCategory")
	defer span.End()

	// Try cache first
	category, err := s.categoryCache.Get(ctx, id)
	if err == nil && category != nil {
		return category, nil
	}

	category, err = s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if category != nil {
		s.categoryCache.Set(ctx, category)
	}

	return category, nil
}

// UpdateCategory updates a category
func (s *CategoryService) UpdateCategory(ctx context.Context, id uuid.UUID, req *models.UpdateCategoryRequest) (*models.Category, error) {
	ctx, span := otel.StartSpan(ctx, "CategoryService.UpdateCategory")
	defer span.End()

	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, nil
	}

	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Description != nil {
		category.Description = *req.Description
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	s.categoryCache.Delete(ctx, id)
	s.categoryCache.InvalidateTree(ctx)

	return category, nil
}

// DeleteCategory deletes a category
func (s *CategoryService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	ctx, span := otel.StartSpan(ctx, "CategoryService.DeleteCategory")
	defer span.End()

	if err := s.categoryRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.categoryCache.Delete(ctx, id)
	s.categoryCache.InvalidateTree(ctx)
	return nil
}

// ListCategories lists categories
func (s *CategoryService) ListCategories(ctx context.Context, opts repositories.ListOptions) ([]*models.Category, int64, error) {
	return s.categoryRepo.List(ctx, opts)
}

// GetTree returns the category tree
func (s *CategoryService) GetTree(ctx context.Context) ([]*models.CategoryTree, error) {
	return s.categoryRepo.GetTree(ctx)
}

func (s *CategoryService) generateSlug(name, slug string) string {
	if slug != "" {
		return slug
	}
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}

