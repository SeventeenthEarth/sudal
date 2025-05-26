-- Migration: Create quiz_set_tags table
-- Description: Many-to-many join table linking quiz sets to tags
-- Version: 000010
-- Direction: UP

-- Create the quiz_set_tags table in the sudal schema
-- This is a many-to-many join table linking quiz sets to tags
CREATE TABLE sudal.quiz_set_tags (
    -- Foreign key to quiz_sets table (required)
    quiz_set_id BIGINT NOT NULL REFERENCES sudal.quiz_sets(id) ON DELETE CASCADE,

    -- Foreign key to tags table (required)
    tag_id BIGINT NOT NULL REFERENCES sudal.tags(id) ON DELETE CASCADE,

    -- Composite primary key to ensure unique quiz_set/tag combinations
    PRIMARY KEY (quiz_set_id, tag_id)
);

-- Create indexes for performance optimization
-- Index on quiz_set_id for finding tags by quiz set
CREATE INDEX idx_quiz_set_tags_quiz_set_id ON sudal.quiz_set_tags(quiz_set_id);

-- Index on tag_id for finding quiz sets by tag
CREATE INDEX idx_quiz_set_tags_tag_id ON sudal.quiz_set_tags(tag_id);

-- Add table and column comments for documentation
COMMENT ON TABLE sudal.quiz_set_tags IS 'Many-to-many join table linking quiz sets to tags';
COMMENT ON COLUMN sudal.quiz_set_tags.quiz_set_id IS 'Reference to the quiz set';
COMMENT ON COLUMN sudal.quiz_set_tags.tag_id IS 'Reference to the tag';