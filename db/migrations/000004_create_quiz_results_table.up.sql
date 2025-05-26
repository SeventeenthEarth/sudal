-- Migration: Create quiz_results table
-- Description: User quiz attempts and results for the Social Quiz Platform
-- Version: 000004
-- Direction: UP

-- Create the quiz_results table in the sudal schema
-- This table stores records of users' attempts at quizzes
CREATE TABLE sudal.quiz_results (
    -- Primary key: BIGSERIAL for performance (general content, not security-sensitive)
    id BIGSERIAL PRIMARY KEY,

    -- Foreign key to users table (required)
    user_id UUID NOT NULL REFERENCES sudal.users(id) ON DELETE CASCADE,

    -- Foreign key to quiz_sets table (required)
    quiz_set_id BIGINT NOT NULL REFERENCES sudal.quiz_sets(id) ON DELETE CASCADE,

    -- User's answers stored as JSONB array of booleans
    -- Example: [true, false, true, false] for a 4-question quiz
    -- true = option A, false = option B
    answers JSONB NOT NULL,

    -- Timestamp when the quiz was submitted (UTC)
    submitted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance optimization
-- Index on user_id for filtering results by user
CREATE INDEX idx_quiz_results_user_id ON sudal.quiz_results(user_id);

-- Index on quiz_set_id for filtering results by quiz set
CREATE INDEX idx_quiz_results_quiz_set_id ON sudal.quiz_results(quiz_set_id);

-- Index on submitted_at for sorting by submission date
CREATE INDEX idx_quiz_results_submitted_at ON sudal.quiz_results(submitted_at);

-- Composite index for user's quiz history
CREATE INDEX idx_quiz_results_user_submitted ON sudal.quiz_results(user_id, submitted_at);

-- Add table and column comments for documentation
COMMENT ON TABLE sudal.quiz_results IS 'Records of users quiz attempts and their answers';
COMMENT ON COLUMN sudal.quiz_results.id IS 'BIGSERIAL-based unique quiz result identifier';
COMMENT ON COLUMN sudal.quiz_results.user_id IS 'Reference to the user who took the quiz';
COMMENT ON COLUMN sudal.quiz_results.quiz_set_id IS 'Reference to the quiz set that was taken';
COMMENT ON COLUMN sudal.quiz_results.answers IS 'JSONB array of boolean answers (true=A, false=B)';
COMMENT ON COLUMN sudal.quiz_results.submitted_at IS 'Quiz submission timestamp (UTC)';

-- Add constraints for data validation
-- Ensure answers is a valid JSON array
ALTER TABLE sudal.quiz_results ADD CONSTRAINT chk_quiz_results_answers_is_array
    CHECK (jsonb_typeof(answers) = 'array');

-- Ensure answers array is not empty
ALTER TABLE sudal.quiz_results ADD CONSTRAINT chk_quiz_results_answers_not_empty
    CHECK (jsonb_array_length(answers) > 0);