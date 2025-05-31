-- Migration: Update users table to match PR specifications
-- Description: Rollback changes - remove deleted_at, restore auth_provider, remove display_name uniqueness
-- Version: 000011
-- Direction: DOWN

-- Drop indexes and constraints created in the up migration
DROP INDEX IF EXISTS sudal.idx_users_deleted_at;
ALTER TABLE sudal.users
DROP CONSTRAINT IF EXISTS uq_users_display_name_unique;

-- Remove the deleted_at column
ALTER TABLE sudal.users
DROP COLUMN IF EXISTS deleted_at;

-- Restore the original index on auth_provider
CREATE INDEX idx_users_auth_provider ON sudal.users(auth_provider);

-- Restore original table comment
COMMENT ON TABLE sudal.users IS 'User accounts for the Social Quiz Platform';

-- Restore comment for auth_provider column
COMMENT ON COLUMN sudal.users.auth_provider IS 'OAuth provider used (e.g., google, apple, facebook, email)';