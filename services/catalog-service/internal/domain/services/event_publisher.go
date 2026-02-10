package services

import (
	"context"
	"encoding/json"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/ecomm-stack/catalog-service/internal/otel"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

const (
	catalogEventsStream = "catalog:events"
)

// EventPublisher publishes domain events to Redis Streams
type EventPublisher struct {
	redisClient *redis.Client
	serviceName string
	logger      *zap.Logger
}

// NewEventPublisher creates a new EventPublisher
func NewEventPublisher(redisClient *redis.Client, serviceName string, logger *zap.Logger) *EventPublisher {
	return &EventPublisher{
		redisClient: redisClient,
		serviceName: serviceName,
		logger:      logger,
	}
}

// Publish publishes an event to Redis Streams
func (p *EventPublisher) Publish(ctx context.Context, event *models.Event) error {
	ctx, span := otel.StartSpan(ctx, "EventPublisher.Publish")
	defer span.End()
	span.SetAttributes(
		attribute.String("event_type", string(event.Type)),
		attribute.String("event_id", event.ID.String()),
	)

	// Add trace context to event metadata
	event.Metadata.Source = p.serviceName
	if span.SpanContext().HasTraceID() {
		event.Metadata.TraceID = span.SpanContext().TraceID().String()
		event.Metadata.SpanID = span.SpanContext().SpanID().String()
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("failed to marshal event", zap.Error(err))
		return err
	}

	// Publish to Redis Stream
	args := &redis.XAddArgs{
		Stream: catalogEventsStream,
		Values: map[string]interface{}{
			"event_id":   event.ID.String(),
			"event_type": string(event.Type),
			"data":       string(eventData),
			"timestamp":  event.Timestamp.UnixMilli(),
		},
	}

	_, err = p.redisClient.XAdd(ctx, args).Result()
	if err != nil {
		p.logger.Error("failed to publish event to stream",
			zap.Error(err),
			zap.String("event_type", string(event.Type)),
		)
		return err
	}

	p.logger.Debug("event published",
		zap.String("event_id", event.ID.String()),
		zap.String("event_type", string(event.Type)),
	)

	return nil
}

// PublishProductCreated publishes a product created event
func (p *EventPublisher) PublishProductCreated(ctx context.Context, product *models.Product) error {
	event, err := models.NewEvent(models.EventTypeProductCreated, product.ID, models.ProductCreatedData{
		ProductID:  product.ID,
		SKU:        product.SKU,
		Name:       product.Name,
		CategoryID: product.CategoryID,
		BrandID:    product.BrandID,
	})
	if err != nil {
		return err
	}
	return p.Publish(ctx, event)
}

// PublishProductUpdated publishes a product updated event
func (p *EventPublisher) PublishProductUpdated(ctx context.Context, product *models.Product, changedFields map[string]interface{}) error {
	event, err := models.NewEvent(models.EventTypeProductUpdated, product.ID, models.ProductUpdatedData{
		ProductID:     product.ID,
		ChangedFields: changedFields,
	})
	if err != nil {
		return err
	}
	return p.Publish(ctx, event)
}

// PublishProductDeleted publishes a product deleted event
func (p *EventPublisher) PublishProductDeleted(ctx context.Context, productID, sku string) error {
	event, err := models.NewEvent(models.EventTypeProductDeleted, models.ParseUUID(productID), models.ProductDeletedData{
		ProductID: models.ParseUUID(productID),
		SKU:       sku,
	})
	if err != nil {
		return err
	}
	return p.Publish(ctx, event)
}

// PublishCategoryCreated publishes a category created event
func (p *EventPublisher) PublishCategoryCreated(ctx context.Context, category *models.Category) error {
	event, err := models.NewEvent(models.EventTypeCategoryCreated, category.ID, models.CategoryCreatedData{
		CategoryID: category.ID,
		Name:       category.Name,
		Slug:       category.Slug,
		ParentID:   category.ParentID,
	})
	if err != nil {
		return err
	}
	return p.Publish(ctx, event)
}

