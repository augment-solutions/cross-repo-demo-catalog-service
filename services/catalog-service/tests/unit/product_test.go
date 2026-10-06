package unit

import (
	"context"
	"encoding/json"
	"testing"
	"time"

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

func TestProductDetailsResponse_Pricing(t *testing.T) {
	first := models.Variant{SKU: "FIRST", Price: 19.99, CompareAtPrice: 24.99, CostPrice: 10}
	defaultVariant := models.Variant{SKU: "DEFAULT", Price: 29.99, CompareAtPrice: 39.99, CostPrice: 15, IsDefault: true}
	tests := []struct {
		name     string
		variants []models.Variant
		pricing  models.ProductPricing
	}{
		{"default variant", []models.Variant{first, defaultVariant}, models.ProductPricing{Price: 29.99, CompareAtPrice: 39.99, CostPrice: 15}},
		{"first variant fallback", []models.Variant{first}, models.ProductPricing{Price: 19.99, CompareAtPrice: 24.99, CostPrice: 10}},
		{"first default wins", []models.Variant{defaultVariant, {Price: 99, IsDefault: true}}, models.ProductPricing{Price: 29.99, CompareAtPrice: 39.99, CostPrice: 15}},
		{"zero-valued default", []models.Variant{first, {IsDefault: true}}, models.ProductPricing{}},
		{"no variants", nil, models.ProductPricing{}},
		{"empty variants", []models.Variant{}, models.ProductPricing{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := &models.Product{ID: uuid.New(), Name: "Test Product", Variants: tt.variants}
			response := models.NewProductDetailsResponse(product)
			assert.Equal(t, tt.pricing, response.Pricing)
			assert.Equal(t, product.Variants, tt.variants)

			data, err := json.Marshal(response)
			if !assert.NoError(t, err) {
				return
			}
			var body map[string]interface{}
			if !assert.NoError(t, json.Unmarshal(data, &body)) {
				return
			}
			assert.Equal(t, product.ID.String(), body["id"])
			assert.Equal(t, product.Name, body["name"])
			assert.Equal(t, map[string]interface{}{
				"price": tt.pricing.Price, "compare_at_price": tt.pricing.CompareAtPrice, "cost_price": tt.pricing.CostPrice,
			}, body["pricing"])
			for _, field := range []string{"price", "compare_at_price", "cost_price"} {
				assert.NotContains(t, body, field)
			}
			if tt.variants == nil {
				assert.Nil(t, body["variants"])
			} else {
				variants := body["variants"].([]interface{})
				assert.Len(t, variants, len(tt.variants))
				for i, variant := range variants {
					fields := variant.(map[string]interface{})
					assert.Equal(t, tt.variants[i].SKU, fields["sku"])
					for _, field := range []string{"price", "compare_at_price", "cost_price"} {
						assert.NotContains(t, fields, field)
					}
				}
			}
		})
	}
}

func TestProductDetailsResponse_PreservesVariantDetailsAndStorage(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	variant := models.Variant{
		ID: uuid.New(), ProductID: uuid.New(), SKU: "TEST-VAR", Name: "Blue",
		Price: 29.99, CompareAtPrice: 39.99, CostPrice: 15, Weight: 1.5, Barcode: "12345",
		IsDefault: true, IsActive: true, SortOrder: 2, ImageURL: "https://example.com/blue.png",
		Options:      []models.VariantOption{{ID: uuid.New(), OptionName: "Color", OptionValue: "Blue"}},
		InventoryQty: 7, CreatedAt: now, UpdatedAt: now, DeletedAt: &now,
	}
	product := &models.Product{Variants: []models.Variant{variant}}
	responseJSON, err := json.Marshal(models.NewProductDetailsResponse(product))
	if !assert.NoError(t, err) {
		return
	}
	storageJSON, err := json.Marshal(product)
	if !assert.NoError(t, err) {
		return
	}
	var responseBody, storageBody map[string]interface{}
	assert.NoError(t, json.Unmarshal(responseJSON, &responseBody))
	assert.NoError(t, json.Unmarshal(storageJSON, &storageBody))
	responseVariant := responseBody["variants"].([]interface{})[0].(map[string]interface{})
	storageVariant := storageBody["variants"].([]interface{})[0].(map[string]interface{})
	for _, field := range []string{"price", "compare_at_price", "cost_price"} {
		assert.Contains(t, storageVariant, field)
		delete(storageVariant, field)
	}
	assert.Equal(t, storageVariant, responseVariant)
	assert.NotContains(t, storageBody, "pricing")
	var restored models.Product
	assert.NoError(t, json.Unmarshal(storageJSON, &restored))
	assert.Equal(t, product, &restored)
}

func TestProductDetailsResponses_ListAndEnrichment(t *testing.T) {
	assert.Nil(t, models.NewProductDetailsResponse(nil))
	assert.Nil(t, models.NewProductDetailsResponses(nil))
	assert.NotNil(t, models.NewProductDetailsResponses([]*models.Product{}))
	product := &models.Product{Variants: []models.Variant{{Price: 29.99}}}
	responses := models.NewProductDetailsResponses([]*models.Product{product, nil})
	assert.Len(t, responses, 2)
	assert.Equal(t, 29.99, responses[0].Pricing.Price)
	assert.Nil(t, responses[1])

	enriched := models.ProductResponse{Product: responses[0]}
	data, err := json.Marshal(enriched)
	if !assert.NoError(t, err) {
		return
	}
	var body map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &body))
	assert.Contains(t, body["product"], "pricing")
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
