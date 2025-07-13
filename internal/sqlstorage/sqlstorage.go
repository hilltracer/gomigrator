package sqlstorage

import (
	"context"
	"hash/fnv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
)

const metaTableDDL = `
CREATE TABLE IF NOT EXISTS gomigrator_schema_migrations (
	version     BIGINT      PRIMARY KEY,
	name        TEXT        NOT NULL,
	is_applied  BOOLEAN     NOT NULL,
	applied_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);`

type Store struct {
	db     *sqlx.DB
	lockID int64 // pg_advisory_lock(key)
}

// NewWithMock is only for tests; allows injection of custom DB.
func NewWithMock(db *sqlx.DB, lockID int64) *Store {
	return &Store{
		db:     db,
		lockID: lockID,
	}
}

func Connect(ctx context.Context, dsn string) (*Store, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(time.Hour)

	s := &Store{
		db:     db,
		lockID: hashLockID("gomigrator"),
	}
	if err := s.ensureMetaTable(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}
func (s *Store) Close() error { return s.db.Close() }

// Take advisory-lock, start transaction, call fn and commit.
func (s *Store) WithExclusive(ctx context.Context, fn func(*sqlx.Tx) error) error {
	if err := s.acquireLock(ctx); err != nil {
		return err
	}
	defer s.releaseLock(ctx)

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// Return map[version]isApplied.
func (s *Store) AppliedVersions(ctx context.Context) (map[int64]bool, error) {
	rows, err := s.db.QueryxContext(ctx,
		`SELECT version, is_applied FROM gomigrator_schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[int64]bool)
	var v int64
	var applied bool
	for rows.Next() {
		if err := rows.Scan(&v, &applied); err != nil {
			return nil, err
		}
		res[v] = applied
	}
	return res, rows.Err()
}

// Add migration record.
func (s *Store) MarkApplied(ctx context.Context, tx *sqlx.Tx, version int64, name string) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO gomigrator_schema_migrations (version, name, is_applied)
		 VALUES ($1, $2, true)
		 ON CONFLICT (version) DO UPDATE SET is_applied = true, applied_at = now()`,
		version, name)
	return err
}

// Remove migration record.
func (s *Store) MarkRolledBack(ctx context.Context, tx *sqlx.Tx, version int64) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM gomigrator_schema_migrations WHERE version = $1`, version)
	return err
}

// Create meta table if not exists.
func (s *Store) ensureMetaTable(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, metaTableDDL)
	return err
}

// Manage advisory locks.
func (s *Store) acquireLock(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "SELECT pg_advisory_lock($1)", s.lockID)
	return err
}

func (s *Store) releaseLock(ctx context.Context) {
	_, _ = s.db.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", s.lockID)
}

// Take a hash of the key to use as an advisory lock ID.
func hashLockID(key string) int64 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return int64(h.Sum32())
}
