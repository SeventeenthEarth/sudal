-- Migration: Create comparisons table
-- Description: Rollback comparisons table creation
-- Version: 000005
-- Direction: DOWN

-- Drop indexes first (in reverse order of creation)
DROP INDEX IF EXISTS sudal.idx_comparisons_created_at;
DROP INDEX IF EXISTS sudal.idx_comparisons_room_id;
DROP INDEX IF EXISTS sudal.idx_comparisons_quiz_set_id;

-- Drop the comparisons table
-- This will also drop all constraints and comments associated with the table
DROP TABLE IF EXISTS sudal.comparisons;