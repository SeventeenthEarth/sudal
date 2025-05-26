-- Migration: Create quiz_set_tags table
-- Description: Rollback quiz_set_tags table creation
-- Version: 000010
-- Direction: DOWN

-- Drop indexes first (in reverse order of creation)
DROP INDEX IF EXISTS sudal.idx_quiz_set_tags_tag_id;
DROP INDEX IF EXISTS sudal.idx_quiz_set_tags_quiz_set_id;

-- Drop the quiz_set_tags table
-- This will also drop all constraints and comments associated with the table
DROP TABLE IF EXISTS sudal.quiz_set_tags;