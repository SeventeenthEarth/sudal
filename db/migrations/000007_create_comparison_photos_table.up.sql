-- Migration: Create comparison_photos table
-- Description: Photos uploaded for comparison events
-- Version: 000007
-- Direction: UP

-- Create the comparison_photos table in the sudal schema
-- This table stores URLs of photos uploaded for a comparison event
CREATE TABLE sudal.comparison_photos (
    -- Primary key: UUID for security (photos are sensitive content)
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Foreign key to comparisons table (required)
    comparison_id BIGINT NOT NULL REFERENCES sudal.comparisons(id) ON DELETE CASCADE,

    -- Foreign key to users table (required)
    -- The user who uploaded this photo
    uploader_user_id UUID NOT NULL REFERENCES sudal.users(id),

    -- URL to the uploaded photo (required)
    -- Using TEXT to accommodate long URLs
    photo_url TEXT NOT NULL,

    -- Timestamp when the photo was uploaded (UTC)
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance optimization
-- Index on comparison_id for filtering photos by comparison
CREATE INDEX idx_comparison_photos_comparison_id ON sudal.comparison_photos(comparison_id);

-- Index on uploader_user_id for filtering photos by uploader
CREATE INDEX idx_comparison_photos_uploader_user_id ON sudal.comparison_photos(uploader_user_id);

-- Index on uploaded_at for sorting by upload date
CREATE INDEX idx_comparison_photos_uploaded_at ON sudal.comparison_photos(uploaded_at);

-- Add table and column comments for documentation
COMMENT ON TABLE sudal.comparison_photos IS 'Photos uploaded for comparison events';
COMMENT ON COLUMN sudal.comparison_photos.id IS 'UUID-based unique photo record identifier';
COMMENT ON COLUMN sudal.comparison_photos.comparison_id IS 'Reference to the comparison event';
COMMENT ON COLUMN sudal.comparison_photos.uploader_user_id IS 'Reference to the user who uploaded the photo';
COMMENT ON COLUMN sudal.comparison_photos.photo_url IS 'URL to the uploaded photo';
COMMENT ON COLUMN sudal.comparison_photos.uploaded_at IS 'Photo upload timestamp (UTC)';

-- Add constraints for data validation
-- Ensure photo_url is not empty
ALTER TABLE sudal.comparison_photos ADD CONSTRAINT chk_comparison_photos_url_not_empty
    CHECK (LENGTH(TRIM(photo_url)) > 0);