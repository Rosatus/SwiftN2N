#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
target_os="${1:-$(go env GOOS 2>/dev/null || uname -s | tr '[:upper:]' '[:lower:]')}"
target_arch="${2:-$(go env GOARCH 2>/dev/null || uname -m)}"
n2n_ref="${N2N_REF:-3.0-stable}"
n2n_repo="${N2N_REPO:-https://github.com/ntop/n2n.git}"

case "$target_os" in
  windows)
    edge_name="edge.exe"
    ;;
  darwin|linux)
    edge_name="edge"
    ;;
  *)
    echo "unsupported target os: ${target_os}" >&2
    exit 2
    ;;
esac

cache_dir="${repo_root}/.cache/n2n-src/${n2n_ref}"
build_dir="${repo_root}/.cache/n2n-build/${target_os}-${target_arch}"
edge_out="${repo_root}/bin/${target_os}/${target_arch}/${edge_name}"

mkdir -p "$(dirname "$cache_dir")" "$(dirname "$edge_out")"

if [[ ! -d "${cache_dir}/.git" ]]; then
  rm -rf "$cache_dir"
  git clone --depth 1 --branch "$n2n_ref" "$n2n_repo" "$cache_dir"
else
  git -C "$cache_dir" fetch --depth 1 origin "$n2n_ref"
  git -C "$cache_dir" checkout "$n2n_ref"
  git -C "$cache_dir" reset --hard "origin/${n2n_ref}"
fi

rm -rf "$build_dir"
cmake -S "$cache_dir" -B "$build_dir" \
  -DCMAKE_BUILD_TYPE=Release \
  -DN2N_OPTION_USE_OPENSSL=OFF \
  -DN2N_OPTION_USE_PCAPLIB=OFF \
  -DN2N_OPTION_USE_ZSTD=OFF
cmake --build "$build_dir" --config Release --target edge --parallel

candidate=""
for path in \
  "${build_dir}/edge" \
  "${build_dir}/edge.exe" \
  "${build_dir}/Release/edge" \
  "${build_dir}/Release/edge.exe" \
  "${build_dir}/src/edge" \
  "${build_dir}/src/edge.exe"; do
  if [[ -f "$path" ]]; then
    candidate="$path"
    break
  fi
done

if [[ -z "$candidate" ]]; then
  echo "built edge binary not found under ${build_dir}" >&2
  find "$build_dir" -maxdepth 3 -type f -name 'edge*' -print >&2 || true
  exit 1
fi

install -m 0755 "$candidate" "$edge_out"
"$edge_out" -h >/dev/null
echo "installed ${edge_out}"
