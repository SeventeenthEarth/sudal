package postgres

import (
	"errors"

	"go.uber.org/zap"

	ssql "github.com/seventeenthearth/sudal/internal/service/sql"
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
	// exec provides the minimal SQL execution surface
	exec ssql.Executor

	// logger provides structured logging for repository operations
	logger *zap.Logger
}

// NewRepository creates a new base repository instance with the provided Executor and logger.
// This constructor initializes the shared repository component that can be embedded
// in feature-specific repository implementations.
//
// Parameters:
//   - exec: Minimal SQL execution surface (usually backed by *sql.DB or *sql.Tx)
//   - logger: Structured logger for recording repository operations and errors
//
// Returns:
//   - *Repository: A new repository instance ready for embedding in concrete implementations
//
// Example Usage:
//
//	baseRepo := postgres.NewRepository(exec, logger)
//	userRepo := &userRepoImpl{Repository: baseRepo}
func NewRepository(exec ssql.Executor, logger *zap.Logger) *Repository {
	return &Repository{
		exec:   exec,
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
func (r *Repository) WithTx(tx ssql.Tx) *Repository {
	return &Repository{
		exec:   tx,
		logger: r.logger.With(zap.String("scope", "transaction")),
	}
}

// GetExecutor returns the minimal SQL execution surface for repositories
func (r *Repository) GetExecutor() ssql.Executor {
	return r.exec
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
