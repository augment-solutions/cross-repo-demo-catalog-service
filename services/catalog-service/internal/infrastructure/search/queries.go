package search

// SearchQueries contains Elasticsearch query templates
var SearchQueries = struct {
	FullTextSearch   string
	CategoryFilter   string
	BrandFilter      string
	PriceRangeFilter string
	FeaturedProducts string
	SuggestQuery     string
}{
	FullTextSearch: `{
		"query": {
			"bool": {
				"should": [
					{
						"multi_match": {
							"query": "%s",
							"fields": ["name^3", "description", "sku", "tags"],
							"type": "best_fields",
							"fuzziness": "AUTO"
						}
					},
					{
						"match_phrase_prefix": {
							"name": {
								"query": "%s",
								"boost": 2
							}
						}
					}
				],
				"minimum_should_match": 1,
				"filter": [
					{"term": {"status": "active"}}
				]
			}
		}
	}`,

	CategoryFilter: `{
		"query": {
			"bool": {
				"filter": [
					{"term": {"category_id": "%s"}},
					{"term": {"status": "active"}}
				]
			}
		}
	}`,

	BrandFilter: `{
		"query": {
			"bool": {
				"filter": [
					{"term": {"brand_id": "%s"}},
					{"term": {"status": "active"}}
				]
			}
		}
	}`,

	PriceRangeFilter: `{
		"query": {
			"bool": {
				"filter": [
					{"range": {"price": {"gte": %f, "lte": %f}}},
					{"term": {"status": "active"}}
				]
			}
		}
	}`,

	FeaturedProducts: `{
		"query": {
			"bool": {
				"filter": [
					{"term": {"is_featured": true}},
					{"term": {"status": "active"}}
				]
			}
		},
		"sort": [
			{"updated_at": "desc"}
		],
		"size": %d
	}`,

	SuggestQuery: `{
		"suggest": {
			"product-suggest": {
				"prefix": "%s",
				"completion": {
					"field": "name.suggest",
					"size": 10,
					"skip_duplicates": true,
					"fuzzy": {
						"fuzziness": "AUTO"
					}
				}
			}
		}
	}`,
}

// Aggregations for faceted search
const FacetAggregations = `{
	"aggs": {
		"categories": {
			"terms": {
				"field": "category_name",
				"size": 20
			}
		},
		"brands": {
			"terms": {
				"field": "brand_name",
				"size": 20
			}
		},
		"price_ranges": {
			"range": {
				"field": "price",
				"ranges": [
					{"to": 25},
					{"from": 25, "to": 50},
					{"from": 50, "to": 100},
					{"from": 100, "to": 250},
					{"from": 250}
				]
			}
		}
	}
}`

