package migrator

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hilltracer/gomigrator/internal/parser"
	"github.com/hilltracer/gomigrator/internal/sqlstorage"
	"github.com/jmoiron/sqlx"
)

// Migrator owns the lifecycle of a sqlstorage.Store.
type Migrator struct {
	store *sqlstorage.Store
	dir   string
}

// New creates a Migrator from an *already-opened* Store (keeps old tests intact).
func New(store *sqlstorage.Store, dir string) *Migrator {
	return &Migrator{store: store, dir: dir}
}

// NewFromDSN is handy for CLI and integration-tests.
func NewFromDSN(ctx context.Context, dsn, dir string) (*Migrator, error) {
	store, err := sqlstorage.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &Migrator{store: store, dir: dir}, nil
}

func (m *Migrator) Close() error { return m.store.Close() }

func isExecutableSQL(sql string) bool {
	lines := strings.Split(sql, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		// line is not empty and doesn't start with comment: assume real SQL
		return true
	}
	return false
}

// Up applies every {is_applied = false} migration.
func (m *Migrator) Up(ctx context.Context) error {
	all, err := parser.ParseDir(m.dir)
	if err != nil {
		return err
	}
	applied, err := m.store.AppliedVersions(ctx)
	if err != nil {
		return err
	}

	return m.store.WithExclusive(ctx, func(tx *sqlx.Tx) error {
		for _, mig := range all {
			if applied[mig.Version] { // already done
				continue
			}
			if !isExecutableSQL(mig.UpSQL) {
				return fmt.Errorf("%s has empty Up block", mig.Name)
			}
			if _, err := tx.ExecContext(ctx, mig.UpSQL); err != nil {
				return fmt.Errorf("up %s: %w", mig.Name, err)
			}
			if err := m.store.MarkApplied(ctx, tx, mig.Version, mig.Name); err != nil {
				return err
			}
		}
		return nil
	})
}

// lastAppliedMigration returns the highest-applied migration file.
// If no migration was applied yet, it returns (nil, nil).
func (m *Migrator) lastAppliedMigration(ctx context.Context) (*parser.Migration, error) {
	applied, err := m.store.AppliedVersions(ctx)
	if err != nil {
		return nil, err
	}
	var last int64
	for v := range applied {
		if v > last {
			last = v
		}
	}
	if last == 0 {
		return nil, nil
	}
	all, err := parser.ParseDir(m.dir)
	if err != nil {
		return nil, err
	}
	for _, mig := range all {
		if mig.Version == last {
			return &mig, nil
		}
	}
	return nil, fmt.Errorf("migration file for version %d not found", last)
}

// Down rolls back the latest applied migration.
func (m *Migrator) Down(ctx context.Context) error {
	mig, err := m.lastAppliedMigration(ctx)
	if err != nil {
		return err
	}
	if mig == nil {
		return nil // nothing applied yet
	}
	if !isExecutableSQL(mig.DownSQL) {
		return fmt.Errorf("%s has empty Down block (cannot rollback)", mig.Name)
	}

	return m.store.WithExclusive(ctx, func(tx *sqlx.Tx) error {
		if _, err := tx.ExecContext(ctx, mig.DownSQL); err != nil {
			return fmt.Errorf("down %s: %w", mig.Name, err)
		}
		return m.store.MarkRolledBack(ctx, tx, mig.Version)
	})
}

// Redo = Down + Up of the last migration, in a single transaction.
func (m *Migrator) Redo(ctx context.Context) error {
	mig, err := m.lastAppliedMigration(ctx)
	if err != nil {
		return err
	}
	if mig == nil {
		return nil // nothing to redo
	}
	if !isExecutableSQL(mig.UpSQL) || !isExecutableSQL(mig.DownSQL) {
		return fmt.Errorf("%s must have both Up and Down blocks for redo", mig.Name)
	}

	return m.store.WithExclusive(ctx, func(tx *sqlx.Tx) error {
		if _, err := tx.ExecContext(ctx, mig.DownSQL); err != nil {
			return fmt.Errorf("redo-down %s: %w", mig.Name, err)
		}
		if _, err := tx.ExecContext(ctx, mig.UpSQL); err != nil {
			return fmt.Errorf("redo-up %s: %w", mig.Name, err)
		}
		return m.store.MarkApplied(ctx, tx, mig.Version, mig.Name)
	})
}

func sortedStatus(applied map[int64]bool) []int64 {
	vers := make([]int64, 0, len(applied))
	for v := range applied {
		vers = append(vers, v)
	}
	sort.Slice(vers, func(i, j int) bool { return vers[i] < vers[j] })
	return vers
}

// Status reports version ⇒ is_applied, sorted by version for stable output.
func (m *Migrator) Status(ctx context.Context) ([]int64, map[int64]bool, error) {
	applied, err := m.store.AppliedVersions(ctx)
	if err != nil {
		return nil, nil, err
	}
	return sortedStatus(applied), applied, nil
}

// DBVersion returns the highest applied version or 0 if none.
func (m *Migrator) DBVersion(ctx context.Context) (int64, error) {
	applied, err := m.store.AppliedVersions(ctx)
	if err != nil {
		return 0, err
	}
	var last int64
	for v, ok := range applied {
		if ok && v > last {
			last = v
		}
	}
	return last, nil
}
