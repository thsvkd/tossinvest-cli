#!/usr/bin/env pwsh
# Dev environment setup for Windows (no make or POSIX shell required).
#
# Installs git hooks and fetches Go dependencies.
# On Linux/macOS use scripts/setup.sh instead.
#
# Usage:  pwsh scripts/setup.ps1   (or: powershell -ExecutionPolicy Bypass -File scripts/setup.ps1)
$ErrorActionPreference = 'Stop'

Set-Location (git rev-parse --show-toplevel)

# Print instructions for installing Go, preferring a package manager that is
# actually present, with the version required by go.mod when available.
function Show-GoInstallHelp {
	$req = $null
	if (Test-Path go.mod) {
		$m = Select-String -Path go.mod -Pattern '^go\s+(\S+)' | Select-Object -First 1
		if ($m) { $req = $m.Matches[0].Groups[1].Value }
	}
	$suffix = if ($req) { " (>= $req, per go.mod)" } else { '' }
	Write-Warning "Go is required$suffix but was not found on PATH."
	Write-Host '[setup] Install it with one of:'
	$shown = $false
	if (Get-Command winget -ErrorAction SilentlyContinue) { Write-Host '         winget install --id GoLang.Go'; $shown = $true }
	if (Get-Command choco  -ErrorAction SilentlyContinue) { Write-Host '         choco install golang';        $shown = $true }
	if (Get-Command scoop  -ErrorAction SilentlyContinue) { Write-Host '         scoop install go';            $shown = $true }
	if (-not $shown) {
		Write-Host '         winget install --id GoLang.Go   (or: choco install golang / scoop install go)'
	}
	Write-Host '[setup] Then re-run this script. Manual downloads: https://go.dev/dl/'
}

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
	Show-GoInstallHelp
	exit 1
}

Write-Host '[setup] done.'
