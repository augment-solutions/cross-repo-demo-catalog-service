-- Migration: 008_create_product_categories_table
-- Description: Create product-category association table

-- +migrate Up
CREATE TABLE product_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    position INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(product_id, category_id)
);

CREATE INDEX idx_product_categories_product ON product_categories(product_id);
CREATE INDEX idx_product_categories_category ON product_categories(category_id);
CREATE INDEX idx_product_categories_primary ON product_categories(product_id, is_primary) WHERE is_primary = TRUE;

-- Ensure only one primary category per product
CREATE UNIQUE INDEX idx_product_categories_single_primary ON product_categories(product_id) 
    WHERE is_primary = TRUE;

-- Function to update category product counts
CREATE OR REPLACE FUNCTION update_category_product_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        UPDATE categories SET product_count = (
            SELECT COUNT(DISTINCT pc.product_id) 
            FROM product_categories pc
            JOIN products p ON p.id = pc.product_id
            WHERE pc.category_id = NEW.category_id 
              AND p.status = 'active' 
              AND p.deleted_at IS NULL
        ) WHERE id = NEW.category_id;
    END IF;
    
    IF TG_OP = 'DELETE' OR (TG_OP = 'UPDATE' AND OLD.category_id IS DISTINCT FROM NEW.category_id) THEN
        UPDATE categories SET product_count = (
            SELECT COUNT(DISTINCT pc.product_id) 
            FROM product_categories pc
            JOIN products p ON p.id = pc.product_id
            WHERE pc.category_id = OLD.category_id 
              AND p.status = 'active' 
              AND p.deleted_at IS NULL
        ) WHERE id = OLD.category_id;
    END IF;
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER product_categories_count_trigger
    AFTER INSERT OR UPDATE OR DELETE ON product_categories
    FOR EACH ROW
    EXECUTE FUNCTION update_category_product_count();

-- +migrate Down
DROP TRIGGER IF EXISTS product_categories_count_trigger ON product_categories;
DROP FUNCTION IF EXISTS update_category_product_count();
DROP TABLE IF EXISTS product_categories;

