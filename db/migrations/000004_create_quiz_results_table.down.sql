-- Migration: Create quiz_results table
-- Description: Rollback quiz_results table creation
-- Version: 000004
-- Direction: DOWN

-- Drop indexes first (in reverse order of creation)
DROP INDEX IF EXISTS sudal.idx_quiz_results_user_submitted;
DROP INDEX IF EXISTS sudal.idx_quiz_results_submitted_at;
DROP INDEX IF EXISTS sudal.idx_quiz_results_quiz_set_id;
DROP INDEX IF EXISTS sudal.idx_quiz_results_user_id;

-- Drop the quiz_results table
-- This will also drop all constraints and comments associated with the table
DROP TABLE IF EXISTS sudal.quiz_results;