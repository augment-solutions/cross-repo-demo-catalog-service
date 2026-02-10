-- Migration: 006_create_variants_table
-- Description: Create product variants table

-- +migrate Up
CREATE TYPE variant_status AS ENUM ('active', 'inactive', 'discontinued');

CREATE TABLE variants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255),
    
    status variant_status NOT NULL DEFAULT 'active',
    
    -- Pricing
    price DECIMAL(12,2) NOT NULL,
    compare_at_price DECIMAL(12,2),
    cost_price DECIMAL(12,2),
    
    -- Inventory
    quantity INTEGER NOT NULL DEFAULT 0,
    low_stock_threshold INTEGER DEFAULT 5,
    
    -- Shipping
    weight DECIMAL(10,3),
    weight_unit VARCHAR(10) DEFAULT 'kg',
    
    -- Display
    position INTEGER NOT NULL DEFAULT 0,
    image_url VARCHAR(500),
    
    -- Variant options (e.g., {"color": "red", "size": "XL"})
    options JSONB NOT NULL DEFAULT '{}',
    
    barcode VARCHAR(100),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_variants_product ON variants(product_id);
CREATE INDEX idx_variants_sku ON variants(sku);
CREATE INDEX idx_variants_status ON variants(status);
CREATE INDEX idx_variants_barcode ON variants(barcode) WHERE barcode IS NOT NULL;
CREATE INDEX idx_variants_options ON variants USING GIN(options jsonb_path_ops);
CREATE INDEX idx_variants_position ON variants(product_id, position);
CREATE INDEX idx_variants_quantity ON variants(quantity);
CREATE INDEX idx_variants_deleted ON variants(deleted_at) WHERE deleted_at IS NOT NULL;

CREATE TRIGGER update_variants_updated_at
    BEFORE UPDATE ON variants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Variant option values table for quick filtering
CREATE TABLE variant_option_values (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    variant_id UUID NOT NULL REFERENCES variants(id) ON DELETE CASCADE,
    attribute_id UUID NOT NULL REFERENCES attributes(id) ON DELETE CASCADE,
    option_id UUID REFERENCES attribute_options(id) ON DELETE SET NULL,
    value VARCHAR(255) NOT NULL,
    UNIQUE(variant_id, attribute_id)
);

CREATE INDEX idx_variant_options_variant ON variant_option_values(variant_id);
CREATE INDEX idx_variant_options_attribute ON variant_option_values(attribute_id);
CREATE INDEX idx_variant_options_option ON variant_option_values(option_id) WHERE option_id IS NOT NULL;
CREATE INDEX idx_variant_options_value ON variant_option_values(attribute_id, value);

-- +migrate Down
DROP TABLE IF EXISTS variant_option_values;
DROP TRIGGER IF EXISTS update_variants_updated_at ON variants;
DROP TABLE IF EXISTS variants;
DROP TYPE IF EXISTS variant_status;

