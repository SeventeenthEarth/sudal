package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/seventeenthearth/sudal/internal/feature/candy_transaction/domain/entity"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../../mocks/mock_candy_transaction_repository.go -package=mocks -mock_names=CandyTransactionRepository=MockCandyTransactionRepository github.com/seventeenthearth/sudal/internal/feature/candy_transaction/domain/repo CandyTransactionRepository

// CandyTransactionRepository defines the interface for candy transaction data access operations
// This interface abstracts the data layer and supports both PostgreSQL and Redis implementations
// following the Repository Pattern to maintain clean separation between domain and data layers.
//
// Implementation Strategy:
// - Write Operations (Create): Write to PostgreSQL first, then update/invalidate Redis cache
// - Read Operations: Attempt Redis first (cache), fallback to PostgreSQL on cache miss
// - Balance Operations: Use Redis for real-time balance tracking with PostgreSQL as source of truth
//
// Error Handling:
// - Use domain-specific sentinel errors (entity.ErrCandyTransactionNotFound, entity.ErrCandyTransactionInsufficientFunds, etc.)
// - Wrap infrastructure errors with appropriate context
// - Ensure consistent error types across different implementations
//
// Transaction Safety:
// - All balance-affecting operations must be atomic
// - Use database transactions for consistency
// - Implement optimistic locking for concurrent balance updates
type CandyTransactionRepository interface {
	// Create creates a new candy transaction record
	// This method handles both credit and debit transactions
	// The transaction amount and type must be consistent (credits positive, debits negative)
	// Returns entity.ErrCandyTransactionInsufficientFunds if user doesn't have enough candy for debits
	// Returns entity.ErrCandyTransactionInvalidType if transaction type doesn't match amount sign
	// Returns entity.ErrCandyTransactionDuplicateReference if reference ID is already used
	Create(ctx context.Context, transaction *entity.CandyTransaction) (*entity.CandyTransaction, error)

	// GetByID retrieves a candy transaction by its unique ID
	// Returns entity.ErrCandyTransactionNotFound if no transaction exists with the given ID
	// Returns entity.ErrCandyTransactionInvalidID if the provided ID is invalid
	GetByID(ctx context.Context, id int64) (*entity.CandyTransaction, error)

	// GetByUserID retrieves all candy transactions for a specific user
	// page: page number (1-based)
	// pageSize: maximum number of transactions to return per page
	// transactionType: optional filter by transaction type (empty string means no filtering)
	// Returns transactions ordered by created_at DESC, total count, and error
	GetByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int, transactionType string) ([]*entity.CandyTransaction, int64, error)

	// GetByReferenceID retrieves all transactions linked to a specific reference ID
	// This is useful for finding all transactions related to a comparison, purchase, etc.
	// Returns an empty slice if no transactions are found (not an error)
	GetByReferenceID(ctx context.Context, referenceID string) ([]*entity.CandyTransaction, error)

	// GetByDateRange retrieves transactions within a specific date range
	// userID: optional user filter (uuid.Nil means all users)
	// startDate: start of the date range (inclusive)
	// endDate: end of the date range (inclusive)
	// page: page number (1-based)
	// pageSize: maximum number of transactions to return per page
	// Returns transactions ordered by created_at DESC, total count, and error
	GetByDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time, page, pageSize int) ([]*entity.CandyTransaction, int64, error)

	// GetByType retrieves transactions by transaction type
	// transactionType: the type of transactions to retrieve
	// page: page number (1-based)
	// pageSize: maximum number of transactions to return per page
	// Returns transactions ordered by created_at DESC, total count, and error
	GetByType(ctx context.Context, transactionType entity.CandyTransactionType, page, pageSize int) ([]*entity.CandyTransaction, int64, error)

	// Update updates an existing transaction's information
	// Note: Updating transactions should be rare and carefully controlled
	// Returns entity.ErrCandyTransactionNotFound if no transaction exists with the given ID
	Update(ctx context.Context, transaction *entity.CandyTransaction) (*entity.CandyTransaction, error)

	// Delete removes a transaction from the system
	// Note: Deleting transactions should be rare and carefully controlled
	// Returns entity.ErrCandyTransactionNotFound if no transaction exists with the given ID
	Delete(ctx context.Context, id int64) error

	// DeleteByUserID removes all transactions for a specific user
	// This is useful for user account deletion
	// Returns the number of transactions deleted
	DeleteByUserID(ctx context.Context, userID uuid.UUID) (int64, error)

	// List retrieves a paginated list of all transactions
	// page: page number (1-based)
	// pageSize: maximum number of transactions to return per page
	// Returns transactions ordered by created_at DESC, total count, and error
	List(ctx context.Context, page, pageSize int) ([]*entity.CandyTransaction, int64, error)

	// Count returns the total number of candy transactions in the system
	// This is useful for analytics and administrative dashboards
	Count(ctx context.Context) (int64, error)

	// CountByUserID returns the total number of transactions for a specific user
	// This is useful for user analytics and pagination
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)

	// CountByType returns the number of transactions for a specific type
	// This is useful for analytics and transaction type popularity
	CountByType(ctx context.Context, transactionType entity.CandyTransactionType) (int64, error)

	// Exists checks if a transaction exists with the given ID
	// Returns true if the transaction exists, false otherwise
	// Returns an error only if there's a system/database error
	Exists(ctx context.Context, id int64) (bool, error)

	// ExistsByReferenceID checks if a transaction exists with the given reference ID
	// Returns true if the transaction exists, false otherwise
	// Returns an error only if there's a system/database error
	ExistsByReferenceID(ctx context.Context, referenceID string) (bool, error)

	// GetUserBalance calculates the current candy balance for a user
	// This should return the most up-to-date balance based on transaction history
	GetUserBalance(ctx context.Context, userID uuid.UUID) (int, error)

	// GetRecentTransactions retrieves recently created transactions
	// hours: number of hours to look back from now
	// page: page number (1-based)
	// pageSize: maximum number of transactions to return per page
	// Returns transactions ordered by created_at DESC, total count, and error
	GetRecentTransactions(ctx context.Context, hours int, page, pageSize int) ([]*entity.CandyTransaction, int64, error)

	// GetTransactionStats retrieves system-wide transaction statistics
	// startDate: optional start date for the stats period (nil means all time)
	// endDate: optional end date for the stats period (nil means all time)
	// Returns total transactions, total candy issued, total candy spent, and error
	GetTransactionStats(ctx context.Context, startDate, endDate *time.Time) (totalTransactions int64, totalIssued int64, totalSpent int64, err error)

	// GetTopUsers retrieves users with the highest candy balances
	// limit: maximum number of users to return
	// Returns user IDs and their balances ordered by balance DESC
	GetTopUsers(ctx context.Context, limit int) ([]uuid.UUID, []int, error)

	// GetTransactionsByTypeStats retrieves transactions grouped by type within a date range
	// startDate: start of the date range (inclusive)
	// endDate: end of the date range (inclusive)
	// Returns map of transaction type to count and total amount
	GetTransactionsByTypeStats(ctx context.Context, startDate, endDate time.Time) (map[entity.CandyTransactionType]entity.CandyTransactionTypeStats, error)

	// ValidateUserBalance verifies that a user's calculated balance matches their stored balance
	// This is useful for audit and integrity checks
	// Returns true if balances match, false otherwise
	// Returns an error only if there's a system/database error
	ValidateUserBalance(ctx context.Context, userID uuid.UUID) (bool, error)

	// GetActiveUsers retrieves users who have made transactions recently
	// hours: number of hours to look back from now
	// limit: maximum number of users to return
	// Returns user IDs ordered by most recent transaction
	GetActiveUsers(ctx context.Context, hours int, limit int) ([]uuid.UUID, error)
}
