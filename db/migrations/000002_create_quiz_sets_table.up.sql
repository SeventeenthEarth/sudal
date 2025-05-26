-- Migration: Create quiz_sets table
-- Description: Quiz collections for the Social Quiz Platform
-- Version: 000002
-- Direction: UP

-- Create the quiz_sets table in the sudal schema
-- This table stores collections of quiz questions
CREATE TABLE sudal.quiz_sets (
    -- Primary key: BIGSERIAL for performance (general content, not security-sensitive)
    id BIGSERIAL PRIMARY KEY,

    -- Quiz set title (required)
    -- Reasonable length limit for quiz titles
    title VARCHAR(255) NOT NULL,

    -- Quiz set description (optional)
    -- Using TEXT to accommodate longer descriptions
    description TEXT,

    -- Timestamp when the quiz set was created (UTC)
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Timestamp when the quiz set was last updated (UTC)
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance optimization
-- Index on title for searching and sorting
CREATE INDEX idx_quiz_sets_title ON sudal.quiz_sets(title);

-- Index on created_at for sorting by creation date
CREATE INDEX idx_quiz_sets_created_at ON sudal.quiz_sets(created_at);

-- Add table and column comments for documentation
COMMENT ON TABLE sudal.quiz_sets IS 'Collections of quiz questions for the Social Quiz Platform';
COMMENT ON COLUMN sudal.quiz_sets.id IS 'BIGSERIAL-based unique quiz set identifier';
COMMENT ON COLUMN sudal.quiz_sets.title IS 'Quiz set title displayed to users';
COMMENT ON COLUMN sudal.quiz_sets.description IS 'Optional description of the quiz set';
COMMENT ON COLUMN sudal.quiz_sets.created_at IS 'Quiz set creation timestamp (UTC)';
COMMENT ON COLUMN sudal.quiz_sets.updated_at IS 'Last quiz set update timestamp (UTC)';

-- Add constraint to ensure title is not empty
ALTER TABLE sudal.quiz_sets ADD CONSTRAINT chk_quiz_sets_title_not_empty
    CHECK (LENGTH(TRIM(title)) > 0);