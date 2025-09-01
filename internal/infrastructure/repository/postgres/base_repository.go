package postgres

import (
	"context"
	stdsql "database/sql"
	"errors"

	"github.com/lib/pq"
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

// QueryRow executes a query that is expected to return at most one row and logs it at debug level.
// It wraps the underlying Executor's QueryRowContext.
func (r *Repository) QueryRow(ctx context.Context, query string, args ...any) *stdsql.Row { // nolint:ireturn
	if ce := r.logger.Check(zap.DebugLevel, "sql.query"); ce != nil {
		ce.Write(zap.String("query", query), zap.Any("args", args))
	}
	return r.exec.QueryRowContext(ctx, query, args...)
}

// Query executes a query that returns multiple rows and logs it at debug level.
func (r *Repository) Query(ctx context.Context, query string, args ...any) (*stdsql.Rows, error) {
	if ce := r.logger.Check(zap.DebugLevel, "sql.query"); ce != nil {
		ce.Write(zap.String("query", query), zap.Any("args", args))
	}
	rows, err := r.exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, mapPGError(err)
	}
	return rows, nil
}

// Exec executes a statement without returning any rows and logs it at debug level.
func (r *Repository) Exec(ctx context.Context, query string, args ...any) (stdsql.Result, error) { // nolint:ireturn
	if ce := r.logger.Check(zap.DebugLevel, "sql.exec"); ce != nil {
		ce.Write(zap.String("query", query), zap.Any("args", args))
	}
	res, err := r.exec.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, mapPGError(err)
	}
	return res, nil
}

// ScanOne scans a single row into the provided destination values and maps common errors.
func (r *Repository) ScanOne(row *stdsql.Row, dest ...any) error {
	if row == nil {
		return ErrInvalidInput
	}
	if err := row.Scan(dest...); err != nil {
		return mapPGError(err)
	}
	return nil
}

// mapPGError maps driver-specific errors to standardized repository errors.
// - sql.ErrNoRows -> ErrNotFound
// - pq errors: unique_violation -> ErrDuplicateEntry, others -> ErrConstraintViolation (limited set)
func mapPGError(err error) error { // nolint:cyclop
	if err == nil {
		return nil
	}
	if errors.Is(err, stdsql.ErrNoRows) {
		return ErrNotFound
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch string(pqErr.Code) {
		case "23505": // unique_violation
			return ErrDuplicateEntry
		case "23503": // foreign_key_violation
			return ErrConstraintViolation
		case "23502": // not_null_violation
			return ErrConstraintViolation
		case "23514": // check_violation
			return ErrConstraintViolation
		}
	}
	return err
}
