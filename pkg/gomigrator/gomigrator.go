package gomigrator

import (
	"context"

	core "github.com/hilltracer/gomigrator/internal/migrator"
)

// Describes the minimum set of parameters necessary for connecting
// to the base and the operation of the migrator.
type Config struct {
	DSN string // Postgres connection line
	Dir string // Dir with SQL migration files
}

// Describes the status of a migration.
type StatusEntry struct {
	Version   int64
	IsApplied bool
}

// Allows to use, roll back and check migrations.
// Safe for multi-flow use, provided that each operation is
// in its own Migrator copy.
type Migrator struct{ m *core.Migrator }

// Open connection to the database and return a Migrator instance.
func New(ctx context.Context, cfg Config) (*Migrator, error) {
	m, err := core.NewFromDSN(ctx, cfg.DSN, cfg.Dir)
	if err != nil {
		return nil, err
	}
	return &Migrator{m: m}, nil
}

// Closes the connection to the database.
func (m *Migrator) Close() error { return m.m.Close() }

// Applies all migrations that have not yet been applied.
func (m *Migrator) Up(ctx context.Context) error { return m.m.Up(ctx) }

// Rolls back the latest applied migration.
func (m *Migrator) Down(ctx context.Context) error { return m.m.Down(ctx) }

// Redo = Down + Up of the last migration, in a single transaction.
func (m *Migrator) Redo(ctx context.Context) error { return m.m.Redo(ctx) }

// Returns sorted migration statuses.
func (m *Migrator) Status(ctx context.Context) ([]StatusEntry, error) {
	internalStatuses, err := m.m.Status(ctx)
	if err != nil {
		return nil, err
	}
	statuses := make([]StatusEntry, len(internalStatuses))
	for i, s := range internalStatuses {
		statuses[i] = StatusEntry{
			Version:   s.Version,
			IsApplied: s.IsApplied,
		}
	}
	return statuses, nil
}

// Returns the highest applied version or 0 if none.
func (m *Migrator) DBVersion(ctx context.Context) (int64, error) {
	return m.m.DBVersion(ctx)
}
