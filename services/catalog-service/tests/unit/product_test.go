package unit

import (
	"context"
	"testing"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductRepository is a mock product repository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) GetBySKU(ctx context.Context, sku string) (*models.Product, error) {
	args := m.Called(ctx, sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) Update(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestProduct_Creation(t *testing.T) {
	product := &models.Product{
		ID:     uuid.New(),
		SKU:    "TEST-001",
		Name:   "Test Product",
		Slug:   "test-product",
		Status: models.ProductStatusDraft,
	}

	assert.NotEmpty(t, product.ID)
	assert.Equal(t, "TEST-001", product.SKU)
	assert.Equal(t, "Test Product", product.Name)
	assert.Equal(t, models.ProductStatusDraft, product.Status)
}

func TestProduct_Validation(t *testing.T) {
	tests := []struct {
		name    string
		product *models.Product
		wantErr bool
	}{
		{
			name: "valid product",
			product: &models.Product{
				SKU:  "TEST-001",
				Name: "Test Product",
				Slug: "test-product",
			},
			wantErr: false,
		},
		{
			name: "missing SKU",
			product: &models.Product{
				Name: "Test Product",
				Slug: "test-product",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			product: &models.Product{
				SKU:  "TEST-001",
				Slug: "test-product",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProduct(tt.product)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProduct_Status(t *testing.T) {
	product := &models.Product{
		ID:     uuid.New(),
		Status: models.ProductStatusDraft,
	}

	assert.Equal(t, models.ProductStatusDraft, product.Status)

	product.Status = models.ProductStatusActive
	assert.Equal(t, models.ProductStatusActive, product.Status)

	product.Status = models.ProductStatusInactive
	assert.Equal(t, models.ProductStatusInactive, product.Status)
}

func TestProduct_Variants(t *testing.T) {
	product := &models.Product{
		ID:   uuid.New(),
		Name: "Test Product",
	}

	variant := models.Variant{
		ID:        uuid.New(),
		ProductID: product.ID,
		SKU:       "TEST-001-VAR1",
		Price:     29.99,
		IsDefault: true,
	}

	product.Variants = append(product.Variants, variant)

	assert.Len(t, product.Variants, 1)
	assert.Equal(t, 29.99, product.Variants[0].Price)
	assert.True(t, product.Variants[0].IsDefault)
}

func validateProduct(p *models.Product) error {
	if p.SKU == "" {
		return assert.AnError
	}
	if p.Name == "" {
		return assert.AnError
	}
	return nil
}

