package services

import (
	"context"
	"time"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/ecomm-stack/catalog-service/internal/infrastructure/search"
	"github.com/ecomm-stack/catalog-service/internal/otel"
	"github.com/elastic/go-elasticsearch/v8"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// SearchResult represents a search result
type SearchResult struct {
	Products    []*models.ProductDetailsResponse `json:"products"`
	Total       int64                            `json:"total"`
	Took        int64                            `json:"took_ms"`
	Facets      map[string][]Facet               `json:"facets,omitempty"`
	Suggestions []string                         `json:"suggestions,omitempty"`
}

// Facet represents a search facet
type Facet struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

// SearchOptions represents search options from the handler
type SearchOptions struct {
	Query      string   `json:"query"`
	CategoryID string   `json:"category_id"`
	BrandID    string   `json:"brand_id"`
	MinPrice   float64  `json:"min_price"`
	MaxPrice   float64  `json:"max_price"`
	Tags       []string `json:"tags"`
	SortBy     string   `json:"sort_by"`
	SortOrder  string   `json:"sort_order"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
}

// SearchService handles search operations
type SearchService struct {
	esClient *elasticsearch.Client
	indexer  *search.Indexer
	logger   *zap.Logger
}

// NewSearchService creates a new SearchService
func NewSearchService(esClient *elasticsearch.Client, indexer *search.Indexer, logger *zap.Logger) *SearchService {
	return &SearchService{
		esClient: esClient,
		indexer:  indexer,
		logger:   logger,
	}
}

// Search performs a product search
func (s *SearchService) Search(ctx context.Context, opts *SearchOptions) (*SearchResult, error) {
	ctx, span := otel.StartSpan(ctx, "SearchService.Search")
	defer span.End()
	span.SetAttributes(attribute.String("query", opts.Query))

	start := time.Now()

	// Build search query
	query := s.buildSearchQueryFromOpts(opts)

	// Execute search
	products, total, err := s.executeSearch(ctx, query, opts.Page, opts.PageSize)
	if err != nil {
		s.logger.Error("search failed", zap.Error(err), zap.String("query", opts.Query))
		return nil, err
	}

	took := time.Since(start).Milliseconds()
	otel.SearchQueries.Add(ctx, 1)
	otel.SearchLatency.Record(ctx, float64(took))

	return &SearchResult{
		Products: models.NewProductDetailsResponses(products),
		Total:    total,
		Took:     took,
	}, nil
}

// buildSearchQueryFromOpts builds an Elasticsearch query from SearchOptions
func (s *SearchService) buildSearchQueryFromOpts(opts *SearchOptions) map[string]interface{} {
	must := []map[string]interface{}{}
	filter := []map[string]interface{}{}

	if opts.Query != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  opts.Query,
				"fields": []string{"name^3", "description", "sku", "tags"},
				"type":   "best_fields",
			},
		})
	}

	if opts.CategoryID != "" {
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{"category_id": opts.CategoryID},
		})
	}

	if opts.BrandID != "" {
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{"brand_id": opts.BrandID},
		})
	}

	if opts.MinPrice > 0 {
		filter = append(filter, map[string]interface{}{
			"range": map[string]interface{}{"price": map[string]interface{}{"gte": opts.MinPrice}},
		})
	}

	if opts.MaxPrice > 0 {
		filter = append(filter, map[string]interface{}{
			"range": map[string]interface{}{"price": map[string]interface{}{"lte": opts.MaxPrice}},
		})
	}

	// Always filter for active products
	filter = append(filter, map[string]interface{}{
		"term": map[string]interface{}{"status": "active"},
	})

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":   must,
				"filter": filter,
			},
		},
	}
}

// buildSearchQuery builds an Elasticsearch query
func (s *SearchService) buildSearchQuery(req *models.SearchRequest) map[string]interface{} {
	must := []map[string]interface{}{}
	filter := []map[string]interface{}{}

	if req.Query != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  req.Query,
				"fields": []string{"name^3", "description", "sku", "tags"},
				"type":   "best_fields",
			},
		})
	}

	if req.CategoryID != nil {
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{"category_id": req.CategoryID.String()},
		})
	}

	if req.BrandID != nil {
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{"brand_id": req.BrandID.String()},
		})
	}

	if req.MinPrice != nil {
		filter = append(filter, map[string]interface{}{
			"range": map[string]interface{}{"price": map[string]interface{}{"gte": *req.MinPrice}},
		})
	}

	if req.MaxPrice != nil {
		filter = append(filter, map[string]interface{}{
			"range": map[string]interface{}{"price": map[string]interface{}{"lte": *req.MaxPrice}},
		})
	}

	// Always filter for active products
	filter = append(filter, map[string]interface{}{
		"term": map[string]interface{}{"status": "active"},
	})

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":   must,
				"filter": filter,
			},
		},
	}
}

// executeSearch executes the search query
func (s *SearchService) executeSearch(ctx context.Context, query map[string]interface{}, page, pageSize int) ([]*models.Product, int64, error) {
	// This would integrate with Elasticsearch
	// For now, return empty results as placeholder
	return []*models.Product{}, 0, nil
}

// IndexProduct indexes a product
func (s *SearchService) IndexProduct(ctx context.Context, product *models.Product) error {
	return s.indexer.IndexProduct(ctx, product)
}

// DeleteProductFromIndex removes a product from the index
func (s *SearchService) DeleteProductFromIndex(ctx context.Context, productID string) error {
	return s.indexer.DeleteProduct(ctx, productID)
}

// ReindexAll reindexes all products
func (s *SearchService) ReindexAll(ctx context.Context) error {
	// In a real implementation, this would fetch all products and reindex them
	// For now, return nil as placeholder
	return nil
}

// Suggest provides search suggestions based on a query prefix
func (s *SearchService) Suggest(ctx context.Context, query string, limit int) ([]string, error) {
	ctx, span := otel.StartSpan(ctx, "SearchService.Suggest")
	defer span.End()
	span.SetAttributes(attribute.String("query", query), attribute.Int("limit", limit))

	// In a real implementation, this would query Elasticsearch for suggestions
	// For now, return empty suggestions as placeholder
	return []string{}, nil
}

// GetFacets returns available facets for filtering
func (s *SearchService) GetFacets(ctx context.Context, categoryID string) (map[string][]Facet, error) {
	ctx, span := otel.StartSpan(ctx, "SearchService.GetFacets")
	defer span.End()
	span.SetAttributes(attribute.String("category_id", categoryID))

	// In a real implementation, this would query Elasticsearch for facets
	// For now, return empty facets as placeholder
	return map[string][]Facet{}, nil
}
