package postgres

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestNewFromDB_ExecutorBasics(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() // nolint:errcheck

	exec, txr := NewFromDB(db)
	require.NotNil(t, exec)
	require.NotNil(t, txr)

	// ExecContext
	mock.ExpectExec(`UPDATE test SET value = \$1 WHERE id = \$2`).
		WithArgs("v", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err = exec.ExecContext(context.Background(), "UPDATE test SET value = $1 WHERE id = $2", "v", 1)
	require.NoError(t, err)

	// QueryRowContext
	mock.ExpectQuery(`SELECT value FROM test WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("v"))

	row := exec.QueryRowContext(context.Background(), "SELECT value FROM test WHERE id = $1", 1)
	var got string
	require.NoError(t, row.Scan(&got))
	require.Equal(t, "v", got)

	// PrepareContext
	mock.ExpectPrepare(`INSERT INTO test\(id, value\) VALUES \(\$1, \$2\)`)
	stmt, err := exec.PrepareContext(context.Background(), "INSERT INTO test(id, value) VALUES ($1, $2)")
	require.NoError(t, err)
	_ = stmt.Close() // nolint:errcheck

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNewFromDB_Transactor(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() // nolint:errcheck

	_, txr := NewFromDB(db)
	require.NotNil(t, txr)

	mock.ExpectBegin()
	tx, err := txr.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	// Exec inside tx
	mock.ExpectExec(`UPDATE test SET value = \$1 WHERE id = \$2`).
		WithArgs("v2", 2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	_, err = tx.ExecContext(context.Background(), "UPDATE test SET value = $1 WHERE id = $2", "v2", 2)
	require.NoError(t, err)

	mock.ExpectCommit()
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNewFromDB_TransactorRollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() // nolint:errcheck

	_, txr := NewFromDB(db)
	require.NotNil(t, txr)

	mock.ExpectBegin()
	tx, err := txr.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	mock.ExpectRollback()
	require.NoError(t, tx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}
