package app

import (
	"context"
	"time"

	"github.com/hilltracer/otus-go/hw12_13_14_15_calendar/internal/storage"
)

type App struct {
	logger Logger
	store  storage.Repository
}

type Logger interface {
	Info(string)
	Error(string)
}

func New(logger Logger, storage storage.Repository) *App {
	return &App{logger: logger, store: storage}
}

func (a *App) CreateEvent(ctx context.Context, id, title string) error {
	e := storage.Event{ID: id, Title: title}
	return a.store.CreateEvent(ctx, e)
}

func (a *App) CreateFullEvent(ctx context.Context, e storage.Event) error {
	return a.store.CreateEvent(ctx, e)
}

func (a *App) UpdateEvent(ctx context.Context, e storage.Event) error {
	return a.store.UpdateEvent(ctx, e)
}

func (a *App) DeleteEvent(ctx context.Context, id string) error {
	return a.store.DeleteEvent(ctx, id)
}

func (a *App) ListDay(ctx context.Context, userID string, date time.Time) ([]storage.Event, error) {
	return a.store.ListDay(ctx, userID, date)
}

func (a *App) ListWeek(ctx context.Context, userID string, weekStart time.Time) ([]storage.Event, error) {
	return a.store.ListWeek(ctx, userID, weekStart)
}

func (a *App) ListMonth(ctx context.Context, userID string, monthStart time.Time) ([]storage.Event, error) {
	return a.store.ListMonth(ctx, userID, monthStart)
}
