-- Migration: Create comparison_photos table
-- Description: Rollback comparison_photos table creation
-- Version: 000007
-- Direction: DOWN

-- Drop indexes first (in reverse order of creation)
DROP INDEX IF EXISTS sudal.idx_comparison_photos_uploaded_at;
DROP INDEX IF EXISTS sudal.idx_comparison_photos_uploader_user_id;
DROP INDEX IF EXISTS sudal.idx_comparison_photos_comparison_id;

-- Drop the comparison_photos table
-- This will also drop all constraints and comments associated with the table
DROP TABLE IF EXISTS sudal.comparison_photos;