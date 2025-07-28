# gomigrator

A lightweight PostgreSQL schema migration tool written in Go.

Inspired by tools like Goose, gomigrator lets you version, apply, and roll back SQL migrations safely in multi‑instance deployments.

## Features

* PostgreSQL support
* Plain SQL migrations with `-- +gomigrator Up/Down` sections
* Safe concurrent execution via `pg_advisory_lock`
* CLI and embeddable Go API (`pkg/gomigrator`)
* Configuration through YAML, flags, or environment variables (`${VAR}` expansion)
* Commands: `create`, `up`, `down`, `redo`, `status`, `dbversion`

## Installation

```bash
go install github.com/hilltracer/gomigrator/cmd/gomigrator@latest
```

This puts the binary in `$GOPATH/bin` (usually `~/go/bin`).

## Quick start

```bash
# Environment used by configs/config.yaml
export PG_HOST=localhost
export PG_PORT=5432
export PG_USER=postgres
export PG_PASSWORD=postgres
export PG_DB=postgres
export PG_SSLMODE=disable

export LOG_LEVEL=debug
```

### Check migration status using a config file

```bash
gomigrator --config configs/config.yaml --dir migrations status
```

### Apply all pending migrations

```bash
gomigrator --dir migrations up
```

### Override log level

```bash
gomigrator --log-level debug --dir migrations status
```

### Use a direct DSN instead of a config file

```bash
gomigrator --log-level debug \
"host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable" \
--dir migrations \
status
```

### Generate a new migration stub

```bash
gomigrator --dir ./migrations create init
```

## Command reference

| Command            | Purpose                                              |
| ------------------ | ---------------------------------------------------- |
| `create <name>`    | Generate `<timestamp>_<name>.sql` with Up/Down stubs |
| `up`               | Apply all pending migrations                         |
| `down`             | Roll back the last applied migration                 |
| `redo`             | `down` then `up` of the last migration               |
| `status`           | Print table of versions & applied state              |
| `dbversion`        | Show the highest applied version                     |
| `help` / `version` | Show CLI help or binary version                      |

## Build & test locally

```bash
make build     # build ./bin/gomigrator with version info
make test      # run unit tests with race detector
make lint      # golangci‑lint with the bundled config
```

