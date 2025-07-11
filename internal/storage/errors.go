package storage

import "errors"

var (
	ErrDateBusy = errors.New("date/time already busy by another event")
	ErrNotFound = errors.New("event not found")
)
