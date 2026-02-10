package models

import (
	"time"

	"github.com/google/uuid"
)

// Brand represents a product brand
type Brand struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name            string     `json:"name" gorm:"not null;uniqueIndex"`
	Slug            string     `json:"slug" gorm:"uniqueIndex;not null"`
	Description     string     `json:"description"`
	LogoURL         string     `json:"logo_url"`
	WebsiteURL      string     `json:"website_url"`
	IsActive        bool       `json:"is_active" gorm:"default:true"`
	IsFeatured      bool       `json:"is_featured" gorm:"default:false"`
	SortOrder       int        `json:"sort_order" gorm:"default:0"`
	MetaTitle       string     `json:"meta_title"`
	MetaDescription string     `json:"meta_description"`
	ProductCount    int        `json:"product_count" gorm:"-"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at" gorm:"index"`
}

// TableName returns the table name for Brand
func (Brand) TableName() string {
	return "brands"
}

// BeforeCreate sets the UUID before creating
func (b *Brand) BeforeCreate() error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

