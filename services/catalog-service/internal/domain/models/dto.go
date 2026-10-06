package models

import (
	"time"

	"github.com/google/uuid"
)

// CreateProductRequest represents a request to create a product
type CreateProductRequest struct {
	SKU              string                 `json:"sku" binding:"required"`
	Name             string                 `json:"name" binding:"required"`
	Slug             string                 `json:"slug"`
	Description      string                 `json:"description"`
	ShortDescription string                 `json:"short_description"`
	CategoryID       *uuid.UUID             `json:"category_id"`
	BrandID          *uuid.UUID             `json:"brand_id"`
	Status           ProductStatus          `json:"status"`
	IsFeatured       bool                   `json:"is_featured"`
	Weight           float64                `json:"weight"`
	MetaTitle        string                 `json:"meta_title"`
	MetaDescription  string                 `json:"meta_description"`
	Tags             []string               `json:"tags"`
	Attributes       []AttributeValue       `json:"attributes"`
	Variants         []CreateVariantRequest `json:"variants"`
}

// UpdateProductRequest represents a request to update a product
type UpdateProductRequest struct {
	Name             *string        `json:"name"`
	Description      *string        `json:"description"`
	ShortDescription *string        `json:"short_description"`
	CategoryID       *uuid.UUID     `json:"category_id"`
	BrandID          *uuid.UUID     `json:"brand_id"`
	Status           *ProductStatus `json:"status"`
	IsFeatured       *bool          `json:"is_featured"`
	Weight           *float64       `json:"weight"`
	MetaTitle        *string        `json:"meta_title"`
	MetaDescription  *string        `json:"meta_description"`
	Tags             []string       `json:"tags"`
}

// AttributeValue represents an attribute value in a request
type AttributeValue struct {
	AttributeID uuid.UUID `json:"attribute_id" binding:"required"`
	Value       string    `json:"value" binding:"required"`
}

// CreateVariantRequest represents a request to create a variant
type CreateVariantRequest struct {
	SKU            string               `json:"sku" binding:"required"`
	Name           string               `json:"name"`
	Price          float64              `json:"price" binding:"required"`
	CompareAtPrice float64              `json:"compare_at_price"`
	CostPrice      float64              `json:"cost_price"`
	Weight         float64              `json:"weight"`
	Barcode        string               `json:"barcode"`
	IsDefault      bool                 `json:"is_default"`
	Options        []VariantOptionValue `json:"options"`
}

// VariantOptionValue represents a variant option in a request
type VariantOptionValue struct {
	Name  string `json:"name" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// CreateCategoryRequest represents a request to create a category
type CreateCategoryRequest struct {
	Name            string     `json:"name" binding:"required"`
	Slug            string     `json:"slug"`
	Description     string     `json:"description"`
	ParentID        *uuid.UUID `json:"parent_id"`
	ImageURL        string     `json:"image_url"`
	IsActive        bool       `json:"is_active"`
	IsFeatured      bool       `json:"is_featured"`
	MetaTitle       string     `json:"meta_title"`
	MetaDescription string     `json:"meta_description"`
}

// UpdateCategoryRequest represents a request to update a category
type UpdateCategoryRequest struct {
	Name            *string    `json:"name"`
	Description     *string    `json:"description"`
	ParentID        *uuid.UUID `json:"parent_id"`
	ImageURL        *string    `json:"image_url"`
	IsActive        *bool      `json:"is_active"`
	IsFeatured      *bool      `json:"is_featured"`
	MetaTitle       *string    `json:"meta_title"`
	MetaDescription *string    `json:"meta_description"`
}

// CreateBrandRequest represents a request to create a brand
type CreateBrandRequest struct {
	Name            string `json:"name" binding:"required"`
	Slug            string `json:"slug"`
	Description     string `json:"description"`
	LogoURL         string `json:"logo_url"`
	WebsiteURL      string `json:"website_url"`
	IsActive        bool   `json:"is_active"`
	IsFeatured      bool   `json:"is_featured"`
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
}

// ProductPricing represents the pricing of a product's default variant.
type ProductPricing struct {
	Price          float64 `json:"price"`
	CompareAtPrice float64 `json:"compare_at_price"`
	CostPrice      float64 `json:"cost_price"`
}

// ProductVariantResponse contains variant details without product-level pricing.
type ProductVariantResponse struct {
	ID           uuid.UUID       `json:"id"`
	ProductID    uuid.UUID       `json:"product_id"`
	SKU          string          `json:"sku"`
	Name         string          `json:"name"`
	Weight       float64         `json:"weight"`
	Barcode      string          `json:"barcode"`
	IsDefault    bool            `json:"is_default"`
	IsActive     bool            `json:"is_active"`
	SortOrder    int             `json:"sort_order"`
	ImageURL     string          `json:"image_url"`
	Options      []VariantOption `json:"options"`
	InventoryQty int             `json:"inventory_qty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    *time.Time      `json:"deleted_at"`
}

