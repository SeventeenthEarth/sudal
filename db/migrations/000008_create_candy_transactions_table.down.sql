-- Migration: Create candy_transactions table
-- Description: Rollback candy_transactions table creation
-- Version: 000008
-- Direction: DOWN

-- Drop indexes first (in reverse order of creation)
DROP INDEX IF EXISTS sudal.idx_candy_transactions_user_created;
DROP INDEX IF EXISTS sudal.idx_candy_transactions_reference_id;
DROP INDEX IF EXISTS sudal.idx_candy_transactions_created_at;
DROP INDEX IF EXISTS sudal.idx_candy_transactions_type;
DROP INDEX IF EXISTS sudal.idx_candy_transactions_user_id;

-- Drop the candy_transactions table
-- This will also drop all constraints and comments associated with the table
DROP TABLE IF EXISTS sudal.candy_transactions;