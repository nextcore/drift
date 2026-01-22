#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DRIFT_CACHE_ROOT="${DRIFT_CACHE_DIR:-$HOME/.drift}"
REPO="go-drift/drift"

detect_drift_version() {
  if [[ -n "${DRIFT_VERSION:-}" ]]; then
    echo "$DRIFT_VERSION"
    return
  fi

  local base
  base="$(basename "$ROOT_DIR")"
  if [[ "$base" == *@* ]]; then
    echo "${base##*@}"
    return
  fi

  if git -C "$ROOT_DIR" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    local tag
    tag=$(git -C "$ROOT_DIR" describe --tags --abbrev=0 2>/dev/null || true)
    if [[ -n "$tag" ]]; then
      if [[ "$tag" == drift-* ]]; then
        echo "${tag#drift-}"
      else
        echo "$tag"
      fi
      return
    fi
  fi

  echo ""
}

usage() {
  cat <<EOF
Usage: $(basename "$0") [--android] [--ios]

Downloads prebuilt Drift Skia artifacts from GitHub Releases and extracts them
into the versioned cache directory.

Output: \$DRIFT_CACHE_DIR/lib/<version>/<platform>/<arch>/libdrift_skia.a
        (defaults to ~/.drift/lib/<version>/...)

Environment:
  DRIFT_CACHE_DIR   Override cache directory (default: ~/.drift)
  DRIFT_VERSION     Override version detection

If no platform flags are provided, both are fetched.
EOF
}

platforms=()
for arg in "$@"; do
  case "$arg" in
    --android)
      platforms+=("android")
      ;;
    --ios)
      platforms+=("ios")
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $arg" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ ${#platforms[@]} -eq 0 ]]; then
  platforms=("android" "ios")
fi

drift_version="$(detect_drift_version)"
if [[ -z "$drift_version" ]]; then
  echo "Unable to determine Drift version. Set DRIFT_VERSION or run from the module cache." >&2
  exit 1
fi

# Normalize version to match CLI cache paths (see cache.normalizeVersion)
# Strip "drift-" prefix, "-dev" suffix, ensure "v" prefix
drift_version="${drift_version#drift-}"
drift_version="${drift_version%-dev}"
if [[ "$drift_version" != v* ]]; then
  drift_version="v$drift_version"
fi

release_tag="$drift_version"
base_url="https://github.com/$REPO/releases/download/$release_tag"
manifest_url="$base_url/manifest.json"

DRIFT_SKIA_OUT="$DRIFT_CACHE_ROOT/lib/$drift_version"

work_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$work_dir"
}
trap cleanup EXIT

echo "Fetching Drift Skia $drift_version..."

manifest="$work_dir/manifest.json"
echo "  Downloading manifest..."
curl -fsSL "$manifest_url" -o "$manifest"

extract_tarball() {
  local platform="$1"
  local tarball="drift-$drift_version-$platform.tar.gz"
  local tar_path="$work_dir/$tarball"
  local expected_sha

  expected_sha=$(python3 - <<PY
import json
with open("$manifest", "r", encoding="utf-8") as f:
    data = json.load(f)
print(data["$platform"]["sha256"])
PY
)

  echo "  Downloading $tarball..."
  curl -fsSL "$base_url/$tarball" -o "$tar_path"
  local actual_sha
  actual_sha=$(sha256sum "$tar_path" | cut -d ' ' -f1)

  if [[ "$expected_sha" != "$actual_sha" ]]; then
    echo "Checksum mismatch for $tarball" >&2
    echo "Expected: $expected_sha" >&2
    echo "Actual:   $actual_sha" >&2
    exit 1
  fi

  echo "  Extracting $platform..."
  mkdir -p "$DRIFT_SKIA_OUT"
  tar -C "$DRIFT_SKIA_OUT" -xzf "$tar_path"
}

for platform in "${platforms[@]}"; do
  extract_tarball "$platform"
done

echo "Drift Skia artifacts extracted to $DRIFT_SKIA_OUT"
