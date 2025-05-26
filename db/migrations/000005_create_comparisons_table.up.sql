-- Migration: Create comparisons table
-- Description: Comparison events for the Social Quiz Platform
-- Version: 000005
-- Direction: UP

-- Create the comparisons table in the sudal schema
-- This table stores information about comparison events where users compare their quiz results
CREATE TABLE sudal.comparisons (
    -- Primary key: BIGSERIAL for performance (general content, not security-sensitive)
    id BIGSERIAL PRIMARY KEY,

    -- Foreign key to quiz_sets table (required)
    -- The quiz set that this comparison is based on
    quiz_set_id BIGINT NOT NULL REFERENCES sudal.quiz_sets(id),

    -- Room identifier for the comparison session (required)
    -- Used to group users in the same comparison session
    room_id VARCHAR(255) NOT NULL,

    -- Timestamp when the comparison was created (UTC)
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance optimization
-- Index on quiz_set_id for filtering comparisons by quiz set
CREATE INDEX idx_comparisons_quiz_set_id ON sudal.comparisons(quiz_set_id);

-- Index on room_id for finding comparisons by room (frequently queried)
CREATE INDEX idx_comparisons_room_id ON sudal.comparisons(room_id);

-- Index on created_at for sorting by creation date
CREATE INDEX idx_comparisons_created_at ON sudal.comparisons(created_at);

-- Add table and column comments for documentation
COMMENT ON TABLE sudal.comparisons IS 'Comparison events where users compare their quiz results';
COMMENT ON COLUMN sudal.comparisons.id IS 'BIGSERIAL-based unique comparison identifier';
COMMENT ON COLUMN sudal.comparisons.quiz_set_id IS 'Reference to the quiz set being compared';
COMMENT ON COLUMN sudal.comparisons.room_id IS 'Room identifier for grouping users in comparison session';
COMMENT ON COLUMN sudal.comparisons.created_at IS 'Comparison creation timestamp (UTC)';

-- Add constraints for data validation
-- Ensure room_id is not empty
ALTER TABLE sudal.comparisons ADD CONSTRAINT chk_comparisons_room_id_not_empty
    CHECK (LENGTH(TRIM(room_id)) > 0);