#!/usr/bin/env pwsh
# Dev environment setup for Windows (no make or POSIX shell required).
#
# Installs git hooks and fetches Go dependencies.
# On Linux/macOS use scripts/setup.sh instead.
#
# Usage:  pwsh scripts/setup.ps1   (or: powershell -ExecutionPolicy Bypass -File scripts/setup.ps1)
$ErrorActionPreference = 'Stop'

Set-Location (git rev-parse --show-toplevel)

# 1. Git hooks — copy into the repo's real hooks dir.
#    `git rev-parse --git-path hooks` resolves the correct location even for
#    linked worktrees. Copy-Item preserves the file's LF endings, so the hook
#    stays runnable under sh on Windows.
$hooksDir = git rev-parse --git-path hooks
New-Item -ItemType Directory -Force -Path $hooksDir | Out-Null
$dest = Join-Path $hooksDir 'pre-commit'
Copy-Item -Path 'scripts/git-hooks/pre-commit' -Destination $dest -Force
Write-Host "[setup] installed pre-commit hook -> $dest"

# 2. Go toolchain + module dependencies.
if (Get-Command go -ErrorAction SilentlyContinue) {
	Write-Host "[setup] $(go version)"
	go mod download
	Write-Host '[setup] go module dependencies ready'
} else {
	Write-Warning 'go not found on PATH - install Go (see go.mod)'
}

Write-Host '[setup] done.'
