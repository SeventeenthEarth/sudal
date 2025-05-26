-- Migration: Create questions table
-- Description: Individual quiz questions for the Social Quiz Platform
-- Version: 000003
-- Direction: UP

-- Create the questions table in the sudal schema
-- This table stores individual quiz questions linked to quiz sets
CREATE TABLE sudal.questions (
    -- Primary key: BIGSERIAL for performance (general content, not security-sensitive)
    id BIGSERIAL PRIMARY KEY,

    -- Foreign key to quiz_sets table (required)
    quiz_set_id BIGINT NOT NULL REFERENCES sudal.quiz_sets(id) ON DELETE CASCADE,

    -- Question text (required)
    -- Using TEXT to accommodate longer questions
    text TEXT NOT NULL,

    -- Option A text (required)
    option_a VARCHAR(255) NOT NULL,

    -- Option B text (required)
    option_b VARCHAR(255) NOT NULL,

    -- Question order within the quiz set (required)
    -- Used for displaying questions in a specific sequence
    question_order INTEGER NOT NULL,

    -- Timestamp when the question was created (UTC)
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Timestamp when the question was last updated (UTC)
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Unique constraint to ensure no duplicate question orders within a quiz set
    CONSTRAINT uq_quiz_set_question_order UNIQUE (quiz_set_id, question_order)
);

-- Create indexes for performance optimization
-- Index on quiz_set_id for filtering questions by quiz set
CREATE INDEX idx_questions_quiz_set_id ON sudal.questions(quiz_set_id);

-- Index on quiz_set_id and question_order for ordered retrieval
CREATE INDEX idx_questions_quiz_set_order ON sudal.questions(quiz_set_id, question_order);

-- Index on created_at for sorting by creation date
CREATE INDEX idx_questions_created_at ON sudal.questions(created_at);

-- Add table and column comments for documentation
COMMENT ON TABLE sudal.questions IS 'Individual quiz questions linked to quiz sets';
COMMENT ON COLUMN sudal.questions.id IS 'BIGSERIAL-based unique question identifier';
COMMENT ON COLUMN sudal.questions.quiz_set_id IS 'Reference to the quiz set containing this question';
COMMENT ON COLUMN sudal.questions.text IS 'The question text displayed to users';
COMMENT ON COLUMN sudal.questions.option_a IS 'First answer option';
COMMENT ON COLUMN sudal.questions.option_b IS 'Second answer option';
COMMENT ON COLUMN sudal.questions.question_order IS 'Order of question within the quiz set (1-based)';
COMMENT ON COLUMN sudal.questions.created_at IS 'Question creation timestamp (UTC)';
COMMENT ON COLUMN sudal.questions.updated_at IS 'Last question update timestamp (UTC)';

-- Add constraints for data validation
-- Ensure question text is not empty
ALTER TABLE sudal.questions ADD CONSTRAINT chk_questions_text_not_empty
    CHECK (LENGTH(TRIM(text)) > 0);

-- Ensure option_a is not empty
ALTER TABLE sudal.questions ADD CONSTRAINT chk_questions_option_a_not_empty
    CHECK (LENGTH(TRIM(option_a)) > 0);

-- Ensure option_b is not empty
ALTER TABLE sudal.questions ADD CONSTRAINT chk_questions_option_b_not_empty
    CHECK (LENGTH(TRIM(option_b)) > 0);

-- Ensure question_order is positive
ALTER TABLE sudal.questions ADD CONSTRAINT chk_questions_order_positive
    CHECK (question_order > 0);