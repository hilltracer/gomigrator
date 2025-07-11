-- +goose Up
CREATE TABLE IF NOT EXISTS events (
    id            UUID PRIMARY KEY,
    title         TEXT        NOT NULL,
    start_time    TIMESTAMP   NOT NULL,
    duration      BIGINT      NOT NULL,
    description   TEXT,
    user_id       TEXT        NOT NULL,
    notify_before BIGINT      DEFAULT '0'
);

CREATE INDEX IF NOT EXISTS idx_events_user_time ON events (user_id, start_time);

-- +goose Down
DROP TABLE events;
