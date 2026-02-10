-- Migration: 003_create_brands_table
-- Description: Create brands table for brand management

-- +migrate Up
CREATE TABLE brands (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    
    -- Display
    logo_url VARCHAR(500),
    banner_url VARCHAR(500),
    website_url VARCHAR(500),
    
    -- Status
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_featured BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- SEO
    meta_title VARCHAR(255),
    meta_description VARCHAR(500),
    
    -- Stats
    product_count INTEGER NOT NULL DEFAULT 0,
    
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_brands_slug ON brands(slug);
CREATE INDEX idx_brands_active ON brands(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_brands_featured ON brands(is_featured) WHERE is_featured = TRUE;
CREATE INDEX idx_brands_name_search ON brands USING GIN(name gin_trgm_ops);

-- Add foreign key to products
ALTER TABLE products ADD CONSTRAINT fk_products_brand 
    FOREIGN KEY (brand_id) REFERENCES brands(id) ON DELETE SET NULL;

CREATE TRIGGER update_brands_updated_at
    BEFORE UPDATE ON brands
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to update brand product count
CREATE OR REPLACE FUNCTION update_brand_product_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        UPDATE brands SET product_count = (
            SELECT COUNT(*) FROM products 
            WHERE brand_id = NEW.brand_id AND deleted_at IS NULL
        ) WHERE id = NEW.brand_id;
    END IF;
    
    IF TG_OP = 'DELETE' OR (TG_OP = 'UPDATE' AND OLD.brand_id IS DISTINCT FROM NEW.brand_id) THEN
        UPDATE brands SET product_count = (
            SELECT COUNT(*) FROM products 
            WHERE brand_id = OLD.brand_id AND deleted_at IS NULL
        ) WHERE id = OLD.brand_id;
    END IF;
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER products_brand_count_trigger
    AFTER INSERT OR UPDATE OR DELETE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_brand_product_count();

-- +migrate Down
DROP TRIGGER IF EXISTS products_brand_count_trigger ON products;
DROP FUNCTION IF EXISTS update_brand_product_count();
DROP TRIGGER IF EXISTS update_brands_updated_at ON brands;
ALTER TABLE products DROP CONSTRAINT IF EXISTS fk_products_brand;
DROP TABLE IF EXISTS brands;

