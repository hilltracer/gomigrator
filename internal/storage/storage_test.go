package sqlstorage

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hilltracer/otus-go/hw12_13_14_15_calendar/internal/storage"
	"github.com/jmoiron/sqlx"
)

func mustEvent(id string, start time.Time, dur time.Duration) storage.Event {
	return storage.Event{ID: id, UserID: "u1", StartTime: start, Duration: dur}
}

func newMock() (*Storage, sqlmock.Sqlmock, func()) {
	db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	return New(sqlx.NewDb(db, "sqlmock")), mock, func() { _ = db.Close() }
}

func TestCreateEvent_Overlap(t *testing.T) {
	s, mock, cleanup := newMock()
	defer cleanup()

	ctx := context.Background()
	ev := mustEvent("1", time.Now(), time.Hour)

	// expect overlap query returning row => ErrDateBusy
	overlapRe := regexp.QuoteMeta(
		`SELECT true FROM events WHERE user_id=$1 AND start_time < $3 AND 
		(start_time + (duration * interval '1 microsecond') / 1000) > $2 LIMIT 1`)
	mock.ExpectBegin()
	mock.ExpectQuery(overlapRe).
		WithArgs(ev.UserID, ev.StartTime, ev.StartTime.Add(ev.Duration)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	if err := s.CreateEvent(ctx, ev); !errors.Is(err, storage.ErrDateBusy) {
		t.Fatalf("want ErrDateBusy, got %v", err)
	}
}

func TestDeleteEvent_NotFound(t *testing.T) {
	s, mock, cleanup := newMock()
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM events WHERE id=$1`)).
		WithArgs("42").
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := s.DeleteEvent(context.Background(), "42"); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestListDayWeekMonth_Mapping(t *testing.T) {
	s, mock, cleanup := newMock()
	defer cleanup()

	dayStart := time.Date(2025, 7, 2, 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)

	weekStart := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 7)

	monthStart := weekStart
	monthEnd := monthStart.AddDate(0, 1, 0)

	// Expected SQL
	query := regexp.QuoteMeta(`SELECT id, title, start_time, duration, description, user_id, notify_before
                     FROM events WHERE user_id=$1 AND start_time >= $2 AND start_time < $3
                     ORDER BY start_time`)

	// --- ListDay ---
	mock.ExpectQuery(query).
		WithArgs("u1", dayStart, dayEnd).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "start_time", "duration", "description", "user_id", "notify_before",
		}).AddRow("d1", "day event", dayStart, int64(3600000000000), "desc", "u1", int64(0)))

	dayEvents, err := s.ListDay(context.Background(), "u1", dayStart)
	if err != nil || len(dayEvents) != 1 || dayEvents[0].ID != "d1" {
		t.Fatalf("ListDay failed: %+v (%v)", dayEvents, err)
	}

	// --- ListWeek ---
	mock.ExpectQuery(query).
		WithArgs("u1", weekStart, weekEnd).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "start_time", "duration", "description", "user_id", "notify_before",
		}).AddRow("w1", "week event", weekStart.AddDate(0, 0, 2), int64(7200000000000), "desc", "u1", int64(0)))

	weekEvents, err := s.ListWeek(context.Background(), "u1", weekStart)
	if err != nil || len(weekEvents) != 1 || weekEvents[0].ID != "w1" {
		t.Fatalf("ListWeek failed: %+v (%v)", weekEvents, err)
	}

	// --- ListMonth ---
	mock.ExpectQuery(query).
		WithArgs("u1", monthStart, monthEnd).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "start_time", "duration", "description", "user_id", "notify_before",
		}).AddRow("m1", "month event", monthStart.AddDate(0, 0, 10), int64(1800000000000), "desc", "u1", int64(0)))

	monthEvents, err := s.ListMonth(context.Background(), "u1", monthStart)
	if err != nil || len(monthEvents) != 1 || monthEvents[0].ID != "m1" {
		t.Fatalf("ListMonth failed: %+v (%v)", monthEvents, err)
	}
}
