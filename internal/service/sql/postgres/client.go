package postgres

import (
	"context"
	stdsql "database/sql"

	ssql "github.com/seventeenthearth/sudal/internal/service/sql"
)

// NewFromDB creates an Executor and Transactor backed by the given *sql.DB.
// This provides a thin wrapper over database/sql without adding configuration
// or pool management concerns (handled by infra layer).
func NewFromDB(db *stdsql.DB) (ssql.Executor, ssql.Transactor) {
	exec := &dbExecutor{db: db}
	return exec, exec
}

// NewTx wraps an existing *sql.Tx as an ssql.Tx.
func NewTx(tx *stdsql.Tx) ssql.Tx {
	return &txExecutor{tx: tx}
}

type dbExecutor struct {
	db *stdsql.DB
}

func (d *dbExecutor) QueryContext(ctx context.Context, query string, args ...any) (*stdsql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

func (d *dbExecutor) QueryRowContext(ctx context.Context, query string, args ...any) *stdsql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

func (d *dbExecutor) ExecContext(ctx context.Context, query string, args ...any) (stdsql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

func (d *dbExecutor) PrepareContext(ctx context.Context, query string) (*stdsql.Stmt, error) {
	return d.db.PrepareContext(ctx, query)
}

func (d *dbExecutor) BeginTx(ctx context.Context, opts *stdsql.TxOptions) (ssql.Tx, error) {
	tx, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &txExecutor{tx: tx}, nil
}

type txExecutor struct {
	tx *stdsql.Tx
}

func (t *txExecutor) QueryContext(ctx context.Context, query string, args ...any) (*stdsql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *txExecutor) QueryRowContext(ctx context.Context, query string, args ...any) *stdsql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *txExecutor) ExecContext(ctx context.Context, query string, args ...any) (stdsql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *txExecutor) PrepareContext(ctx context.Context, query string) (*stdsql.Stmt, error) {
	return t.tx.PrepareContext(ctx, query)
}

func (t *txExecutor) Commit() error {
	return t.tx.Commit()
}

func (t *txExecutor) Rollback() error {
	return t.tx.Rollback()
}
