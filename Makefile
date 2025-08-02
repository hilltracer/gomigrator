BIN := "./bin/gomigrator"
DOCKER_IMG="gomigrator:develop"

GIT_HASH   := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -X 'main.release=develop' \
              -X 'main.buildDate=$(BUILD_DATE)' \
              -X 'main.gitHash=$(GIT_HASH)'

## ---------- build ----------
build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/gomigrator

run: build
	$(BIN) -config ./configs/config.yaml

## ---------- tests & lint ----------
install-lint-deps:
	(which golangci-lint > /dev/null) || \
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
	sh -s -- -b $(shell go env GOPATH)/bin v1.63.4

lint: install-lint-deps
	golangci-lint run ./...

test:
	go test -race -count=100 ./internal/creator ./internal/migrator ./internal/parser


## ---------- integration tests inside docker ----------
integration-test: build
	@./scripts/integration_test.sh

.PHONY: build run build-img run-img lint test integration-test 
