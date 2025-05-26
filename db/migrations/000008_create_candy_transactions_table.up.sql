-- Migration: Create candy_transactions table
-- Description: Virtual currency transactions for the Social Quiz Platform
-- Version: 000008
-- Direction: UP

-- Create the candy_transactions table in the sudal schema
-- This table records all transactions related to the virtual currency "candy"
CREATE TABLE sudal.candy_transactions (
    -- Primary key: BIGSERIAL for performance (general content, not security-sensitive)
    id BIGSERIAL PRIMARY KEY,

    -- Foreign key to users table (required)
    user_id UUID NOT NULL REFERENCES sudal.users(id) ON DELETE CASCADE,

    -- Transaction type (required)
    -- Examples: 'PURCHASED', 'EARNED_QUIZ_COMPLETION', 'USED_COMPARISON_FEE', 'CONTRIBUTED_TO_POT'
    type VARCHAR(50) NOT NULL,

    -- Transaction amount (required)
    -- Positive for credits (earning/purchasing), negative for debits (spending)
    amount INTEGER NOT NULL,

    -- User's candy balance after this transaction (required)
    -- This provides an audit trail and helps with balance verification
    balance_after_transaction INTEGER NOT NULL,

    -- Optional description of the transaction
    -- Using TEXT to accommodate longer descriptions
    description TEXT,

    -- Optional reference ID for linking to other entities
    -- Examples: comparison_id, room_id, external_payment_id
    reference_id VARCHAR(255),

    -- Timestamp when the transaction was created (UTC)
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance optimization
-- Index on user_id for filtering transactions by user
CREATE INDEX idx_candy_transactions_user_id ON sudal.candy_transactions(user_id);

-- Index on type for filtering transactions by type
CREATE INDEX idx_candy_transactions_type ON sudal.candy_transactions(type);

-- Index on created_at for sorting by transaction date
CREATE INDEX idx_candy_transactions_created_at ON sudal.candy_transactions(created_at);

-- Index on reference_id for linking transactions to other entities
CREATE INDEX idx_candy_transactions_reference_id ON sudal.candy_transactions(reference_id);

-- Composite index for user transaction history
CREATE INDEX idx_candy_transactions_user_created ON sudal.candy_transactions(user_id, created_at);

-- Add table and column comments for documentation
COMMENT ON TABLE sudal.candy_transactions IS 'Virtual currency (candy) transaction records';
COMMENT ON COLUMN sudal.candy_transactions.id IS 'BIGSERIAL-based unique transaction identifier';
COMMENT ON COLUMN sudal.candy_transactions.user_id IS 'Reference to the user involved in the transaction';
COMMENT ON COLUMN sudal.candy_transactions.type IS 'Type of transaction (PURCHASED, EARNED, USED, etc.)';
COMMENT ON COLUMN sudal.candy_transactions.amount IS 'Transaction amount (positive=credit, negative=debit)';
COMMENT ON COLUMN sudal.candy_transactions.balance_after_transaction IS 'User candy balance after this transaction';
COMMENT ON COLUMN sudal.candy_transactions.description IS 'Optional description of the transaction';
COMMENT ON COLUMN sudal.candy_transactions.reference_id IS 'Optional reference to related entity (comparison_id, etc.)';
COMMENT ON COLUMN sudal.candy_transactions.created_at IS 'Transaction timestamp (UTC)';

-- Add constraints for data validation
-- Ensure type is not empty
ALTER TABLE sudal.candy_transactions ADD CONSTRAINT chk_candy_transactions_type_not_empty
    CHECK (LENGTH(TRIM(type)) > 0);

-- Ensure amount is not zero (transactions must have an impact)
ALTER TABLE sudal.candy_transactions ADD CONSTRAINT chk_candy_transactions_amount_not_zero
    CHECK (amount != 0);

-- Ensure balance_after_transaction is not negative
ALTER TABLE sudal.candy_transactions ADD CONSTRAINT chk_candy_transactions_balance_non_negative
    CHECK (balance_after_transaction >= 0);