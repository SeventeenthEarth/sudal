-- Migration: Create questions table
-- Description: Rollback questions table creation
-- Version: 000003
-- Direction: DOWN

-- Drop indexes first (in reverse order of creation)
DROP INDEX IF EXISTS sudal.idx_questions_created_at;
DROP INDEX IF EXISTS sudal.idx_questions_quiz_set_order;
DROP INDEX IF EXISTS sudal.idx_questions_quiz_set_id;

-- Drop the questions table
-- This will also drop all constraints and comments associated with the table
DROP TABLE IF EXISTS sudal.questions;