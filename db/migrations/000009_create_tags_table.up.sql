-- Migration: Create tags table
-- Description: Tags for categorizing quiz sets
-- Version: 000009
-- Direction: UP

-- Create the tags table in the sudal schema
-- This table stores tags used to categorize quiz sets
CREATE TABLE sudal.tags (
    -- Primary key: BIGSERIAL for performance (general content, not security-sensitive)
    id BIGSERIAL PRIMARY KEY,

    -- Tag name (required, unique)
    -- Reasonable length limit for tag names
    name VARCHAR(100) NOT NULL UNIQUE,

    -- Timestamp when the tag was created (UTC)
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance optimization
-- Index on name for searching and sorting (also enforces uniqueness)
CREATE INDEX idx_tags_name ON sudal.tags(name);

-- Index on created_at for sorting by creation date
CREATE INDEX idx_tags_created_at ON sudal.tags(created_at);

-- Add table and column comments for documentation
COMMENT ON TABLE sudal.tags IS 'Tags used to categorize quiz sets';
COMMENT ON COLUMN sudal.tags.id IS 'BIGSERIAL-based unique tag identifier';
COMMENT ON COLUMN sudal.tags.name IS 'Tag name (unique across all tags)';
COMMENT ON COLUMN sudal.tags.created_at IS 'Tag creation timestamp (UTC)';

-- Add constraints for data validation
-- Ensure tag name is not empty
ALTER TABLE sudal.tags ADD CONSTRAINT chk_tags_name_not_empty
    CHECK (LENGTH(TRIM(name)) > 0);