-- Migration: 002_create_categories_table
-- Description: Create categories table with hierarchical structure

-- +migrate Up
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    
    -- Hierarchy
    level INTEGER NOT NULL DEFAULT 0,
    path TEXT NOT NULL DEFAULT '',
    position INTEGER NOT NULL DEFAULT 0,
    
    -- Display
    image_url VARCHAR(500),
    icon VARCHAR(50),
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,
    is_featured BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- SEO
    meta_title VARCHAR(255),
    meta_description VARCHAR(500),
    
    -- Counts
    product_count INTEGER NOT NULL DEFAULT 0,
    
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    UNIQUE(parent_id, slug)
);

CREATE INDEX idx_categories_parent_id ON categories(parent_id);
CREATE INDEX idx_categories_slug ON categories(slug);
CREATE INDEX idx_categories_level ON categories(level);
CREATE INDEX idx_categories_path ON categories(path);
CREATE INDEX idx_categories_visible ON categories(is_visible) WHERE is_visible = TRUE;
CREATE INDEX idx_categories_featured ON categories(is_featured) WHERE is_featured = TRUE;
CREATE INDEX idx_categories_position ON categories(parent_id, position);

CREATE TRIGGER update_categories_updated_at
    BEFORE UPDATE ON categories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to update category path
CREATE OR REPLACE FUNCTION update_category_path()
RETURNS TRIGGER AS $$
DECLARE
    parent_path TEXT;
BEGIN
    IF NEW.parent_id IS NULL THEN
        NEW.path := NEW.id::TEXT;
        NEW.level := 0;
    ELSE
        SELECT path, level INTO parent_path FROM categories WHERE id = NEW.parent_id;
        NEW.path := parent_path || '/' || NEW.id::TEXT;
        NEW.level := (SELECT level + 1 FROM categories WHERE id = NEW.parent_id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER categories_path_trigger
    BEFORE INSERT OR UPDATE OF parent_id ON categories
    FOR EACH ROW
    EXECUTE FUNCTION update_category_path();

-- +migrate Down
DROP TRIGGER IF EXISTS categories_path_trigger ON categories;
DROP FUNCTION IF EXISTS update_category_path();
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;
DROP TABLE IF EXISTS categories;

