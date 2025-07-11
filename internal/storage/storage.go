package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/hilltracer/otus-go/hw12_13_14_15_calendar/internal/storage"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
)

type Storage struct {
	db *sqlx.DB
}

type Repository interface {
	CreateEvent(ctx context.Context, e Event) error
	UpdateEvent(ctx context.Context, e Event) error
	DeleteEvent(ctx context.Context, id string) error

	ListDay(ctx context.Context, userID string, date time.Time) ([]Event, error)
	ListWeek(ctx context.Context, userID string, weekStart time.Time) ([]Event, error)
	ListMonth(ctx context.Context, userID string, monthStart time.Time) ([]Event, error)
}

func New(db *sqlx.DB) *Storage { return &Storage{db: db} }

func Connect(ctx context.Context, dsn string) (*Storage, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(time.Hour)
	return &Storage{db: db}, nil
}

func (s *Storage) Close(_ context.Context) error { return s.db.Close() }

func (s *Storage) CreateEvent(ctx context.Context, e storage.Event) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // safe if already committed

	// overlap check
	var exists bool
	end := e.StartTime.Add(e.Duration)
	queryOverlap := `SELECT true FROM events WHERE user_id=$1 AND start_time < $3 AND
        (start_time + (duration * interval '1 microsecond') / 1000) > $2 LIMIT 1`
	if err = tx.QueryRowContext(ctx, queryOverlap, e.UserID, e.StartTime, end).Scan(&exists); err != nil &&
		!errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if exists {
		return storage.ErrDateBusy
	}

	// insert
	insert := `INSERT INTO events
        (id, title, start_time, duration, description, user_id, notify_before)
        VALUES (:id, :title, :start_time, :duration, :description, :user_id, :notify_before)`
	if _, err = tx.NamedExecContext(ctx, insert, map[string]any{
		"id":            e.ID,
		"title":         e.Title,
		"start_time":    e.StartTime,
		"duration":      e.Duration,
		"description":   e.Description,
		"user_id":       e.UserID,
		"notify_before": e.NotifyBefore,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Storage) UpdateEvent(ctx context.Context, e storage.Event) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// exists?
	var origCount int
	if err = tx.GetContext(ctx, &origCount, `SELECT 1 FROM events WHERE id=$1`, e.ID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return storage.ErrNotFound
		}
		return err
	}

	// overlap check (exclude self)
	end := e.StartTime.Add(e.Duration)
	var exists bool
	queryOverlap := `SELECT true FROM events WHERE user_id=$1 AND id <> $4 AND start_time < $3 AND
	(start_time + (duration * interval '1 microsecond') / 1000) > $2 LIMIT 1`
	if err = tx.QueryRowContext(ctx, queryOverlap, e.UserID, e.StartTime, end, e.ID).Scan(&exists); err != nil &&
		!errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if exists {
		return storage.ErrDateBusy
	}

	// update
	upd := `UPDATE events
        SET title=:title, start_time=:start_time, duration=:duration,
			description=:description, user_id=:user_id, notify_before=:notify_before
        WHERE id=:id`
	if _, err = tx.NamedExecContext(ctx, upd, map[string]any{
		"id":            e.ID,
		"title":         e.Title,
		"start_time":    e.StartTime,
		"duration":      e.Duration,
		"description":   e.Description,
		"user_id":       e.UserID,
		"notify_before": e.NotifyBefore,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM events WHERE id=$1`, id)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return storage.ErrNotFound
	}
	return nil
}

const baseSelect = `SELECT id, title, start_time, duration, description, user_id, notify_before
                    FROM events WHERE user_id=$1 AND start_time >= $2 AND start_time < $3
                    ORDER BY start_time`

func (s *Storage) ListDay(ctx context.Context, userID string, date time.Time) ([]storage.Event, error) {
	from := date.Truncate(24 * time.Hour)
	to := from.Add(24 * time.Hour)
	var out []storage.Event
	if err := s.db.SelectContext(ctx, &out, baseSelect, userID, from, to); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Storage) ListWeek(ctx context.Context, userID string, weekStart time.Time) ([]storage.Event, error) {
	from := weekStart.Truncate(24 * time.Hour)
	to := from.AddDate(0, 0, 7)
	var out []storage.Event
	if err := s.db.SelectContext(ctx, &out, baseSelect, userID, from, to); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Storage) ListMonth(ctx context.Context, userID string, monthStart time.Time) ([]storage.Event, error) {
	from := time.Date(monthStart.Year(), monthStart.Month(), 1, 0, 0, 0, 0, monthStart.Location())
	to := from.AddDate(0, 1, 0)
	var out []storage.Event
	if err := s.db.SelectContext(ctx, &out, baseSelect, userID, from, to); err != nil {
		return nil, err
	}
	return out, nil
}
