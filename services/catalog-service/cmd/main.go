package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ecomm-stack/catalog-service/internal/api/handlers"
	"github.com/ecomm-stack/catalog-service/internal/api/middleware"
	"github.com/ecomm-stack/catalog-service/internal/config"
	"github.com/ecomm-stack/catalog-service/internal/domain/repositories"
	"github.com/ecomm-stack/catalog-service/internal/domain/services"
	"github.com/ecomm-stack/catalog-service/internal/infrastructure/cache"
	"github.com/ecomm-stack/catalog-service/internal/infrastructure/db"
	"github.com/ecomm-stack/catalog-service/internal/infrastructure/search"
	"github.com/ecomm-stack/catalog-service/internal/otel"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	logger, err := initLogger(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	ctx := context.Background()
	otelCfg := otel.Config{
		ServiceName:    cfg.ServiceName,
		ServiceVersion: cfg.ServiceVersion,
		Environment:    cfg.Environment,
		OTLPEndpoint:   cfg.OTLPEndpoint,
	}
	if err := otel.InitTracer(ctx, otelCfg); err != nil {
		logger.Fatal("failed to initialize OpenTelemetry", zap.Error(err))
	}
	defer otel.Shutdown(ctx)

	logger.Info("starting catalog service",
		zap.String("service", cfg.ServiceName),
		zap.String("version", cfg.ServiceVersion),
	)

	database, err := db.NewPostgresConnection(cfg)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	redisClient := cache.NewRedisClient(cfg)
	defer redisClient.Close()

	esClient, err := search.NewElasticsearchClient(cfg)
	if err != nil {
		logger.Fatal("failed to connect to elasticsearch", zap.Error(err))
	}

	productRepo := repositories.NewProductRepository(database)
	categoryRepo := repositories.NewCategoryRepository(database)
	brandRepo := repositories.NewBrandRepository(database)
	mediaRepo := repositories.NewMediaRepository(database)

	productCache := cache.NewProductCache(redisClient)
	categoryCache := cache.NewCategoryCache(redisClient)
	searchIndexer := search.NewIndexer(esClient, logger)

	eventPublisher := services.NewEventPublisher(redisClient, cfg.ServiceName, logger)
	catalogService := services.NewCatalogService(productRepo, categoryRepo, brandRepo, productCache, eventPublisher, logger)
	categoryService := services.NewCategoryService(categoryRepo, categoryCache, eventPublisher, logger)
	productService := services.NewProductService(productRepo, productCache, eventPublisher, logger)
	searchService := services.NewSearchService(esClient, searchIndexer, logger)
	importService := services.NewImportService(productRepo, categoryRepo, eventPublisher, logger)
	_ = services.NewValidator(logger) // validator available for future use

	eventConsumer := services.NewEventConsumer(redisClient, catalogService, logger, "catalog-consumer")
	consumerCtx, cancelConsumer := context.WithCancel(context.Background())
	defer cancelConsumer()
	if err := eventConsumer.Start(consumerCtx); err != nil {
		logger.Fatal("failed to start event consumer", zap.Error(err))
	}

	productHandler := handlers.NewProductHandler(productService, logger)
	categoryHandler := handlers.NewCategoryHandler(categoryService, logger)
	brandHandler := handlers.NewBrandHandler(*brandRepo, logger)
	attributeHandler := handlers.NewAttributeHandler(database, logger)
	variantHandler := handlers.NewVariantHandler(productService, logger)
	mediaHandler := handlers.NewMediaHandler(*mediaRepo, logger)
	importHandler := handlers.NewImportHandler(importService, logger)
	exportHandler := handlers.NewExportHandler(productService, logger)
	searchHandler := handlers.NewSearchHandler(searchService, logger)
	healthHandler := handlers.NewHealthHandler(database, redisClient)

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(otelgin.Middleware(cfg.ServiceName))
	router.Use(middleware.Logging(logger))
	router.Use(middleware.Tracing(cfg.ServiceName))
	router.Use(middleware.Cache(redisClient, middleware.CacheConfig{
		DefaultTTL: 5 * time.Minute,
		Prefix:     "catalog:",
	}))

	router.GET("/health", healthHandler.Health)
	router.GET("/liveness", healthHandler.Liveness)
	router.GET("/readiness", healthHandler.Readiness)

	v1 := router.Group("/api/v1")
	{
		products := v1.Group("/products")
		{
			products.POST("", productHandler.Create)
			products.GET("", productHandler.List)
			products.GET("/:id", productHandler.Get)
			products.PUT("/:id", productHandler.Update)
			products.DELETE("/:id", productHandler.Delete)
			products.GET("/:id/variants", variantHandler.List)
			products.POST("/:id/variants", variantHandler.Create)
			products.GET("/:id/media", mediaHandler.List)
			products.POST("/:id/media", mediaHandler.Create)
		}
		categories := v1.Group("/categories")
		{
			categories.POST("", categoryHandler.Create)
			categories.GET("", categoryHandler.List)
			categories.GET("/:id", categoryHandler.Get)
			categories.PUT("/:id", categoryHandler.Update)
			categories.DELETE("/:id", categoryHandler.Delete)
		}
		brands := v1.Group("/brands")
		{
			brands.POST("", brandHandler.Create)
			brands.GET("", brandHandler.List)
			brands.GET("/:id", brandHandler.Get)
		}
		attributes := v1.Group("/attributes")
		{
			attributes.POST("", attributeHandler.Create)
			attributes.GET("", attributeHandler.List)
		}
		v1.POST("/import", importHandler.Import)
		v1.GET("/export", exportHandler.Export)
		v1.GET("/search", searchHandler.Search)
	}

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	srv := &http.Server{Addr: addr, Handler: router}
	go func() {
		logger.Info("starting HTTP server", zap.String("address", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}
	logger.Info("server exited")
}

func initLogger(cfg *config.Config) (*zap.Logger, error) {
	if cfg.Environment == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

