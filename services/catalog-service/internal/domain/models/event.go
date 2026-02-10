package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of event
type EventType string

const (
	EventTypeProductCreated    EventType = "product.created"
	EventTypeProductUpdated    EventType = "product.updated"
	EventTypeProductDeleted    EventType = "product.deleted"
	EventTypeCategoryCreated   EventType = "category.created"
	EventTypeCategoryUpdated   EventType = "category.updated"
	EventTypeCategoryDeleted   EventType = "category.deleted"
	EventTypeBrandCreated      EventType = "brand.created"
	EventTypeBrandUpdated      EventType = "brand.updated"
	EventTypeInventoryUpdated  EventType = "inventory.updated"
	EventTypePriceUpdated      EventType = "price.updated"
	EventTypeSearchReindex     EventType = "search.reindex"
)

// Event represents a domain event
type Event struct {
	ID          uuid.UUID       `json:"id"`
	Type        EventType       `json:"type"`
	AggregateID uuid.UUID       `json:"aggregate_id"`
	Timestamp   time.Time       `json:"timestamp"`
	Version     int             `json:"version"`
	Data        json.RawMessage `json:"data"`
	Metadata    EventMetadata   `json:"metadata"`
}

// EventMetadata contains event metadata
type EventMetadata struct {
	CorrelationID string `json:"correlation_id,omitempty"`
	CausationID   string `json:"causation_id,omitempty"`
	UserID        string `json:"user_id,omitempty"`
	Source        string `json:"source"`
	TraceID       string `json:"trace_id,omitempty"`
	SpanID        string `json:"span_id,omitempty"`
}

// NewEvent creates a new event
func NewEvent(eventType EventType, aggregateID uuid.UUID, data interface{}) (*Event, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &Event{
		ID:          uuid.New(),
		Type:        eventType,
		AggregateID: aggregateID,
		Timestamp:   time.Now().UTC(),
		Version:     1,
		Data:        dataBytes,
		Metadata: EventMetadata{
			Source: "catalog-service",
		},
	}, nil
}

// ProductCreatedData represents data for product created event
type ProductCreatedData struct {
	ProductID   uuid.UUID `json:"product_id"`
	SKU         string    `json:"sku"`
	Name        string    `json:"name"`
	CategoryID  *uuid.UUID `json:"category_id,omitempty"`
	BrandID     *uuid.UUID `json:"brand_id,omitempty"`
}

// ProductUpdatedData represents data for product updated event
type ProductUpdatedData struct {
	ProductID     uuid.UUID          `json:"product_id"`
	ChangedFields map[string]interface{} `json:"changed_fields"`
}

// ProductDeletedData represents data for product deleted event
type ProductDeletedData struct {
	ProductID uuid.UUID `json:"product_id"`
	SKU       string    `json:"sku"`
}

// CategoryCreatedData represents data for category created event
type CategoryCreatedData struct {
	CategoryID uuid.UUID  `json:"category_id"`
	Name       string     `json:"name"`
	Slug       string     `json:"slug"`
	ParentID   *uuid.UUID `json:"parent_id,omitempty"`
}

// InventoryUpdatedData represents data for inventory updated event
type InventoryUpdatedData struct {
	SKU         string `json:"sku"`
	OldQuantity int    `json:"old_quantity"`
	NewQuantity int    `json:"new_quantity"`
	Reason      string `json:"reason"`
}

// ParseUUID parses a string to UUID
func ParseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}

