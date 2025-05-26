-- Migration: Create sudal schema and users table
-- Description: Initial database schema setup for the Social Quiz Platform
-- Version: 000001
-- Direction: UP

-- Create the sudal schema if it doesn't exist
-- This schema will contain all tables for the Social Quiz Platform
CREATE SCHEMA IF NOT EXISTS sudal;

-- Create the users table in the sudal schema
-- This table stores user account information for the social quiz platform
CREATE TABLE sudal.users (
    -- Primary key: UUID for better distribution and security
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Firebase UID for authentication (required, unique)
    -- Firebase UIDs are typically 28 characters but can vary, so using 128 for safety
    firebase_uid VARCHAR(128) NOT NULL UNIQUE,

    -- User's display name (optional, can be updated by user)
    -- Reasonable length limit for display names
    display_name VARCHAR(100),

    -- URL to user's avatar image (optional)
    -- Using TEXT to accommodate long URLs
    avatar_url TEXT,

    -- Virtual currency balance (candy) - defaults to 0
    -- Using INTEGER for whole numbers, can be extended to BIGINT if needed
    candy_balance INTEGER NOT NULL DEFAULT 0,

    -- Authentication provider (required)
    -- Possible values: 'google', 'apple', 'facebook', 'email', etc.
    -- This represents the actual OAuth provider, not the platform (Firebase)
    auth_provider VARCHAR(50) NOT NULL DEFAULT 'google',

    -- Timestamp when the user account was created (UTC)
    -- Using TIMESTAMP WITH TIME ZONE to ensure UTC storage
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Timestamp when the user account was last updated (UTC)
    -- Using TIMESTAMP WITH TIME ZONE to ensure UTC storage
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance optimization
-- Index on firebase_uid for authentication lookups (most frequent query)
CREATE INDEX idx_users_firebase_uid ON sudal.users(firebase_uid);

-- Index on created_at for sorting and filtering by registration date
CREATE INDEX idx_users_created_at ON sudal.users(created_at);

-- Index on auth_provider for filtering by authentication method
CREATE INDEX idx_users_auth_provider ON sudal.users(auth_provider);

-- Add table and column comments for documentation
COMMENT ON SCHEMA sudal IS 'Main schema for the Social Quiz Platform (Sudal)';
COMMENT ON TABLE sudal.users IS 'User accounts for the Social Quiz Platform';
COMMENT ON COLUMN sudal.users.id IS 'UUID-based unique user identifier';
COMMENT ON COLUMN sudal.users.firebase_uid IS 'Firebase authentication UID (unique)';
COMMENT ON COLUMN sudal.users.display_name IS 'User display name shown in the platform';
COMMENT ON COLUMN sudal.users.avatar_url IS 'URL to user avatar image';
COMMENT ON COLUMN sudal.users.candy_balance IS 'Virtual currency balance (candy points)';
COMMENT ON COLUMN sudal.users.auth_provider IS 'OAuth provider used (e.g., google, apple, facebook, email)';
COMMENT ON COLUMN sudal.users.created_at IS 'Account creation timestamp (UTC)';
COMMENT ON COLUMN sudal.users.updated_at IS 'Last account update timestamp (UTC)';

-- Add constraint to ensure candy_balance is not negative
ALTER TABLE sudal.users ADD CONSTRAINT chk_users_candy_balance_non_negative
    CHECK (candy_balance >= 0);

-- Add constraint to ensure display_name is not empty if provided
ALTER TABLE sudal.users ADD CONSTRAINT chk_users_display_name_not_empty
    CHECK (display_name IS NULL OR LENGTH(TRIM(display_name)) > 0);