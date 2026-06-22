#!/bin/sh
# Dev environment setup — installs git hooks and fetches Go dependencies.
#
# Cross-platform: Linux, macOS, and Windows (Git Bash / WSL).
# Windows users without a POSIX shell: run scripts/setup.ps1 instead.
#
# Usage:  sh scripts/setup.sh   (or: make setup)
set -eu

cd "$(git rev-parse --show-toplevel)"

# 1. Git hooks — copy into the repo's real hooks dir.
#    `git rev-parse --git-path hooks` resolves the correct location even for
#    linked worktrees, so this works regardless of where .git lives.
hooks_dir=$(git rev-parse --git-path hooks)
mkdir -p "$hooks_dir"
cp scripts/git-hooks/pre-commit "$hooks_dir/pre-commit"
chmod +x "$hooks_dir/pre-commit" 2>/dev/null || true
echo "[setup] installed pre-commit hook -> $hooks_dir/pre-commit"

# 2. Go toolchain + module dependencies.
if command -v go >/dev/null 2>&1; then
	echo "[setup] $(go version)"
	go mod download
	echo "[setup] go module dependencies ready"
else
	echo "[setup] WARNING: go not found on PATH — install Go (see go.mod)" >&2
fi

echo "[setup] done."
