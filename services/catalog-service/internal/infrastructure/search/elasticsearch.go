package search

import (
	"github.com/ecomm-stack/catalog-service/internal/config"
	"github.com/elastic/go-elasticsearch/v8"
)

// NewElasticsearchClient creates a new Elasticsearch client
func NewElasticsearchClient(cfg *config.Config) (*elasticsearch.Client, error) {
	esCfg := elasticsearch.Config{
		Addresses: []string{cfg.ElasticsearchURL},
	}

	return elasticsearch.NewClient(esCfg)
}

// ProductIndex is the Elasticsearch index name for products
const ProductIndex = "catalog-products"

// ProductMapping defines the Elasticsearch mapping for products
const ProductMapping = `
{
	"settings": {
		"number_of_shards": 1,
		"number_of_replicas": 0,
		"analysis": {
			"analyzer": {
				"product_analyzer": {
					"type": "custom",
					"tokenizer": "standard",
					"filter": ["lowercase", "asciifolding", "edge_ngram_filter"]
				}
			},
			"filter": {
				"edge_ngram_filter": {
					"type": "edge_ngram",
					"min_gram": 2,
					"max_gram": 20
				}
			}
		}
	},
	"mappings": {
		"properties": {
			"id": { "type": "keyword" },
			"sku": { "type": "keyword" },
			"name": { 
				"type": "text",
				"analyzer": "product_analyzer",
				"fields": {
					"keyword": { "type": "keyword" }
				}
			},
			"slug": { "type": "keyword" },
			"description": { "type": "text" },
			"short_description": { "type": "text" },
			"category_id": { "type": "keyword" },
			"category_name": { "type": "keyword" },
			"category_path": { "type": "keyword" },
			"brand_id": { "type": "keyword" },
			"brand_name": { "type": "keyword" },
			"status": { "type": "keyword" },
			"is_featured": { "type": "boolean" },
			"is_new": { "type": "boolean" },
			"price": { "type": "float" },
			"compare_at_price": { "type": "float" },
			"tags": { "type": "keyword" },
			"attributes": {
				"type": "nested",
				"properties": {
					"name": { "type": "keyword" },
					"value": { "type": "keyword" }
				}
			},
			"image_url": { "type": "keyword" },
			"created_at": { "type": "date" },
			"updated_at": { "type": "date" }
		}
	}
}
`

