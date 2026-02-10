package db

import (
	"github.com/ecomm-stack/catalog-service/internal/config"
	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresConnection creates a new PostgreSQL connection
func NewPostgresConnection(cfg *config.Config) (*gorm.DB, error) {
	gormConfig := &gorm.Config{}

	if cfg.Environment == "production" {
		gormConfig.Logger = logger.Default.LogMode(logger.Silent)
	} else {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), gormConfig)
	if err != nil {
		return nil, err
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}

// runMigrations runs database migrations
func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Product{},
		&models.Category{},
		&models.Brand{},
		&models.Attribute{},
		&models.AttributeOption{},
		&models.Variant{},
		&models.VariantOption{},
		&models.Media{},
		&models.Price{},
		&models.PriceHistory{},
		&models.ProductAttribute{},
	)
}

// GetDB returns the underlying sql.DB
func GetDB(db *gorm.DB) (*gorm.DB, error) {
	return db, nil
}

