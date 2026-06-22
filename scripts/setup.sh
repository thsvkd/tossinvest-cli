#!/bin/sh
# Dev environment setup — installs git hooks and fetches Go dependencies.
#
# Cross-platform: Linux, macOS, and Windows (Git Bash / WSL).
# Windows users without a POSIX shell: run scripts/setup.ps1 instead.
#
# Usage:  sh scripts/setup.sh   (or: make setup)
set -eu

cd "$(git rev-parse --show-toplevel)"

# Print OS-aware instructions for installing Go (detected via `uname`), with
# the version required by go.mod when available.
print_go_install_help() {
	req=$(awk '/^go /{print $2; exit}' go.mod 2>/dev/null || true)
	echo "[setup] Go is required${req:+ (>= $req, per go.mod)} but was not found on PATH." >&2
	echo "[setup] Install it with one of:" >&2
	case "$(uname -s 2>/dev/null || echo unknown)" in
	Darwin)
		echo "         brew install go                       # Homebrew" >&2
		;;
	Linux)
		if command -v apt-get >/dev/null 2>&1; then
			echo "         sudo apt-get install golang-go        # Debian/Ubuntu" >&2
		elif command -v dnf >/dev/null 2>&1; then
			echo "         sudo dnf install golang               # Fedora/RHEL" >&2
		elif command -v pacman >/dev/null 2>&1; then
			echo "         sudo pacman -S go                     # Arch" >&2
		elif command -v zypper >/dev/null 2>&1; then
			echo "         sudo zypper install go                # openSUSE" >&2
		elif command -v apk >/dev/null 2>&1; then
			echo "         sudo apk add go                       # Alpine" >&2
		fi
		echo "         # distro packages can lag; for the latest see https://go.dev/dl/" >&2
		;;
	MINGW* | MSYS* | CYGWIN*)
		echo "         winget install --id GoLang.Go         # or: choco install golang / scoop install go" >&2
		echo "         # on native Windows, scripts/setup.ps1 is recommended" >&2
		;;
	*)
		echo "         see https://go.dev/dl/" >&2
		;;
	esac
	echo "[setup] Then re-run this script. Manual downloads: https://go.dev/dl/" >&2
}

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
	print_go_install_help
	exit 1
fi

echo "[setup] done."
