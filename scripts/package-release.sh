#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
target_os="${1:-$(go env GOOS)}"
target_arch="${2:-$(go env GOARCH)}"
version="${3:-dev}"

app_name="SwiftN2N"
stage_dir="${repo_root}/dist/${app_name}-${target_os}-${target_arch}"
archive_base="${app_name}-${version}-${target_os}-${target_arch}"

rm -rf "$stage_dir"
mkdir -p "$stage_dir"

copy_notices() {
  for file in LICENSE THIRD_PARTY_NOTICES.md README.md CHANGELOG.md; do
    if [[ -f "${repo_root}/${file}" ]]; then
      cp -a "${repo_root}/${file}" "$stage_dir/"
    fi
  done
  VERSION="$version" node "${repo_root}/scripts/generate-compliance-reports.mjs" "$stage_dir" >/dev/null
}

write_checksum() {
  local archive="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    (cd "$(dirname "$archive")" && sha256sum "$(basename "$archive")" > "$(basename "$archive").sha256")
  else
    (cd "$(dirname "$archive")" && shasum -a 256 "$(basename "$archive")" > "$(basename "$archive").sha256")
  fi
}

case "$target_os" in
  linux)
    edge_src="${repo_root}/bin/linux/${target_arch}/edge"
    edge_dst="${repo_root}/build/bin/bin/linux/${target_arch}/edge"
    mkdir -p "$(dirname "$edge_dst")"
    install -m 0755 "$edge_src" "$edge_dst"
    cp -a "${repo_root}/build/bin/${app_name}" "${repo_root}/build/bin/bin" "$stage_dir/"
    copy_notices
    archive="${repo_root}/dist/${archive_base}.tar.gz"
    (cd "${repo_root}/dist" && tar -czf "${archive_base}.tar.gz" "${app_name}-${target_os}-${target_arch}")
    write_checksum "$archive"
    echo "$archive"
    ;;
  windows)
    edge_src="${repo_root}/bin/windows/${target_arch}/edge.exe"
    edge_dst="${repo_root}/build/bin/bin/windows/${target_arch}/edge.exe"
    mkdir -p "$(dirname "$edge_dst")"
    install -m 0755 "$edge_src" "$edge_dst"
    cp -a "${repo_root}/build/bin/${app_name}.exe" "${repo_root}/build/bin/bin" "$stage_dir/"
    copy_notices
    python_bin="python3"
    if ! command -v "$python_bin" >/dev/null 2>&1; then
      python_bin="python"
    fi
    "$python_bin" - "$repo_root" "$stage_dir" "${repo_root}/dist/${archive_base}.zip" <<'PY'
import pathlib
import sys
import zipfile

repo = pathlib.Path(sys.argv[1])
stage = pathlib.Path(sys.argv[2])
archive = pathlib.Path(sys.argv[3])
with zipfile.ZipFile(archive, "w", zipfile.ZIP_DEFLATED) as zf:
    for path in stage.rglob("*"):
        zf.write(path, path.relative_to(stage.parent))
print(archive)
PY
    write_checksum "${repo_root}/dist/${archive_base}.zip"
    ;;
  darwin)
    edge_src="${repo_root}/bin/darwin/${target_arch}/edge"
    app_bundle="${repo_root}/build/bin/${app_name}.app"
    edge_dst="${app_bundle}/Contents/MacOS/bin/darwin/${target_arch}/edge"
    mkdir -p "$(dirname "$edge_dst")"
    install -m 0755 "$edge_src" "$edge_dst"
    cp -a "$app_bundle" "$stage_dir/"
    copy_notices
    archive="${repo_root}/dist/${archive_base}.tar.gz"
    (cd "${repo_root}/dist" && tar -czf "${archive_base}.tar.gz" "${app_name}-${target_os}-${target_arch}")
    write_checksum "$archive"
    echo "$archive"
    ;;
  *)
    echo "unsupported target os: ${target_os}" >&2
    exit 2
    ;;
esac
