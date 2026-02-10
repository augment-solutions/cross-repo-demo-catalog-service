package db

import (
	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	ID   string
	Up   func(*gorm.DB) error
	Down func(*gorm.DB) error
}

// Migrations holds all migrations
var Migrations = []Migration{
	{
		ID: "001_create_products_table",
		Up: func(db *gorm.DB) error {
			return db.Exec(`
				CREATE TABLE IF NOT EXISTS products (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					sku VARCHAR(64) UNIQUE NOT NULL,
					name VARCHAR(255) NOT NULL,
					slug VARCHAR(255) UNIQUE NOT NULL,
					description TEXT,
					short_description TEXT,
					category_id UUID REFERENCES categories(id),
					brand_id UUID REFERENCES brands(id),
					status VARCHAR(20) DEFAULT 'draft',
					is_featured BOOLEAN DEFAULT FALSE,
					is_new BOOLEAN DEFAULT FALSE,
					weight DECIMAL(10,2),
					length DECIMAL(10,2),
					width DECIMAL(10,2),
					height DECIMAL(10,2),
					meta_title VARCHAR(255),
					meta_description TEXT,
					meta_keywords TEXT,
					tags TEXT[],
					created_at TIMESTAMP DEFAULT NOW(),
					updated_at TIMESTAMP DEFAULT NOW(),
					deleted_at TIMESTAMP
				)
			`).Error
		},
		Down: func(db *gorm.DB) error {
			return db.Exec("DROP TABLE IF EXISTS products").Error
		},
	},
	{
		ID: "002_create_categories_table",
		Up: func(db *gorm.DB) error {
			return db.Exec(`
				CREATE TABLE IF NOT EXISTS categories (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					name VARCHAR(255) NOT NULL,
					slug VARCHAR(255) UNIQUE NOT NULL,
					description TEXT,
					parent_id UUID REFERENCES categories(id),
					level INTEGER DEFAULT 0,
					path TEXT,
					image_url TEXT,
					icon_url TEXT,
					is_active BOOLEAN DEFAULT TRUE,
					is_featured BOOLEAN DEFAULT FALSE,
					sort_order INTEGER DEFAULT 0,
					meta_title VARCHAR(255),
					meta_description TEXT,
					meta_keywords TEXT,
					created_at TIMESTAMP DEFAULT NOW(),
					updated_at TIMESTAMP DEFAULT NOW(),
					deleted_at TIMESTAMP
				)
			`).Error
		},
		Down: func(db *gorm.DB) error {
			return db.Exec("DROP TABLE IF EXISTS categories").Error
		},
	},
	{
		ID: "003_create_brands_table",
		Up: func(db *gorm.DB) error {
			return db.Exec(`
				CREATE TABLE IF NOT EXISTS brands (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					name VARCHAR(255) UNIQUE NOT NULL,
					slug VARCHAR(255) UNIQUE NOT NULL,
					description TEXT,
					logo_url TEXT,
					website_url TEXT,
					is_active BOOLEAN DEFAULT TRUE,
					is_featured BOOLEAN DEFAULT FALSE,
					sort_order INTEGER DEFAULT 0,
					meta_title VARCHAR(255),
					meta_description TEXT,
					created_at TIMESTAMP DEFAULT NOW(),
					updated_at TIMESTAMP DEFAULT NOW(),
					deleted_at TIMESTAMP
				)
			`).Error
		},
		Down: func(db *gorm.DB) error {
			return db.Exec("DROP TABLE IF EXISTS brands").Error
		},
	},
	{
		ID: "004_create_variants_table",
		Up: func(db *gorm.DB) error {
			return db.Exec(`
				CREATE TABLE IF NOT EXISTS variants (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					product_id UUID NOT NULL REFERENCES products(id),
					sku VARCHAR(64) UNIQUE NOT NULL,
					name VARCHAR(255),
					price DECIMAL(10,2) NOT NULL,
					compare_at_price DECIMAL(10,2),
					cost_price DECIMAL(10,2),
					weight DECIMAL(10,2),
					barcode VARCHAR(64),
					is_default BOOLEAN DEFAULT FALSE,
					is_active BOOLEAN DEFAULT TRUE,
					sort_order INTEGER DEFAULT 0,
					image_url TEXT,
					created_at TIMESTAMP DEFAULT NOW(),
					updated_at TIMESTAMP DEFAULT NOW(),
					deleted_at TIMESTAMP
				)
			`).Error
		},
		Down: func(db *gorm.DB) error {
			return db.Exec("DROP TABLE IF EXISTS variants").Error
		},
	},
}

// RunMigrations executes all pending migrations
func RunMigrations(db *gorm.DB) error {
	for _, m := range Migrations {
		if err := m.Up(db); err != nil {
			return err
		}
	}
	return nil
}

