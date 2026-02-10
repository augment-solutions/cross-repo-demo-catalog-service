-- Migration: 007_create_media_table
-- Description: Create media table for product images and files

-- +migrate Up
CREATE TYPE media_type AS ENUM ('image', 'video', 'document', 'model_3d');

CREATE TABLE media (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    variant_id UUID REFERENCES variants(id) ON DELETE CASCADE,
    
    media_type media_type NOT NULL DEFAULT 'image',
    
    -- URLs
    url VARCHAR(1000) NOT NULL,
    thumbnail_url VARCHAR(1000),
    
    -- Metadata
    filename VARCHAR(255) NOT NULL,
    alt_text VARCHAR(255),
    title VARCHAR(255),
    
    -- Technical
    mime_type VARCHAR(100),
    file_size INTEGER,
    width INTEGER,
    height INTEGER,
    duration_seconds INTEGER,
    
    -- Organization
    position INTEGER NOT NULL DEFAULT 0,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Processing
    processing_status VARCHAR(50) DEFAULT 'completed',
    processing_error TEXT,
    
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_media_product ON media(product_id);
CREATE INDEX idx_media_variant ON media(variant_id) WHERE variant_id IS NOT NULL;
CREATE INDEX idx_media_type ON media(media_type);
CREATE INDEX idx_media_primary ON media(product_id, is_primary) WHERE is_primary = TRUE;
CREATE INDEX idx_media_position ON media(product_id, position);

CREATE TRIGGER update_media_updated_at
    BEFORE UPDATE ON media
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Ensure only one primary image per product
CREATE UNIQUE INDEX idx_media_single_primary ON media(product_id) 
    WHERE is_primary = TRUE AND variant_id IS NULL;

-- +migrate Down
DROP TRIGGER IF EXISTS update_media_updated_at ON media;
DROP TABLE IF EXISTS media;
DROP TYPE IF EXISTS media_type;

