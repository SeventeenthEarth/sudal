-- Migration: Update users table to match PR specifications
-- Description: Add deleted_at column, unique constraint on display_name, and remove auth_provider
-- Version: 000011
-- Direction: UP

-- Add deleted_at column for soft deletes
ALTER TABLE sudal.users
ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

-- Add comment for the new column
COMMENT ON COLUMN sudal.users.deleted_at IS 'Soft delete timestamp (UTC) - NULL means not deleted';

-- Add unique constraint on display_name that allows multiple NULL values
-- PostgreSQL treats NULL values as distinct, so multiple NULL values are allowed
-- This constraint will only enforce uniqueness for non-NULL display_name values
ALTER TABLE sudal.users
ADD CONSTRAINT uq_users_display_name_unique
UNIQUE (display_name) DEFERRABLE INITIALLY DEFERRED;

-- Add index on deleted_at for soft delete queries
CREATE INDEX idx_users_deleted_at ON sudal.users(deleted_at);

-- Update table comment to reflect soft delete capability
COMMENT ON TABLE sudal.users IS 'User accounts for the Social Quiz Platform (supports soft deletes)';