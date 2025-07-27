package migrator

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hilltracer/gomigrator/internal/sqlstorage"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// helper builds a Store backed by sqlmock and returns the store, mock and dir.
func helper(t *testing.T, upSQL, downSQL string, applied bool) (*Migrator, sqlmock.Sqlmock, func()) {
	t.Helper()

	// temp migration file
	dir := t.TempDir()
	version := time.Now().UTC().Format("20060102150405")
	file := filepath.Join(dir, version+"_demo.sql")
	mig := `-- +gomigrator Up
` + upSQL + `
-- +gomigrator Down
` + downSQL + `
`
	require.NoError(t, os.WriteFile(file, []byte(mig), 0o644))

	// mock DB
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	mock.MatchExpectationsInOrder(false)

	dbx := sqlx.NewDb(db, "gomigrator")

	store := sqlstorage.NewWithMock(dbx, 42)
	m := New(store, dir)

	// common expectation – list of applied migrations
	rows := sqlmock.NewRows([]string{"version", "is_applied"})
	if applied {
		rows.AddRow(version, true)
	}
	mock.ExpectQuery("SELECT version, is_applied FROM gomigrator_schema_migrations").
		WillReturnRows(rows)

	return m, mock, func() {
		require.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestIsExecutableSQL(t *testing.T) {
	require.False(t, isExecutableSQL("-- comment only"))
	require.False(t, isExecutableSQL("\n\t  "))
	require.True(t, isExecutableSQL("CREATE TABLE x(id INT);"))
}

func TestUp_AppliesMissingMigrations(t *testing.T) {
	const upSQL = `CREATE TABLE qwe(id INT);`
	m, mock, done := helper(t, upSQL, "DROP TABLE qwe;", false)
	defer done()

	mock.ExpectExec(`SELECT pg_advisory_lock\(\$1\)`).WithArgs(int64(42)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectBegin()
	mock.ExpectExec(`CREATE TABLE qwe\(id INT\);`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO gomigrator_schema_migrations").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectExec(`SELECT pg_advisory_unlock\(\$1\)`).WithArgs(int64(42)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, m.Up(context.Background()))
}

func TestDown_RollsBackLast(t *testing.T) {
	upSQL := `CREATE TABLE qwe(id INT);`
	downSQL := "DROP TABLE qwe;"
	m, mock, done := helper(t, upSQL, downSQL, true)
	defer done()

	mock.ExpectExec(`SELECT pg_advisory_lock\(\$1\)`).WithArgs(int64(42)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectBegin()
	mock.ExpectExec(`DROP TABLE qwe;`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM gomigrator_schema_migrations").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectExec(`SELECT pg_advisory_unlock\(\$1\)`).WithArgs(int64(42)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, m.Down(context.Background()))
}

func TestRedo_DownThenUp(t *testing.T) {
	upSQL := `CREATE TABLE qwe(id INT);`
	downSQL := "DROP TABLE qwe;"
	m, mock, done := helper(t, upSQL, downSQL, true)
	defer done()

	mock.ExpectExec(`SELECT pg_advisory_lock\(\$1\)`).WithArgs(int64(42)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectBegin()
	mock.ExpectExec(`DROP TABLE qwe;`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`CREATE TABLE qwe\(id INT\);`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO gomigrator_schema_migrations").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectExec(`SELECT pg_advisory_unlock\(\$1\)`).WithArgs(int64(42)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, m.Redo(context.Background()))
}
