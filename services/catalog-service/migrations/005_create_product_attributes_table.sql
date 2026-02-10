-- Migration: 005_create_product_attributes_table
-- Description: Create product-attribute association table

-- +migrate Up
CREATE TABLE product_attributes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    attribute_id UUID NOT NULL REFERENCES attributes(id) ON DELETE CASCADE,
    
    -- Value storage
    text_value TEXT,
    number_value DECIMAL(15,4),
    boolean_value BOOLEAN,
    date_value DATE,
    json_value JSONB,
    
    -- For select options
    option_id UUID REFERENCES attribute_options(id) ON DELETE SET NULL,
    
    -- For multiselect
    option_ids UUID[] DEFAULT '{}',
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(product_id, attribute_id)
);

CREATE INDEX idx_product_attributes_product ON product_attributes(product_id);
CREATE INDEX idx_product_attributes_attribute ON product_attributes(attribute_id);
CREATE INDEX idx_product_attributes_option ON product_attributes(option_id) WHERE option_id IS NOT NULL;
CREATE INDEX idx_product_attributes_text ON product_attributes(attribute_id, text_value) WHERE text_value IS NOT NULL;
CREATE INDEX idx_product_attributes_number ON product_attributes(attribute_id, number_value) WHERE number_value IS NOT NULL;
CREATE INDEX idx_product_attributes_option_ids ON product_attributes USING GIN(option_ids) WHERE option_ids != '{}';

CREATE TRIGGER update_product_attributes_updated_at
    BEFORE UPDATE ON product_attributes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Helper function to get product attribute value
CREATE OR REPLACE FUNCTION get_product_attribute(
    p_product_id UUID,
    p_attribute_code VARCHAR
)
RETURNS TEXT AS $$
DECLARE
    v_type attribute_type;
    v_result TEXT;
BEGIN
    SELECT a.attribute_type INTO v_type
    FROM attributes a
    WHERE a.code = p_attribute_code;
    
    CASE v_type
        WHEN 'text' THEN
            SELECT pa.text_value INTO v_result
            FROM product_attributes pa
            JOIN attributes a ON a.id = pa.attribute_id
            WHERE pa.product_id = p_product_id AND a.code = p_attribute_code;
        WHEN 'number' THEN
            SELECT pa.number_value::TEXT INTO v_result
            FROM product_attributes pa
            JOIN attributes a ON a.id = pa.attribute_id
            WHERE pa.product_id = p_product_id AND a.code = p_attribute_code;
        WHEN 'boolean' THEN
            SELECT pa.boolean_value::TEXT INTO v_result
            FROM product_attributes pa
            JOIN attributes a ON a.id = pa.attribute_id
            WHERE pa.product_id = p_product_id AND a.code = p_attribute_code;
        WHEN 'select' THEN
            SELECT ao.label INTO v_result
            FROM product_attributes pa
            JOIN attributes a ON a.id = pa.attribute_id
            JOIN attribute_options ao ON ao.id = pa.option_id
            WHERE pa.product_id = p_product_id AND a.code = p_attribute_code;
        ELSE
            v_result := NULL;
    END CASE;
    
    RETURN v_result;
END;
$$ LANGUAGE plpgsql;

-- +migrate Down
DROP FUNCTION IF EXISTS get_product_attribute(UUID, VARCHAR);
DROP TRIGGER IF EXISTS update_product_attributes_updated_at ON product_attributes;
DROP TABLE IF EXISTS product_attributes;

