#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <output-path>" >&2
  exit 1
fi

: "${VERSION:?VERSION is required}"
: "${ARM_SHA:?ARM_SHA is required}"
: "${X86_SHA:?X86_SHA is required}"
: "${LINUX_ARM_SHA:?LINUX_ARM_SHA is required}"
: "${LINUX_X86_SHA:?LINUX_X86_SHA is required}"
: "${REPO:?REPO is required}"
: "${TAG:?TAG is required}"

OUTPUT_PATH="$1"
mkdir -p "$(dirname "$OUTPUT_PATH")"

cat > "$OUTPUT_PATH" <<FORMULA
class Tossctl < Formula
  desc "Unofficial CLI for Toss Securities web workflows"
  homepage "https://github.com/${REPO}"
  version "${VERSION}"
  license "MIT"

  depends_on "python@3.11"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/${REPO}/releases/download/${TAG}/tossctl-darwin-arm64.tar.gz"
      sha256 "${ARM_SHA}"
    else
      url "https://github.com/${REPO}/releases/download/${TAG}/tossctl-darwin-amd64.tar.gz"
      sha256 "${X86_SHA}"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/${REPO}/releases/download/${TAG}/tossctl-linux-arm64.tar.gz"
      sha256 "${LINUX_ARM_SHA}"
    else
      url "https://github.com/${REPO}/releases/download/${TAG}/tossctl-linux-amd64.tar.gz"
      sha256 "${LINUX_X86_SHA}"
    end
  end

  def install
    libexec.install "tossctl"
    libexec.install "auth-helper"

    env = {
      "TOSSCTL_AUTH_HELPER_DIR" => libexec/"auth-helper",
      "TOSSCTL_AUTH_HELPER_PYTHON" => Formula["python@3.11"].opt_bin/"python3.11",
    }
    (bin/"tossctl").write_env_script libexec/"tossctl", env
  end

  test do
    assert_match "tossctl", shell_output("#{bin}/tossctl version")
  end
end
FORMULA
