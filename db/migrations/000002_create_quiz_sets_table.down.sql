-- Migration: Create quiz_sets table
-- Description: Rollback quiz_sets table creation
-- Version: 000002
-- Direction: DOWN

-- Drop indexes first (in reverse order of creation)
DROP INDEX IF EXISTS sudal.idx_quiz_sets_created_at;
DROP INDEX IF EXISTS sudal.idx_quiz_sets_title;

-- Drop the quiz_sets table
-- This will also drop all constraints and comments associated with the table
DROP TABLE IF EXISTS sudal.quiz_sets;