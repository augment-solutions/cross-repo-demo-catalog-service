package db

import (
	"gorm.io/gorm"
)

// IndexDefinition represents a database index
type IndexDefinition struct {
	Name    string
	Table   string
	Columns []string
	Unique  bool
}

// Indexes defines all indexes for the catalog service
var Indexes = []IndexDefinition{
	// Product indexes
	{Name: "idx_products_sku", Table: "products", Columns: []string{"sku"}, Unique: true},
	{Name: "idx_products_slug", Table: "products", Columns: []string{"slug"}, Unique: true},
	{Name: "idx_products_category_id", Table: "products", Columns: []string{"category_id"}},
	{Name: "idx_products_brand_id", Table: "products", Columns: []string{"brand_id"}},
	{Name: "idx_products_status", Table: "products", Columns: []string{"status"}},
	{Name: "idx_products_is_featured", Table: "products", Columns: []string{"is_featured"}},
	{Name: "idx_products_created_at", Table: "products", Columns: []string{"created_at"}},
	{Name: "idx_products_deleted_at", Table: "products", Columns: []string{"deleted_at"}},

	// Category indexes
	{Name: "idx_categories_slug", Table: "categories", Columns: []string{"slug"}, Unique: true},
	{Name: "idx_categories_parent_id", Table: "categories", Columns: []string{"parent_id"}},
	{Name: "idx_categories_is_active", Table: "categories", Columns: []string{"is_active"}},

	// Brand indexes
	{Name: "idx_brands_slug", Table: "brands", Columns: []string{"slug"}, Unique: true},
	{Name: "idx_brands_is_active", Table: "brands", Columns: []string{"is_active"}},

	// Variant indexes
	{Name: "idx_variants_sku", Table: "variants", Columns: []string{"sku"}, Unique: true},
	{Name: "idx_variants_product_id", Table: "variants", Columns: []string{"product_id"}},
	{Name: "idx_variants_barcode", Table: "variants", Columns: []string{"barcode"}},

	// Media indexes
	{Name: "idx_media_product_id", Table: "product_media", Columns: []string{"product_id"}},
	{Name: "idx_media_is_primary", Table: "product_media", Columns: []string{"product_id", "is_primary"}},

	// Price indexes
	{Name: "idx_prices_variant_id", Table: "prices", Columns: []string{"variant_id"}},
	{Name: "idx_prices_type", Table: "prices", Columns: []string{"type"}},

	// Attribute indexes
	{Name: "idx_attributes_code", Table: "attributes", Columns: []string{"code"}, Unique: true},
	{Name: "idx_product_attributes_product_id", Table: "product_attributes", Columns: []string{"product_id"}},
}

// CreateIndexes creates all defined indexes
func CreateIndexes(db *gorm.DB) error {
	for _, idx := range Indexes {
		columns := ""
		for i, col := range idx.Columns {
			if i > 0 {
				columns += ", "
			}
			columns += col
		}

		unique := ""
		if idx.Unique {
			unique = "UNIQUE "
		}

		sql := "CREATE " + unique + "INDEX IF NOT EXISTS " + idx.Name + " ON " + idx.Table + " (" + columns + ")"
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}

// CreateFullTextSearchIndex creates a full-text search index for products
func CreateFullTextSearchIndex(db *gorm.DB) error {
	sql := `
		CREATE INDEX IF NOT EXISTS idx_products_search 
		ON products USING gin(to_tsvector('english', name || ' ' || COALESCE(description, '')))
	`
	return db.Exec(sql).Error
}

