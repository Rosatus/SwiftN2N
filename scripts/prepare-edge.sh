#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
target_os="${1:-$(go env GOOS)}"
target_arch="${2:-$(go env GOARCH)}"

case "${target_os}/${target_arch}" in
  linux/amd64)
    "${repo_root}/scripts/download-edge-linux.sh" amd64
    ;;
  windows/amd64|darwin/amd64|darwin/arm64)
    "${repo_root}/scripts/build-edge-from-source.sh" "$target_os" "$target_arch"
    ;;
  *)
    echo "unsupported edge target: ${target_os}/${target_arch}" >&2
    echo "supported: linux/amd64, windows/amd64, darwin/amd64, darwin/arm64" >&2
    exit 2
    ;;
esac
