package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/ecomm-stack/catalog-service/internal/otel"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

const (
	inventoryEventsStream = "inventory:events"
	priceEventsStream     = "pricing:events"
	consumerGroup         = "catalog-service"
)

// EventConsumer consumes events from Redis Streams
type EventConsumer struct {
	redisClient    *redis.Client
	catalogService *CatalogService
	logger         *zap.Logger
	consumerName   string
}

// NewEventConsumer creates a new EventConsumer
func NewEventConsumer(
	redisClient *redis.Client,
	catalogService *CatalogService,
	logger *zap.Logger,
	consumerName string,
) *EventConsumer {
	return &EventConsumer{
		redisClient:    redisClient,
		catalogService: catalogService,
		logger:         logger,
		consumerName:   consumerName,
	}
}

// Start starts consuming events
func (c *EventConsumer) Start(ctx context.Context) error {
	// Create consumer groups if they don't exist
	streams := []string{inventoryEventsStream, priceEventsStream}
	for _, stream := range streams {
		c.redisClient.XGroupCreateMkStream(ctx, stream, consumerGroup, "0").Err()
	}

	go c.consumeLoop(ctx)
	c.logger.Info("event consumer started", zap.String("consumer", c.consumerName))
	return nil
}

// consumeLoop continuously reads events from streams
func (c *EventConsumer) consumeLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("event consumer stopped")
			return
		default:
			c.readAndProcessEvents(ctx)
		}
	}
}

// readAndProcessEvents reads and processes events from all streams
func (c *EventConsumer) readAndProcessEvents(ctx context.Context) {
	streams := []string{inventoryEventsStream, priceEventsStream, ">", ">"}

	result, err := c.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    consumerGroup,
		Consumer: c.consumerName,
		Streams:  streams,
		Count:    10,
		Block:    5 * time.Second,
	}).Result()

	if err != nil && err != redis.Nil {
		c.logger.Error("failed to read from stream", zap.Error(err))
		time.Sleep(time.Second)
		return
	}

	for _, xStream := range result {
		for _, message := range xStream.Messages {
			c.processMessage(ctx, xStream.Stream, message)
		}
	}
}

// processMessage processes a single message
func (c *EventConsumer) processMessage(ctx context.Context, stream string, message redis.XMessage) {
	ctx, span := otel.StartSpan(ctx, "EventConsumer.ProcessMessage")
	defer span.End()
	span.SetAttributes(
		attribute.String("stream", stream),
		attribute.String("message_id", message.ID),
	)

	eventData, ok := message.Values["data"].(string)
	if !ok {
		c.logger.Error("invalid message format", zap.String("id", message.ID))
		c.ackMessage(ctx, stream, message.ID)
		return
	}

	var event models.Event
	if err := json.Unmarshal([]byte(eventData), &event); err != nil {
		c.logger.Error("failed to unmarshal event", zap.Error(err))
		c.ackMessage(ctx, stream, message.ID)
		return
	}

	// Handle based on event type
	switch event.Type {
	case models.EventTypeInventoryUpdated:
		c.handleInventoryUpdated(ctx, &event)
	case models.EventTypePriceUpdated:
		c.handlePriceUpdated(ctx, &event)
	default:
		c.logger.Debug("unhandled event type", zap.String("type", string(event.Type)))
	}

	c.ackMessage(ctx, stream, message.ID)
}

// handleInventoryUpdated handles inventory update events
func (c *EventConsumer) handleInventoryUpdated(ctx context.Context, event *models.Event) {
	var data models.InventoryUpdatedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		c.logger.Error("failed to unmarshal inventory data", zap.Error(err))
		return
	}

	c.logger.Info("inventory updated",
		zap.String("sku", data.SKU),
		zap.Int("new_quantity", data.NewQuantity),
	)
	// Refresh cache for affected products
}

// handlePriceUpdated handles price update events
func (c *EventConsumer) handlePriceUpdated(ctx context.Context, event *models.Event) {
	c.logger.Info("price updated", zap.String("aggregate_id", event.AggregateID.String()))
	// Refresh cache for affected products
}

// ackMessage acknowledges a message
func (c *EventConsumer) ackMessage(ctx context.Context, stream, messageID string) {
	c.redisClient.XAck(ctx, stream, consumerGroup, messageID)
}

