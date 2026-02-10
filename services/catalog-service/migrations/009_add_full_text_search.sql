-- Migration: 009_add_full_text_search
-- Description: Add full-text search capabilities to catalog

-- +migrate Up

-- Add search vector to products
ALTER TABLE products ADD COLUMN search_vector tsvector;

-- Create GIN index for full-text search
CREATE INDEX idx_products_search_vector ON products USING GIN(search_vector);

-- Function to update search vector
CREATE OR REPLACE FUNCTION products_search_vector_update()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector := 
        setweight(to_tsvector('english', COALESCE(NEW.name, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.sku, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.short_description, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'C') ||
        setweight(to_tsvector('english', COALESCE(array_to_string(NEW.tags, ' '), '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update search vector
CREATE TRIGGER products_search_vector_trigger
    BEFORE INSERT OR UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION products_search_vector_update();

-- Update existing rows
UPDATE products SET search_vector = 
    setweight(to_tsvector('english', COALESCE(name, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(sku, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(short_description, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(description, '')), 'C') ||
    setweight(to_tsvector('english', COALESCE(array_to_string(tags, ' '), '')), 'B');

-- Add search vector to categories
ALTER TABLE categories ADD COLUMN search_vector tsvector;
CREATE INDEX idx_categories_search_vector ON categories USING GIN(search_vector);

CREATE OR REPLACE FUNCTION categories_search_vector_update()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector := 
        setweight(to_tsvector('english', COALESCE(NEW.name, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER categories_search_vector_trigger
    BEFORE INSERT OR UPDATE ON categories
    FOR EACH ROW
    EXECUTE FUNCTION categories_search_vector_update();

UPDATE categories SET search_vector = 
    setweight(to_tsvector('english', COALESCE(name, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(description, '')), 'B');

-- Search function for products
CREATE OR REPLACE FUNCTION search_products(
    search_query TEXT,
    category_uuid UUID DEFAULT NULL,
    brand_uuid UUID DEFAULT NULL,
    min_price DECIMAL DEFAULT NULL,
    max_price DECIMAL DEFAULT NULL,
    limit_count INTEGER DEFAULT 50,
    offset_count INTEGER DEFAULT 0
)
RETURNS TABLE (
    product_id UUID,
    product_name VARCHAR,
    product_sku VARCHAR,
    product_slug VARCHAR,
    base_price DECIMAL,
    rank REAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        p.id,
        p.name,
        p.sku,
        p.slug,
        p.base_price,
        ts_rank(p.search_vector, plainto_tsquery('english', search_query)) as rank
    FROM products p
    LEFT JOIN product_categories pc ON pc.product_id = p.id AND pc.is_primary = TRUE
    WHERE p.search_vector @@ plainto_tsquery('english', search_query)
      AND p.status = 'active'
      AND p.is_visible = TRUE
      AND p.deleted_at IS NULL
      AND (category_uuid IS NULL OR pc.category_id = category_uuid)
      AND (brand_uuid IS NULL OR p.brand_id = brand_uuid)
      AND (min_price IS NULL OR p.base_price >= min_price)
      AND (max_price IS NULL OR p.base_price <= max_price)
    ORDER BY rank DESC, p.created_at DESC
    LIMIT limit_count
    OFFSET offset_count;
END;
$$ LANGUAGE plpgsql;

-- +migrate Down
DROP FUNCTION IF EXISTS search_products(TEXT, UUID, UUID, DECIMAL, DECIMAL, INTEGER, INTEGER);
DROP TRIGGER IF EXISTS categories_search_vector_trigger ON categories;
DROP FUNCTION IF EXISTS categories_search_vector_update();
DROP INDEX IF EXISTS idx_categories_search_vector;
ALTER TABLE categories DROP COLUMN IF EXISTS search_vector;
DROP TRIGGER IF EXISTS products_search_vector_trigger ON products;
DROP FUNCTION IF EXISTS products_search_vector_update();
DROP INDEX IF EXISTS idx_products_search_vector;
ALTER TABLE products DROP COLUMN IF EXISTS search_vector;

