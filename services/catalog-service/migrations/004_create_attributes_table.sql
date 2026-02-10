-- Migration: 004_create_attributes_table
-- Description: Create product attributes and options

-- +migrate Up
CREATE TYPE attribute_type AS ENUM ('text', 'number', 'select', 'multiselect', 'boolean', 'date', 'color', 'size');

CREATE TABLE attributes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    attribute_type attribute_type NOT NULL DEFAULT 'text',
    
    -- Configuration
    is_required BOOLEAN NOT NULL DEFAULT FALSE,
    is_filterable BOOLEAN NOT NULL DEFAULT FALSE,
    is_searchable BOOLEAN NOT NULL DEFAULT FALSE,
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,
    is_comparable BOOLEAN NOT NULL DEFAULT FALSE,
    is_variant_attribute BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Validation
    validation_rules JSONB DEFAULT '{}',
    default_value TEXT,
    
    -- Display
    position INTEGER NOT NULL DEFAULT 0,
    group_name VARCHAR(100),
    
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_attributes_code ON attributes(code);
CREATE INDEX idx_attributes_type ON attributes(attribute_type);
CREATE INDEX idx_attributes_filterable ON attributes(is_filterable) WHERE is_filterable = TRUE;
CREATE INDEX idx_attributes_searchable ON attributes(is_searchable) WHERE is_searchable = TRUE;
CREATE INDEX idx_attributes_variant ON attributes(is_variant_attribute) WHERE is_variant_attribute = TRUE;
CREATE INDEX idx_attributes_group ON attributes(group_name) WHERE group_name IS NOT NULL;

-- Attribute options for select/multiselect types
CREATE TABLE attribute_options (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    attribute_id UUID NOT NULL REFERENCES attributes(id) ON DELETE CASCADE,
    value VARCHAR(255) NOT NULL,
    label VARCHAR(255) NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    
    -- Visual options (for colors, etc.)
    swatch_value VARCHAR(50),
    image_url VARCHAR(500),
    
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(attribute_id, value)
);

CREATE INDEX idx_attribute_options_attribute ON attribute_options(attribute_id);
CREATE INDEX idx_attribute_options_value ON attribute_options(attribute_id, value);
CREATE INDEX idx_attribute_options_position ON attribute_options(attribute_id, position);

CREATE TRIGGER update_attributes_updated_at
    BEFORE UPDATE ON attributes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +migrate Down
DROP TABLE IF EXISTS attribute_options;
DROP TRIGGER IF EXISTS update_attributes_updated_at ON attributes;
DROP TABLE IF EXISTS attributes;
DROP TYPE IF EXISTS attribute_type;

