-- Migration: 010_add_catalog_indexes
-- Description: Add performance indexes for catalog queries

-- +migrate Up
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Composite indexes for common product queries
CREATE INDEX idx_products_status_visible ON products(status, is_visible) 
    WHERE status = 'active' AND is_visible = TRUE AND deleted_at IS NULL;

CREATE INDEX idx_products_brand_status ON products(brand_id, status) WHERE brand_id IS NOT NULL;
CREATE INDEX idx_products_price_range ON products(base_price, currency) WHERE status = 'active';
CREATE INDEX idx_products_featured_active ON products(is_featured, created_at DESC) 
    WHERE is_featured = TRUE AND status = 'active';

-- Trigram indexes for autocomplete
CREATE INDEX idx_products_name_trgm ON products USING GIN(name gin_trgm_ops);
CREATE INDEX idx_products_sku_trgm ON products USING GIN(sku gin_trgm_ops);
CREATE INDEX idx_categories_name_trgm ON categories USING GIN(name gin_trgm_ops);

-- BRIN indexes for time-series
CREATE INDEX idx_products_created_brin ON products USING BRIN(created_at);
CREATE INDEX idx_products_updated_brin ON products USING BRIN(updated_at);

-- Metadata indexes
CREATE INDEX idx_products_metadata ON products USING GIN(metadata jsonb_path_ops);
CREATE INDEX idx_variants_metadata ON variants USING GIN(metadata jsonb_path_ops);

-- Covering indexes for list queries
CREATE INDEX idx_products_list ON products(status, created_at DESC) 
    INCLUDE (id, name, slug, base_price, brand_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_variants_list ON variants(product_id, position)
    INCLUDE (id, sku, price, quantity)
    WHERE deleted_at IS NULL;

-- Category hierarchy queries
CREATE INDEX idx_categories_hierarchy ON categories(parent_id, position, is_visible);
CREATE INDEX idx_categories_path_prefix ON categories(path text_pattern_ops);

-- Inventory queries
CREATE INDEX idx_variants_low_stock ON variants(product_id, quantity)
    WHERE quantity <= low_stock_threshold AND status = 'active';

CREATE INDEX idx_variants_out_of_stock ON variants(product_id)
    WHERE quantity = 0 AND status = 'active';

-- +migrate Down
DROP INDEX IF EXISTS idx_variants_out_of_stock;
DROP INDEX IF EXISTS idx_variants_low_stock;
DROP INDEX IF EXISTS idx_categories_path_prefix;
DROP INDEX IF EXISTS idx_categories_hierarchy;
DROP INDEX IF EXISTS idx_variants_list;
DROP INDEX IF EXISTS idx_products_list;
DROP INDEX IF EXISTS idx_variants_metadata;
DROP INDEX IF EXISTS idx_products_metadata;
DROP INDEX IF EXISTS idx_products_updated_brin;
DROP INDEX IF EXISTS idx_products_created_brin;
DROP INDEX IF EXISTS idx_categories_name_trgm;
DROP INDEX IF EXISTS idx_products_sku_trgm;
DROP INDEX IF EXISTS idx_products_name_trgm;
DROP INDEX IF EXISTS idx_products_featured_active;
DROP INDEX IF EXISTS idx_products_price_range;
DROP INDEX IF EXISTS idx_products_brand_status;
DROP INDEX IF EXISTS idx_products_status_visible;

