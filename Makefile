BINARY := bin/tossctl
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X github.com/JungHoonGhae/tossinvest-cli/internal/version.Version=$(VERSION) \
	-X github.com/JungHoonGhae/tossinvest-cli/internal/version.Commit=$(COMMIT) \
	-X github.com/JungHoonGhae/tossinvest-cli/internal/version.Date=$(DATE)

.PHONY: build run test fmt lint tidy clean hooks

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

# Install the local git hooks (pre-commit: gofmt + go vet + go test).
# Re-run after cloning. Bypass a single commit with: git commit --no-verify
hooks:
	cp scripts/git-hooks/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
	@echo "Installed .git/hooks/pre-commit"
