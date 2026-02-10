-- Migration: 001_create_products_table
-- Description: Create products table for product catalog

-- +migrate Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE product_status AS ENUM ('draft', 'active', 'inactive', 'archived', 'discontinued');
CREATE TYPE product_type AS ENUM ('simple', 'configurable', 'bundle', 'virtual', 'downloadable');

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sku VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    short_description VARCHAR(500),
    
    -- Classification
    product_type product_type NOT NULL DEFAULT 'simple',
    status product_status NOT NULL DEFAULT 'draft',
    
    -- Pricing (base)
    base_price DECIMAL(12,2),
    compare_at_price DECIMAL(12,2),
    cost_price DECIMAL(12,2),
    currency VARCHAR(3) DEFAULT 'USD',
    
    -- Inventory
    track_inventory BOOLEAN NOT NULL DEFAULT TRUE,
    allow_backorder BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Shipping
    weight DECIMAL(10,3),
    weight_unit VARCHAR(10) DEFAULT 'kg',
    length DECIMAL(10,2),
    width DECIMAL(10,2),
    height DECIMAL(10,2),
    dimension_unit VARCHAR(10) DEFAULT 'cm',
    requires_shipping BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- SEO
    meta_title VARCHAR(255),
    meta_description VARCHAR(500),
    meta_keywords TEXT[],
    
    -- Organization
    brand_id UUID,
    vendor VARCHAR(100),
    tags TEXT[] DEFAULT '{}',
    
    -- Visibility
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,
    is_featured BOOLEAN NOT NULL DEFAULT FALSE,
    
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_slug ON products(slug);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_type ON products(product_type);
CREATE INDEX idx_products_brand ON products(brand_id) WHERE brand_id IS NOT NULL;
CREATE INDEX idx_products_featured ON products(is_featured) WHERE is_featured = TRUE;
CREATE INDEX idx_products_visible ON products(is_visible) WHERE is_visible = TRUE;
CREATE INDEX idx_products_tags ON products USING GIN(tags);
CREATE INDEX idx_products_created_at ON products(created_at DESC);
CREATE INDEX idx_products_deleted ON products(deleted_at) WHERE deleted_at IS NOT NULL;

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +migrate Down
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS products;
DROP TYPE IF EXISTS product_type;
DROP TYPE IF EXISTS product_status;

