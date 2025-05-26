-- Migration: Create comparison_participants table
-- Description: Rollback comparison_participants table creation
-- Version: 000006
-- Direction: DOWN

-- Drop indexes first (in reverse order of creation)
DROP INDEX IF EXISTS sudal.idx_comparison_participants_created_at;
DROP INDEX IF EXISTS sudal.idx_comparison_participants_quiz_result_id;
DROP INDEX IF EXISTS sudal.idx_comparison_participants_user_id;
DROP INDEX IF EXISTS sudal.idx_comparison_participants_comparison_id;

-- Drop the comparison_participants table
-- This will also drop all constraints and comments associated with the table
DROP TABLE IF EXISTS sudal.comparison_participants;