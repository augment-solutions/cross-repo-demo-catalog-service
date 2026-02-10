package models

import (
	"time"

	"github.com/google/uuid"
)

// MediaType represents the type of media
type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
	MediaTypeDoc   MediaType = "document"
)

// Media represents product media (images, videos, etc.)
type Media struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProductID   uuid.UUID  `json:"product_id" gorm:"type:uuid;not null;index"`
	Type        MediaType  `json:"type" gorm:"not null;default:image"`
	URL         string     `json:"url" gorm:"not null"`
	ThumbnailURL string    `json:"thumbnail_url"`
	AltText     string     `json:"alt_text"`
	Title       string     `json:"title"`
	Width       int        `json:"width"`
	Height      int        `json:"height"`
	Size        int64      `json:"size"`
	MimeType    string     `json:"mime_type"`
	IsPrimary   bool       `json:"is_primary" gorm:"default:false"`
	SortOrder   int        `json:"sort_order" gorm:"default:0"`
	Metadata    string     `json:"metadata"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TableName returns the table name for Media
func (Media) TableName() string {
	return "product_media"
}

// BeforeCreate sets the UUID before creating
func (m *Media) BeforeCreate() error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// IsImage returns true if media is an image
func (m *Media) IsImage() bool {
	return m.Type == MediaTypeImage
}

// GetDisplayURL returns the display URL
func (m *Media) GetDisplayURL() string {
	if m.ThumbnailURL != "" {
		return m.ThumbnailURL
	}
	return m.URL
}

