package models

import (
	"time"

	"github.com/google/uuid"
)

// ProductStatus represents the status of a product
type ProductStatus string

const (
	ProductStatusDraft     ProductStatus = "draft"
	ProductStatusActive    ProductStatus = "active"
	ProductStatusInactive  ProductStatus = "inactive"
	ProductStatusArchived  ProductStatus = "archived"
)

// Product represents a product in the catalog
type Product struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SKU             string          `json:"sku" gorm:"uniqueIndex;not null"`
	Name            string          `json:"name" gorm:"not null"`
	Slug            string          `json:"slug" gorm:"uniqueIndex;not null"`
	Description     string          `json:"description"`
	ShortDescription string         `json:"short_description"`
	CategoryID      *uuid.UUID      `json:"category_id" gorm:"type:uuid;index"`
	BrandID         *uuid.UUID      `json:"brand_id" gorm:"type:uuid;index"`
	Status          ProductStatus   `json:"status" gorm:"default:draft"`
	IsFeatured      bool            `json:"is_featured" gorm:"default:false"`
	IsNew           bool            `json:"is_new" gorm:"default:false"`
	Weight          float64         `json:"weight"`
	Length          float64         `json:"length"`
	Width           float64         `json:"width"`
	Height          float64         `json:"height"`
	MetaTitle       string          `json:"meta_title"`
	MetaDescription string          `json:"meta_description"`
	MetaKeywords    string          `json:"meta_keywords"`
	Attributes      []ProductAttribute `json:"attributes" gorm:"foreignKey:ProductID"`
	Variants        []Variant       `json:"variants" gorm:"foreignKey:ProductID"`
	Media           []Media         `json:"media" gorm:"foreignKey:ProductID"`
	Tags            []string        `json:"tags" gorm:"type:text[]"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       *time.Time      `json:"deleted_at" gorm:"index"`
}

// ProductAttribute represents a product attribute
type ProductAttribute struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProductID   uuid.UUID  `json:"product_id" gorm:"type:uuid;not null;index"`
	AttributeID uuid.UUID  `json:"attribute_id" gorm:"type:uuid;not null;index"`
	Value       string     `json:"value" gorm:"not null"`
	CreatedAt   time.Time  `json:"created_at"`
}

// TableName returns the table name for Product
func (Product) TableName() string {
	return "products"
}

// TableName returns the table name for ProductAttribute
func (ProductAttribute) TableName() string {
	return "product_attributes"
}

// BeforeCreate sets the UUID before creating
func (p *Product) BeforeCreate() error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// IsActive returns true if the product is active
func (p *Product) IsActive() bool {
	return p.Status == ProductStatusActive
}

// CanBeOrdered returns true if the product can be ordered
func (p *Product) CanBeOrdered() bool {
	return p.Status == ProductStatusActive && len(p.Variants) > 0
}

