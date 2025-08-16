#!/usr/bin/env bash
# scripts/flow.sh
# Usage:
#   scripts/flow.sh                # installs latest
#   scripts/flow.sh v0.1.0         # installs specific tag
#   FLOW_VERSION=v0.1.0 scripts/flow.sh
# Env:
#   FLOW_REPO      (default: selimacerbas/flow)
#   FLOW_VERSION   (default: latest)
#   FLOW_BIN_PATH  (default: ./flow or ./flow.exe)
#   GITHUB_TOKEN   (optional; avoids API rate limiting)

set -euo pipefail

REPO="${FLOW_REPO:-selimacerbas/flow}"
REQUESTED="${1:-${FLOW_VERSION:-latest}}"

# --- detect os/arch ---
uname_s="$(uname -s | tr '[:upper:]' '[:lower:]')"
uname_m="$(uname -m)"

case "$uname_s" in
linux) OS=linux ;;
darwin) OS=darwin ;;
msys* | mingw* | cygwin*) OS=windows ;; # if running under Git Bash on Windows
*)
    echo "Unsupported OS: $uname_s" >&2
    exit 1
    ;;
esac

case "$uname_m" in
x86_64 | amd64) ARCH=amd64 ;;
aarch64 | arm64) ARCH=arm64 ;;
*)
    echo "Unsupported arch: $uname_m" >&2
    exit 1
    ;;
esac

EXT=""
[ "$OS" = "windows" ] && EXT=".exe"

OUT="${FLOW_BIN_PATH:-./flow$EXT}"

# --- resolve tag ---
get_latest_tag() {
    # Prefer gh if available, fall back to GitHub API
    if command -v gh >/dev/null 2>&1; then
        gh release view -R "$REPO" --json tagName -q .tagName
    else
        # Use token if available to avoid rate limits
        if [ -n "${GITHUB_TOKEN:-}" ]; then
            AUTH=(-H "authorization: Bearer ${GITHUB_TOKEN}")
        else
            AUTH=()
        fi
        curl -fsSL "${AUTH[@]}" "https://api.github.com/repos/${REPO}/releases/latest" |
            awk -F'"' '/"tag_name":/ {print $4; exit}'
    fi
}

if [ "$REQUESTED" = "latest" ]; then
    TAG="$(get_latest_tag)"
else
    # Accept "v0.1.0" or "0.1.0"
    if [[ "$REQUESTED" =~ ^v ]]; then
        TAG="$REQUESTED"
    else
        TAG="v$REQUESTED"
    fi
fi

if [ -z "${TAG:-}" ]; then
    echo "Failed to resolve release tag." >&2
    exit 1
fi

VER="${TAG#v}"

# Our releases are uploaded as *raw binaries* with names based on either:
#   1) {{ .Version }} -> flow_0.1.0_linux_amd64
#   2) (fallback) {{ .Tag }} -> flow_v0.1.0_linux_amd64
CANDIDATE_1="flow_${VER}_${OS}_${ARCH}${EXT}"
CANDIDATE_2="flow_${TAG}_${OS}_${ARCH}${EXT}"

download() {
    local name="$1"
    local url="https://github.com/${REPO}/releases/download/${TAG}/${name}"
    echo "Downloading $url"
    if command -v gh >/dev/null 2>&1; then
        gh release download -R "$REPO" "$TAG" --pattern "$name" --output "$name"
    else
        if [ -n "${GITHUB_TOKEN:-}" ]; then
            AUTH=(-H "authorization: Bearer ${GITHUB_TOKEN}")
        else
            AUTH=()
        fi
        curl -fL "${AUTH[@]}" -o "$name" "$url"
    fi
}

tmp="$(mktemp)"
trap 'rm -f "$tmp" "${CANDIDATE_1}" "${CANDIDATE_2}"' EXIT

# Try numeric versioned filename first, then tag-prefixed
if download "$CANDIDATE_1"; then
    mv "$CANDIDATE_1" "$tmp"
elif download "$CANDIDATE_2"; then
    mv "$CANDIDATE_2" "$tmp"
else
    echo "Could not find asset for $TAG ($OS/$ARCH)." >&2
    exit 1
fi

chmod +x "$tmp"
mv "$tmp" "$OUT"

echo "Installed to $OUT"
"$OUT" --version || true
