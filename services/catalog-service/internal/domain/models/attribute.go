package models

import (
	"time"

	"github.com/google/uuid"
)

// AttributeType represents the type of attribute
type AttributeType string

const (
	AttributeTypeText     AttributeType = "text"
	AttributeTypeNumber   AttributeType = "number"
	AttributeTypeBoolean  AttributeType = "boolean"
	AttributeTypeSelect   AttributeType = "select"
	AttributeTypeMulti    AttributeType = "multiselect"
	AttributeTypeDate     AttributeType = "date"
	AttributeTypeColor    AttributeType = "color"
)

// Attribute represents a product attribute definition
type Attribute struct {
	ID            uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name          string          `json:"name" gorm:"not null"`
	Code          string          `json:"code" gorm:"uniqueIndex;not null"`
	Type          AttributeType   `json:"type" gorm:"not null;default:text"`
	Description   string          `json:"description"`
	IsRequired    bool            `json:"is_required" gorm:"default:false"`
	IsFilterable  bool            `json:"is_filterable" gorm:"default:false"`
	IsSearchable  bool            `json:"is_searchable" gorm:"default:false"`
	IsVisible     bool            `json:"is_visible" gorm:"default:true"`
	SortOrder     int             `json:"sort_order" gorm:"default:0"`
	DefaultValue  string          `json:"default_value"`
	Options       []AttributeOption `json:"options" gorm:"foreignKey:AttributeID"`
	ValidationRules string        `json:"validation_rules"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// AttributeOption represents an option for select/multiselect attributes
type AttributeOption struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AttributeID uuid.UUID `json:"attribute_id" gorm:"type:uuid;not null;index"`
	Value       string    `json:"value" gorm:"not null"`
	Label       string    `json:"label" gorm:"not null"`
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	IsDefault   bool      `json:"is_default" gorm:"default:false"`
	Metadata    string    `json:"metadata"`
}

// TableName returns the table name for Attribute
func (Attribute) TableName() string {
	return "attributes"
}

// TableName returns the table name for AttributeOption
func (AttributeOption) TableName() string {
	return "attribute_options"
}

// BeforeCreate sets the UUID before creating
func (a *Attribute) BeforeCreate() error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// IsSelectType returns true if attribute has options
func (a *Attribute) IsSelectType() bool {
	return a.Type == AttributeTypeSelect || a.Type == AttributeTypeMulti
}

