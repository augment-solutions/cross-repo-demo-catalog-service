package models

import (
	"time"

	"github.com/google/uuid"
)

// Variant represents a product variant
type Variant struct {
	ID              uuid.UUID         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProductID       uuid.UUID         `json:"product_id" gorm:"type:uuid;not null;index"`
	SKU             string            `json:"sku" gorm:"uniqueIndex;not null"`
	Name            string            `json:"name"`
	Price           float64           `json:"price" gorm:"type:decimal(10,2);not null"`
	CompareAtPrice  float64           `json:"compare_at_price" gorm:"type:decimal(10,2)"`
	CostPrice       float64           `json:"cost_price" gorm:"type:decimal(10,2)"`
	Weight          float64           `json:"weight"`
	Barcode         string            `json:"barcode"`
	IsDefault       bool              `json:"is_default" gorm:"default:false"`
	IsActive        bool              `json:"is_active" gorm:"default:true"`
	SortOrder       int               `json:"sort_order" gorm:"default:0"`
	ImageURL        string            `json:"image_url"`
	Options         []VariantOption   `json:"options" gorm:"foreignKey:VariantID"`
	InventoryQty    int               `json:"inventory_qty" gorm:"-"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	DeletedAt       *time.Time        `json:"deleted_at" gorm:"index"`
}

// VariantOption represents a variant option value
type VariantOption struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	VariantID   uuid.UUID `json:"variant_id" gorm:"type:uuid;not null;index"`
	OptionName  string    `json:"option_name" gorm:"not null"`
	OptionValue string    `json:"option_value" gorm:"not null"`
}

// TableName returns the table name for Variant
func (Variant) TableName() string {
	return "variants"
}

// TableName returns the table name for VariantOption
func (VariantOption) TableName() string {
	return "variant_options"
}

// BeforeCreate sets the UUID before creating
func (v *Variant) BeforeCreate() error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

// GetDiscount returns the discount percentage
func (v *Variant) GetDiscount() float64 {
	if v.CompareAtPrice > 0 && v.CompareAtPrice > v.Price {
		return ((v.CompareAtPrice - v.Price) / v.CompareAtPrice) * 100
	}
	return 0
}

// IsOnSale returns true if variant is on sale
func (v *Variant) IsOnSale() bool {
	return v.CompareAtPrice > 0 && v.CompareAtPrice > v.Price
}

// GetMargin returns the profit margin
func (v *Variant) GetMargin() float64 {
	if v.CostPrice > 0 {
		return ((v.Price - v.CostPrice) / v.Price) * 100
	}
	return 0
}

