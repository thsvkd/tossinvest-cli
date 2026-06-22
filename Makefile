BINARY := bin/tossctl
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X github.com/JungHoonGhae/tossinvest-cli/internal/version.Version=$(VERSION) \
	-X github.com/JungHoonGhae/tossinvest-cli/internal/version.Commit=$(COMMIT) \
	-X github.com/JungHoonGhae/tossinvest-cli/internal/version.Date=$(DATE)

.PHONY: build run test fmt lint tidy clean setup hooks

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/tossctl

run:
	go run -ldflags "$(LDFLAGS)" ./cmd/tossctl

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

lint:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -rf bin coverage.out

# Dev environment setup: install git hooks (pre-commit: gofmt + go vet + go
# test) and fetch Go dependencies. Re-run after cloning. `hooks` is an alias.
# Windows without make/bash: run `pwsh scripts/setup.ps1` instead.
# Bypass a single commit with: git commit --no-verify
setup hooks:
	sh scripts/setup.sh
