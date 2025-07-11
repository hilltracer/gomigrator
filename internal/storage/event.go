package storage

import "time"

type Event struct {
	ID           string        `db:"id"`
	Title        string        `db:"title"`
	StartTime    time.Time     `db:"start_time"`
	Duration     time.Duration `db:"duration"`
	Description  string        `db:"description"`
	UserID       string        `db:"user_id"`
	NotifyBefore time.Duration `db:"notify_before"`
}
