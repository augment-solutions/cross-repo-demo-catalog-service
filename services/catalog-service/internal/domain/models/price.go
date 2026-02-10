package models

import (
	"time"

	"github.com/google/uuid"
)

// PriceType represents the type of price
type PriceType string

const (
	PriceTypeRegular    PriceType = "regular"
	PriceTypeSale       PriceType = "sale"
	PriceTypeClearance  PriceType = "clearance"
	PriceTypeMember     PriceType = "member"
	PriceTypeWholesale  PriceType = "wholesale"
)

// Price represents a product price
type Price struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	VariantID   uuid.UUID  `json:"variant_id" gorm:"type:uuid;not null;index"`
	Type        PriceType  `json:"type" gorm:"not null;default:regular"`
	Amount      float64    `json:"amount" gorm:"type:decimal(10,2);not null"`
	Currency    string     `json:"currency" gorm:"not null;default:USD"`
	MinQuantity int        `json:"min_quantity" gorm:"default:1"`
	MaxQuantity *int       `json:"max_quantity"`
	StartsAt    *time.Time `json:"starts_at"`
	EndsAt      *time.Time `json:"ends_at"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TableName returns the table name for Price
func (Price) TableName() string {
	return "prices"
}

// BeforeCreate sets the UUID before creating
func (p *Price) BeforeCreate() error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// IsValidNow returns true if price is currently valid
func (p *Price) IsValidNow() bool {
	now := time.Now()
	if p.StartsAt != nil && now.Before(*p.StartsAt) {
		return false
	}
	if p.EndsAt != nil && now.After(*p.EndsAt) {
		return false
	}
	return p.IsActive
}

// PriceHistory stores price history for analytics
type PriceHistory struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	VariantID uuid.UUID `json:"variant_id" gorm:"type:uuid;not null;index"`
	OldPrice  float64   `json:"old_price" gorm:"type:decimal(10,2)"`
	NewPrice  float64   `json:"new_price" gorm:"type:decimal(10,2);not null"`
	Currency  string    `json:"currency" gorm:"not null;default:USD"`
	Reason    string    `json:"reason"`
	ChangedBy string    `json:"changed_by"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName returns the table name for PriceHistory
func (PriceHistory) TableName() string {
	return "price_history"
}

