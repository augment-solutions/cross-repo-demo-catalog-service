package otel

import (
	"go.opentelemetry.io/otel/metric"
)

var (
	ProductsCreated   metric.Int64Counter
	ProductsUpdated   metric.Int64Counter
	ProductsDeleted   metric.Int64Counter
	ProductsViewed    metric.Int64Counter
	CategoriesCreated metric.Int64Counter
	SearchQueries     metric.Int64Counter
	CacheHits         metric.Int64Counter
	CacheMisses       metric.Int64Counter
	ImportOperations  metric.Int64Counter
	ExportOperations  metric.Int64Counter
	SearchLatency     metric.Float64Histogram
	DBQueryLatency    metric.Float64Histogram
)

func initMetrics() error {
	var err error

	ProductsCreated, err = meter.Int64Counter("catalog.products.created",
		metric.WithDescription("Number of products created"),
		metric.WithUnit("1"))
	if err != nil {
		return err
	}

	ProductsUpdated, err = meter.Int64Counter("catalog.products.updated",
		metric.WithDescription("Number of products updated"),
		metric.WithUnit("1"))
	if err != nil {
		return err
	}

	ProductsDeleted, err = meter.Int64Counter("catalog.products.deleted",
		metric.WithDescription("Number of products deleted"),
		metric.WithUnit("1"))
	if err != nil {
		return err
	}

	ProductsViewed, err = meter.Int64Counter("catalog.products.viewed",
		metric.WithDescription("Number of product views"),
		metric.WithUnit("1"))
	if err != nil {
		return err
	}

	CategoriesCreated, err = meter.Int64Counter("catalog.categories.created",
		metric.WithDescription("Number of categories created"),
		metric.WithUnit("1"))
	if err != nil {
		return err
	}

	SearchQueries, err = meter.Int64Counter("catalog.search.queries",
		metric.WithDescription("Number of search queries"),
		metric.WithUnit("1"))
	if err != nil {
		return err
	}

	CacheHits, err = meter.Int64Counter("catalog.cache.hits",
		metric.WithDescription("Number of cache hits"),
		metric.WithUnit("1"))
	if err != nil {
		return err
	}

	CacheMisses, err = meter.Int64Counter("catalog.cache.misses",
		metric.WithDescription("Number of cache misses"),
		metric.WithUnit("1"))
	if err != nil {
		return err
	}

	ImportOperations, err = meter.Int64Counter("catalog.import.operations",
		metric.WithDescription("Number of import operations"),
		metric.WithUnit("1"))
	if err != nil {
		return err
	}

	ExportOperations, err = meter.Int64Counter("catalog.export.operations",
		metric.WithDescription("Number of export operations"),
		metric.WithUnit("1"))
	if err != nil {
		return err
	}

	SearchLatency, err = meter.Float64Histogram("catalog.search.latency",
		metric.WithDescription("Search query latency"),
		metric.WithUnit("ms"))
	if err != nil {
		return err
	}

	DBQueryLatency, err = meter.Float64Histogram("catalog.db.query_latency",
		metric.WithDescription("Database query latency"),
		metric.WithUnit("ms"))
	if err != nil {
		return err
	}

	return nil
}

