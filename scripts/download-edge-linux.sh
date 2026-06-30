#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
arch="${1:-amd64}"
version="3.0.0-1038"
tag="3.0"

case "$arch" in
  amd64)
    deb_name="n2n_${version}_amd64.deb"
    sha256="dd29694cf8c22487618218687c0b81961736ae1afe1c43bcfde6d48f017f16f6"
    ;;
  *)
    echo "unsupported linux arch: $arch" >&2
    echo "currently supported: amd64" >&2
    exit 2
    ;;
esac

url="https://github.com/ntop/n2n/releases/download/${tag}/${deb_name}"
cache_dir="${repo_root}/.cache/n2n"
deb_path="${cache_dir}/${deb_name}"
extract_dir="${cache_dir}/extract-${arch}"
edge_path="${repo_root}/bin/linux/${arch}/edge"

mkdir -p "$cache_dir" "$(dirname "$edge_path")"

if [[ ! -f "$deb_path" ]]; then
  curl -L --fail --retry 3 -o "$deb_path" "$url"
fi

actual="$(sha256sum "$deb_path" | awk '{print $1}')"
if [[ "$actual" != "$sha256" ]]; then
  echo "checksum mismatch for ${deb_name}" >&2
  echo "expected: ${sha256}" >&2
  echo "actual:   ${actual}" >&2
  exit 1
fi

rm -rf "$extract_dir"
mkdir -p "$extract_dir"
dpkg-deb -x "$deb_path" "$extract_dir"
install -m 0755 "${extract_dir}/usr/sbin/edge" "$edge_path"

"$edge_path" -h >/dev/null
echo "installed ${edge_path}"
