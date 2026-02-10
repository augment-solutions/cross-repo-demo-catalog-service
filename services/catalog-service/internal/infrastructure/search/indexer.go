package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/ecomm-stack/catalog-service/internal/otel"
	"github.com/elastic/go-elasticsearch/v8"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// Indexer handles Elasticsearch indexing operations
type Indexer struct {
	client *elasticsearch.Client
	logger *zap.Logger
}

// NewIndexer creates a new Indexer
func NewIndexer(client *elasticsearch.Client, logger *zap.Logger) *Indexer {
	return &Indexer{
		client: client,
		logger: logger,
	}
}

// ProductDocument represents a product in Elasticsearch
type ProductDocument struct {
	ID               string            `json:"id"`
	SKU              string            `json:"sku"`
	Name             string            `json:"name"`
	Slug             string            `json:"slug"`
	Description      string            `json:"description"`
	ShortDescription string            `json:"short_description"`
	CategoryID       string            `json:"category_id,omitempty"`
	CategoryName     string            `json:"category_name,omitempty"`
	BrandID          string            `json:"brand_id,omitempty"`
	BrandName        string            `json:"brand_name,omitempty"`
	Status           string            `json:"status"`
	IsFeatured       bool              `json:"is_featured"`
	IsNew            bool              `json:"is_new"`
	Price            float64           `json:"price"`
	CompareAtPrice   float64           `json:"compare_at_price"`
	Tags             []string          `json:"tags"`
	ImageURL         string            `json:"image_url,omitempty"`
	Attributes       []AttributeDoc    `json:"attributes,omitempty"`
}

// AttributeDoc represents an attribute in Elasticsearch
type AttributeDoc struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// IndexProduct indexes a single product
func (i *Indexer) IndexProduct(ctx context.Context, product *models.Product) error {
	ctx, span := otel.StartSpan(ctx, "Indexer.IndexProduct")
	defer span.End()
	span.SetAttributes(attribute.String("product_id", product.ID.String()))

	doc := i.toDocument(product)
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	res, err := i.client.Index(
		ProductIndex,
		bytes.NewReader(data),
		i.client.Index.WithContext(ctx),
		i.client.Index.WithDocumentID(product.ID.String()),
		i.client.Index.WithRefresh("true"),
	)
	if err != nil {
		i.logger.Error("failed to index product", zap.Error(err))
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing product: %s", res.String())
	}

	i.logger.Debug("product indexed", zap.String("id", product.ID.String()))
	return nil
}

// DeleteProduct removes a product from the index
func (i *Indexer) DeleteProduct(ctx context.Context, productID string) error {
	ctx, span := otel.StartSpan(ctx, "Indexer.DeleteProduct")
	defer span.End()

	res, err := i.client.Delete(
		ProductIndex,
		productID,
		i.client.Delete.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

// BulkIndex indexes multiple products
func (i *Indexer) BulkIndex(ctx context.Context, products []*models.Product) error {
	ctx, span := otel.StartSpan(ctx, "Indexer.BulkIndex")
	defer span.End()
	span.SetAttributes(attribute.Int("count", len(products)))

	var buf bytes.Buffer
	for _, product := range products {
		meta := fmt.Sprintf(`{"index":{"_index":"%s","_id":"%s"}}%s`, ProductIndex, product.ID.String(), "\n")
		buf.WriteString(meta)

		doc := i.toDocument(product)
		data, err := json.Marshal(doc)
		if err != nil {
			return err
		}
		buf.Write(data)
		buf.WriteString("\n")
	}

	res, err := i.client.Bulk(bytes.NewReader(buf.Bytes()), i.client.Bulk.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	i.logger.Info("bulk indexed products", zap.Int("count", len(products)))
	return nil
}

func (i *Indexer) toDocument(product *models.Product) *ProductDocument {
	doc := &ProductDocument{
		ID:               product.ID.String(),
		SKU:              product.SKU,
		Name:             product.Name,
		Slug:             product.Slug,
		Description:      product.Description,
		ShortDescription: product.ShortDescription,
		Status:           string(product.Status),
		IsFeatured:       product.IsFeatured,
		IsNew:            product.IsNew,
		Tags:             product.Tags,
	}

	if product.CategoryID != nil {
		doc.CategoryID = product.CategoryID.String()
	}
	if product.BrandID != nil {
		doc.BrandID = product.BrandID.String()
	}
	if len(product.Variants) > 0 {
		doc.Price = product.Variants[0].Price
		doc.CompareAtPrice = product.Variants[0].CompareAtPrice
	}

	return doc
}