// ProductDetailsResponse keeps pricing at the product root without changing storage models.
type ProductDetailsResponse struct {
	*Product
	Pricing  ProductPricing           `json:"pricing"`
	Variants []ProductVariantResponse `json:"variants"`
}

// NewProductDetailsResponse uses the default variant, or the first variant if none is default.
// Products without variants still include pricing with zero-valued fields.
func NewProductDetailsResponse(product *Product) *ProductDetailsResponse {
	if product == nil {
		return nil
	}

	response := &ProductDetailsResponse{Product: product}
	if product.Variants == nil {
		return response
	}

	response.Variants = make([]ProductVariantResponse, len(product.Variants))
	var pricingVariant *Variant
	for i := range product.Variants {
		variant := &product.Variants[i]
		if pricingVariant == nil || (variant.IsDefault && !pricingVariant.IsDefault) {
			pricingVariant = variant
		}
		response.Variants[i] = ProductVariantResponse{
			ID:           variant.ID,
			ProductID:    variant.ProductID,
			SKU:          variant.SKU,
			Name:         variant.Name,
			Weight:       variant.Weight,
			Barcode:      variant.Barcode,
			IsDefault:    variant.IsDefault,
			IsActive:     variant.IsActive,
			SortOrder:    variant.SortOrder,
			ImageURL:     variant.ImageURL,
			Options:      variant.Options,
			InventoryQty: variant.InventoryQty,
			CreatedAt:    variant.CreatedAt,
			UpdatedAt:    variant.UpdatedAt,
			DeletedAt:    variant.DeletedAt,
		}
	}
	if pricingVariant != nil {
		response.Pricing = ProductPricing{
			Price:          pricingVariant.Price,
			CompareAtPrice: pricingVariant.CompareAtPrice,
			CostPrice:      pricingVariant.CostPrice,
		}
	}
	return response
}

// NewProductDetailsResponses converts a product list while preserving nil and empty slices.
func NewProductDetailsResponses(products []*Product) []*ProductDetailsResponse {
	if products == nil {
		return nil
	}
	responses := make([]*ProductDetailsResponse, len(products))
	for i, product := range products {
		responses[i] = NewProductDetailsResponse(product)
	}
	return responses
}

// ProductResponse represents an enriched product response.
type ProductResponse struct {
	Product   *ProductDetailsResponse `json:"product"`
	Category  *Category               `json:"category,omitempty"`
	Brand     *Brand                  `json:"brand,omitempty"`
	Inventory *InventoryInfo          `json:"inventory,omitempty"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query      string            `json:"query"`
	CategoryID *uuid.UUID        `json:"category_id"`
	BrandID    *uuid.UUID        `json:"brand_id"`
	MinPrice   *float64          `json:"min_price"`
	MaxPrice   *float64          `json:"max_price"`
	Filters    map[string]string `json:"filters"`
	Sort       string            `json:"sort"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
}

// ImportRequest represents an import request
type ImportRequest struct {
	Format string `json:"format" binding:"required"`
	Data   string `json:"data" binding:"required"`
	Mode   string `json:"mode"`
}

// ExportRequest represents an export request
type ExportRequest struct {
	Format     string     `json:"format"`
	CategoryID *uuid.UUID `json:"category_id"`
	BrandID    *uuid.UUID `json:"brand_id"`
	Status     string     `json:"status"`
}
