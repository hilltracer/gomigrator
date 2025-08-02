package migrator

import (
	"context"
	"sort"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hilltracer/gomigrator/internal/sqlstorage"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestStatus_ReturnsSortedVersionsAndMap(t *testing.T) {
	ctx := context.Background()

	// mock DB
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	dbx := sqlx.NewDb(db, "gomigrator")

	store := sqlstorage.NewWithMock(dbx, 42)
	m := New(store, t.TempDir())

	// unsorted rows → Status must sort them
	vOld := StatusEntry{int64(20240102030405), true}
	vNew := StatusEntry{int64(20250102030405), true}

	rows := sqlmock.
		NewRows([]string{"version", "is_applied"}).
		AddRow(vNew.Version, vNew.IsApplied).
		AddRow(vOld.Version, vOld.IsApplied)

	mock.ExpectQuery("SELECT version, is_applied FROM gomigrator_schema_migrations").
		WillReturnRows(rows)

	statuses, err := m.Status(ctx)
	require.NoError(t, err)

	require.True(t, sort.SliceIsSorted(statuses, func(i, j int) bool { return statuses[i].Version < statuses[j].Version }))
	require.Equal(t, []StatusEntry{vOld, vNew}, statuses)
	require.True(t, vNew.IsApplied)
	require.True(t, vNew.IsApplied)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDBVersion_ReturnsHighestAppliedVersion(t *testing.T) {
	ctx := context.Background()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	dbx := sqlx.NewDb(db, "gomigrator")

	store := sqlstorage.NewWithMock(dbx, 42)
	m := New(store, t.TempDir())

	highestApplied := int64(20250102030405)
	higherButNotApplied := int64(20260102030405)

	rows := sqlmock.
		NewRows([]string{"version", "is_applied"}).
		AddRow(highestApplied, true).
		AddRow(higherButNotApplied, false)

	mock.ExpectQuery("SELECT version, is_applied FROM gomigrator_schema_migrations").
		WillReturnRows(rows)

	got, err := m.DBVersion(ctx)
	require.NoError(t, err)
	require.Equal(t, highestApplied, got)

	require.NoError(t, mock.ExpectationsWereMet())
}
