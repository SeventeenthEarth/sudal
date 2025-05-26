-- Migration: Create comparison_participants table
-- Description: Links users and their quiz results to comparison events
-- Version: 000006
-- Direction: UP

-- Create the comparison_participants table in the sudal schema
-- This table links users and their specific quiz results to a comparison event, and stores personal memos
CREATE TABLE sudal.comparison_participants (
    -- Primary key: BIGSERIAL for performance (general content, not security-sensitive)
    id BIGSERIAL PRIMARY KEY,

    -- Foreign key to comparisons table (required)
    comparison_id BIGINT NOT NULL REFERENCES sudal.comparisons(id) ON DELETE CASCADE,

    -- Foreign key to users table (required)
    user_id UUID NOT NULL REFERENCES sudal.users(id) ON DELETE CASCADE,

    -- Foreign key to quiz_results table (required)
    -- The specific quiz result this user is bringing to the comparison
    quiz_result_id BIGINT NOT NULL REFERENCES sudal.quiz_results(id),

    -- Personal memo/note from the user (optional)
    -- Using TEXT to accommodate longer memos
    personal_memo TEXT,

    -- Timestamp when the user joined the comparison (UTC)
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Unique constraint to ensure a user can only participate once per comparison
    CONSTRAINT uq_comparison_user UNIQUE (comparison_id, user_id)
);

-- Create indexes for performance optimization
-- Index on comparison_id for filtering participants by comparison
CREATE INDEX idx_comparison_participants_comparison_id ON sudal.comparison_participants(comparison_id);

-- Index on user_id for filtering participations by user
CREATE INDEX idx_comparison_participants_user_id ON sudal.comparison_participants(user_id);

-- Index on quiz_result_id for finding which comparisons use a specific quiz result
CREATE INDEX idx_comparison_participants_quiz_result_id ON sudal.comparison_participants(quiz_result_id);

-- Index on created_at for sorting by participation date
CREATE INDEX idx_comparison_participants_created_at ON sudal.comparison_participants(created_at);

-- Add table and column comments for documentation
COMMENT ON TABLE sudal.comparison_participants IS 'Links users and their quiz results to comparison events with personal memos';
COMMENT ON COLUMN sudal.comparison_participants.id IS 'BIGSERIAL-based unique participant record identifier';
COMMENT ON COLUMN sudal.comparison_participants.comparison_id IS 'Reference to the comparison event';
COMMENT ON COLUMN sudal.comparison_participants.user_id IS 'Reference to the participating user';
COMMENT ON COLUMN sudal.comparison_participants.quiz_result_id IS 'Reference to the quiz result being compared';
COMMENT ON COLUMN sudal.comparison_participants.personal_memo IS 'Optional personal note from the user';
COMMENT ON COLUMN sudal.comparison_participants.created_at IS 'Participation timestamp (UTC)';