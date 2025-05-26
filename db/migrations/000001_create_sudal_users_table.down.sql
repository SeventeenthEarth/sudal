-- Migration: Create sudal schema and users table
-- Description: Rollback the initial database schema setup
-- Version: 000001
-- Direction: DOWN

-- Drop indexes first (in reverse order of creation)
DROP INDEX IF EXISTS sudal.idx_users_auth_provider;
DROP INDEX IF EXISTS sudal.idx_users_created_at;
DROP INDEX IF EXISTS sudal.idx_users_firebase_uid;

-- Drop the users table
-- This will also drop all constraints and comments associated with the table
DROP TABLE IF EXISTS sudal.users;

-- Note: We do NOT drop the sudal schema here because:
-- 1. Future migrations will add more tables to this schema
-- 2. Schema cleanup should be handled by broader reset commands
-- 3. This allows for safer rollback operations