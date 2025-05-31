package postgres

import (
	"context"
	"database/sql"
	"errors"

	"go.uber.org/zap"
)

// Standardized PostgreSQL repository errors
// These errors provide consistent error handling across all PostgreSQL-based repository implementations
var (
	// ErrNotFound is returned when a requested entity is not found in the database
	ErrNotFound = errors.New("entity not found")

	// ErrDuplicateEntry is returned when attempting to create an entity that already exists
	ErrDuplicateEntry = errors.New("duplicate entry")

	// ErrConstraintViolation is returned when a database constraint is violated
	ErrConstraintViolation = errors.New("constraint violation")

	// ErrConnectionFailed is returned when database connection fails
	ErrConnectionFailed = errors.New("database connection failed")

	// ErrTransactionFailed is returned when a database transaction fails
	ErrTransactionFailed = errors.New("transaction failed")

	// ErrInvalidInput is returned when invalid input is provided to a repository method
	ErrInvalidInput = errors.New("invalid input")
)

// Repository provides shared PostgreSQL functionality for all repository implementations
// This base component encapsulates common database operations and transaction management
// to ensure consistency across all feature-specific repository implementations.
//
// Key Features:
// - Database connection management
// - Transaction scoping support
// - Standardized error handling
// - Structured logging integration
//
// Usage Pattern:
// Repository implementations should embed this struct and use its methods
// for common database operations while implementing feature-specific logic.
type Repository struct {
	// db holds the database connection pool or transaction
	// This can be either *sql.DB (for regular operations) or *sql.Tx (for transactional operations)
	db interface {
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
		QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
		PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	}

	// logger provides structured logging for repository operations
	logger *zap.Logger
}

// NewRepository creates a new base repository instance with the provided database connection and logger
// This constructor initializes the shared repository component that can be embedded
// in feature-specific repository implementations.
//
// Parameters:
//   - db: The database connection pool (*sql.DB) for executing queries
//   - logger: Structured logger for recording repository operations and errors
//
// Returns:
//   - *Repository: A new repository instance ready for embedding in concrete implementations
//
// Example Usage:
//
//	baseRepo := postgres.NewRepository(dbConnection, logger)
//	userRepo := &userRepoImpl{Repository: baseRepo}
func NewRepository(db *sql.DB, logger *zap.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger.With(zap.String("component", "postgres_repository")),
	}
}

// WithTx returns a new Repository instance scoped to the provided transaction
// This method enables transaction management at the service layer by allowing
// repository operations to be executed within a specific transaction context.
//
// The service layer is responsible for:
// - Beginning the transaction
// - Creating transaction-scoped repository instances using this method
// - Committing or rolling back the transaction based on operation results
//
// Parameters:
//   - tx: The database transaction (*sql.Tx) to scope this repository to
//
// Returns:
//   - *Repository: A new repository instance scoped to the provided transaction
//
// Example Usage:
//
//	tx, err := db.BeginTx(ctx, nil)
//	if err != nil {
//	    return err
//	}
//	defer tx.Rollback() // Will be ignored if tx.Commit() is called
//
//	txRepo := baseRepo.WithTx(tx)
//	// Use txRepo for all operations within this transaction
//	err = txRepo.SomeOperation(ctx, data)
//	if err != nil {
//	    return err // Transaction will be rolled back
//	}
//
//	return tx.Commit()
func (r *Repository) WithTx(tx *sql.Tx) *Repository {
	return &Repository{
		db:     tx,
		logger: r.logger.With(zap.String("scope", "transaction")),
	}
}

// GetDB returns the underlying database connection or transaction
// This method provides access to the raw database interface for operations
// that require direct database access beyond the standard repository patterns.
//
// Returns:
//   - interface{}: The database connection (*sql.DB) or transaction (*sql.Tx)
//
// Note: This method should be used sparingly and only when the standard
// repository patterns are insufficient for the required operation.
func (r *Repository) GetDB() interface{} {
	return r.db
}

// GetLogger returns the repository's logger instance
// This provides access to the structured logger for custom logging needs
// in concrete repository implementations.
//
// Returns:
//   - *zap.Logger: The structured logger configured for this repository
func (r *Repository) GetLogger() *zap.Logger {
	return r.logger
}
