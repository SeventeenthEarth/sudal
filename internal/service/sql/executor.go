package sql

import (
	"context"
	stdsql "database/sql"
)

// Executor defines the minimal surface for executing SQL operations.
// It mirrors the subset of methods commonly used from *sql.DB and *sql.Tx.
type Executor interface {
	QueryContext(ctx context.Context, query string, args ...any) (*stdsql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *stdsql.Row
	ExecContext(ctx context.Context, query string, args ...any) (stdsql.Result, error)
	PrepareContext(ctx context.Context, query string) (*stdsql.Stmt, error)
}

// Tx represents a transaction that can execute SQL operations and be committed/rolled back.
type Tx interface {
	Executor
	Commit() error
	Rollback() error
}

// Transactor can start a new transaction and return a transaction-aware executor.
type Transactor interface {
	BeginTx(ctx context.Context, opts *stdsql.TxOptions) (Tx, error)
}
