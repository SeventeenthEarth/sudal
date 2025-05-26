-- Migration: Create tags table
-- Description: Rollback tags table creation
-- Version: 000009
-- Direction: DOWN

-- Drop indexes first (in reverse order of creation)
DROP INDEX IF EXISTS sudal.idx_tags_created_at;
DROP INDEX IF EXISTS sudal.idx_tags_name;

-- Drop the tags table
-- This will also drop all constraints and comments associated with the table
DROP TABLE IF EXISTS sudal.tags;