package models

import (
	"github.com/google/uuid"
)

// CreateProductRequest represents a request to create a product
type CreateProductRequest struct {
	SKU              string             `json:"sku" binding:"required"`
	Name             string             `json:"name" binding:"required"`
	Slug             string             `json:"slug"`
	Description      string             `json:"description"`
	ShortDescription string             `json:"short_description"`
	CategoryID       *uuid.UUID         `json:"category_id"`
	BrandID          *uuid.UUID         `json:"brand_id"`
	Status           ProductStatus      `json:"status"`
	IsFeatured       bool               `json:"is_featured"`
	Weight           float64            `json:"weight"`
	MetaTitle        string             `json:"meta_title"`
	MetaDescription  string             `json:"meta_description"`
	Tags             []string           `json:"tags"`
	Attributes       []AttributeValue   `json:"attributes"`
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
	SKU            string          `json:"sku" binding:"required"`
	Name           string          `json:"name"`
	Price          float64         `json:"price" binding:"required"`
	CompareAtPrice float64         `json:"compare_at_price"`
	CostPrice      float64         `json:"cost_price"`
	Weight         float64         `json:"weight"`
	Barcode        string          `json:"barcode"`
	IsDefault      bool            `json:"is_default"`
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

// ProductResponse represents a product response
type ProductResponse struct {
	Product    *Product       `json:"product"`
	Category   *Category      `json:"category,omitempty"`
	Brand      *Brand         `json:"brand,omitempty"`
	Inventory  *InventoryInfo `json:"inventory,omitempty"`
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
	Format   string `json:"format" binding:"required"`
	Data     string `json:"data" binding:"required"`
	Mode     string `json:"mode"`
}

// ExportRequest represents an export request
type ExportRequest struct {
	Format     string     `json:"format"`
	CategoryID *uuid.UUID `json:"category_id"`
	BrandID    *uuid.UUID `json:"brand_id"`
	Status     string     `json:"status"`
}

