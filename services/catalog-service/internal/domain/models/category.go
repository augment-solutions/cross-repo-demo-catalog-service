package models

import (
	"time"

	"github.com/google/uuid"
)

// Category represents a product category
type Category struct {
	ID              uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name            string      `json:"name" gorm:"not null"`
	Slug            string      `json:"slug" gorm:"uniqueIndex;not null"`
	Description     string      `json:"description"`
	ParentID        *uuid.UUID  `json:"parent_id" gorm:"type:uuid;index"`
	Parent          *Category   `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children        []Category  `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Level           int         `json:"level" gorm:"default:0"`
	Path            string      `json:"path"`
	ImageURL        string      `json:"image_url"`
	IconURL         string      `json:"icon_url"`
	IsActive        bool        `json:"is_active" gorm:"default:true"`
	IsFeatured      bool        `json:"is_featured" gorm:"default:false"`
	SortOrder       int         `json:"sort_order" gorm:"default:0"`
	MetaTitle       string      `json:"meta_title"`
	MetaDescription string      `json:"meta_description"`
	MetaKeywords    string      `json:"meta_keywords"`
	ProductCount    int         `json:"product_count" gorm:"-"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	DeletedAt       *time.Time  `json:"deleted_at" gorm:"index"`
}

// TableName returns the table name for Category
func (Category) TableName() string {
	return "categories"
}

// BeforeCreate sets the UUID before creating
func (c *Category) BeforeCreate() error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// IsRoot returns true if this is a root category
func (c *Category) IsRoot() bool {
	return c.ParentID == nil
}

// HasChildren returns true if this category has children
func (c *Category) HasChildren() bool {
	return len(c.Children) > 0
}

// GetFullPath returns the full category path
func (c *Category) GetFullPath() string {
	if c.Path != "" {
		return c.Path + "/" + c.Slug
	}
	return c.Slug
}

// CategoryTree represents a category tree structure
type CategoryTree struct {
	Category *Category      `json:"category"`
	Children []*CategoryTree `json:"children,omitempty"`
}

// BuildCategoryTree builds a tree structure from flat categories
func BuildCategoryTree(categories []Category) []*CategoryTree {
	categoryMap := make(map[uuid.UUID]*CategoryTree)
	var roots []*CategoryTree

	for i := range categories {
		categoryMap[categories[i].ID] = &CategoryTree{
			Category: &categories[i],
			Children: make([]*CategoryTree, 0),
		}
	}

	for i := range categories {
		node := categoryMap[categories[i].ID]
		if categories[i].ParentID == nil {
			roots = append(roots, node)
		} else if parent, ok := categoryMap[*categories[i].ParentID]; ok {
			parent.Children = append(parent.Children, node)
		}
	}

	return roots
}

