package api

import (
	"github.com/ecomm-stack/catalog-service/internal/api/handlers"
	"github.com/ecomm-stack/catalog-service/internal/api/middleware"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RouterConfig holds router configuration
type RouterConfig struct {
	ProductHandler   *handlers.ProductHandler
	CategoryHandler  *handlers.CategoryHandler
	BrandHandler     *handlers.BrandHandler
	VariantHandler   *handlers.VariantHandler
	MediaHandler     *handlers.MediaHandler
	SearchHandler    *handlers.SearchHandler
	ImportHandler    *handlers.ImportHandler
	ExportHandler    *handlers.ExportHandler
	AttributeHandler *handlers.AttributeHandler
	HealthHandler    *handlers.HealthHandler
	Logger           *zap.Logger
	DB               *gorm.DB
	Redis            *redis.Client
}

// NewRouter creates a new Gin router with all routes configured
func NewRouter(cfg *RouterConfig) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(middleware.Recovery(cfg.Logger))
	router.Use(middleware.Tracing("catalog-service"))
	router.Use(middleware.Logging(cfg.Logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.InjectTraceID())

	// Health and metrics endpoints
	router.GET("/health", cfg.HealthHandler.Health)
	router.GET("/health/live", cfg.HealthHandler.Liveness)
	router.GET("/health/ready", cfg.HealthHandler.Readiness)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Products
		products := v1.Group("/products")
		{
			products.GET("", cfg.ProductHandler.List)
			products.POST("", cfg.ProductHandler.Create)
			products.GET("/:id", cfg.ProductHandler.Get)
			products.PUT("/:id", cfg.ProductHandler.Update)
			products.DELETE("/:id", cfg.ProductHandler.Delete)
			products.GET("/sku/:sku", cfg.ProductHandler.GetBySKU)

			// Product variants
			products.GET("/:productId/variants", cfg.VariantHandler.List)
			products.POST("/:productId/variants", cfg.VariantHandler.Create)

			// Product media
			products.GET("/:productId/media", cfg.MediaHandler.List)
			products.POST("/:productId/media", cfg.MediaHandler.Create)
			products.PUT("/:productId/media/:id/primary", cfg.MediaHandler.SetPrimary)
			products.PUT("/:productId/media/reorder", cfg.MediaHandler.Reorder)
		}

		// Variants
		variants := v1.Group("/variants")
		{
			variants.GET("/:id", cfg.VariantHandler.Get)
			variants.PUT("/:id", cfg.VariantHandler.Update)
			variants.DELETE("/:id", cfg.VariantHandler.Delete)
		}

		// Media
		media := v1.Group("/media")
		{
			media.GET("/:id", cfg.MediaHandler.Get)
			media.PUT("/:id", cfg.MediaHandler.Update)
			media.DELETE("/:id", cfg.MediaHandler.Delete)
		}

		// Categories
		categories := v1.Group("/categories")
		{
			categories.GET("", cfg.CategoryHandler.List)
			categories.POST("", cfg.CategoryHandler.Create)
			categories.GET("/tree", cfg.CategoryHandler.GetTree)
			categories.GET("/:id", cfg.CategoryHandler.Get)
			categories.PUT("/:id", cfg.CategoryHandler.Update)
			categories.DELETE("/:id", cfg.CategoryHandler.Delete)
		}

		// Brands
		brands := v1.Group("/brands")
		{
			brands.GET("", cfg.BrandHandler.List)
			brands.POST("", cfg.BrandHandler.Create)
			brands.GET("/:id", cfg.BrandHandler.Get)
			brands.PUT("/:id", cfg.BrandHandler.Update)
			brands.DELETE("/:id", cfg.BrandHandler.Delete)
		}

		// Attributes
		attributes := v1.Group("/attributes")
		{
			attributes.GET("", cfg.AttributeHandler.List)
			attributes.POST("", cfg.AttributeHandler.Create)
			attributes.GET("/:id", cfg.AttributeHandler.Get)
			attributes.PUT("/:id", cfg.AttributeHandler.Update)
			attributes.DELETE("/:id", cfg.AttributeHandler.Delete)
			attributes.POST("/:id/options", cfg.AttributeHandler.AddOption)
		}

		// Search
		search := v1.Group("/search")
		{
			search.GET("", cfg.SearchHandler.Search)
			search.GET("/suggest", cfg.SearchHandler.Suggest)
			search.GET("/facets", cfg.SearchHandler.Facets)
			search.POST("/reindex", cfg.SearchHandler.Reindex)
		}

		// Import/Export
		v1.POST("/import", cfg.ImportHandler.Import)
		v1.POST("/import/upload", cfg.ImportHandler.UploadImport)
		v1.GET("/import/:jobId", cfg.ImportHandler.GetImportStatus)
		v1.DELETE("/import/:jobId", cfg.ImportHandler.CancelImport)
		v1.GET("/import/template", cfg.ImportHandler.GetImportTemplate)

		v1.GET("/export", cfg.ExportHandler.Export)
		v1.POST("/export/async", cfg.ExportHandler.ExportAsync)
		v1.GET("/export/:jobId", cfg.ExportHandler.GetExportStatus)
		v1.GET("/export/:jobId/download", cfg.ExportHandler.DownloadExport)
	}

	return router
}

