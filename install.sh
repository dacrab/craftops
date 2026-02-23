#!/usr/bin/env bash
# install.sh — download and install the latest craftops release binary.
# Usage: curl -fsSL https://raw.githubusercontent.com/dacrab/craftops/main/install.sh | bash
#        VERSION=v2.3.0 bash install.sh   # pin a specific version
set -euo pipefail

REPO="dacrab/craftops"
NAME="craftops"

# Install to user bin by default; /usr/local/bin when running as root.
if [ "$(id -u)" -eq 0 ]; then
  DEST="/usr/local/bin"
else
  DEST="${HOME}/.local/bin"
fi

die() { printf 'Error: %s\n' "$1" >&2; exit 1; }

# ---- OS / arch detection -----------------------------------------------------
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux|darwin) ;;
  *) die "Unsupported OS: $OS" ;;
esac

ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64)  ARCH=amd64 ;;
  arm64|aarch64) ARCH=arm64 ;;
  *)             die "Unsupported architecture: $ARCH" ;;
esac

# ---- Resolve version ---------------------------------------------------------
if [ -z "${VERSION:-}" ]; then
  printf 'Checking latest version...\n'
  # Parse tag_name from GitHub API response without relying on jq
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | head -1 \
    | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
fi
[ -n "$VERSION" ] || die "Failed to determine latest version"

# ---- Download ----------------------------------------------------------------
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
BINARY_URL="${BASE_URL}/${NAME}-${OS}-${ARCH}"
SUMS_URL="${BASE_URL}/SHA256SUMS"

printf 'Installing %s %s (%s/%s) → %s\n' "$NAME" "$VERSION" "$OS" "$ARCH" "$DEST"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

TMPBIN="${TMPDIR}/${NAME}"
TMPSUMS="${TMPDIR}/SHA256SUMS"

curl -fsSL -o "$TMPBIN"   "$BINARY_URL" || die "Download failed: $BINARY_URL"
curl -fsSL -o "$TMPSUMS"  "$SUMS_URL"   || die "Download failed: $SUMS_URL"

# ---- Verify checksum ---------------------------------------------------------
EXPECTED=$(grep "${NAME}-${OS}-${ARCH}" "$TMPSUMS" | awk '{print $1}')
[ -n "$EXPECTED" ] || die "Checksum not found for ${NAME}-${OS}-${ARCH}"

if command -v sha256sum > /dev/null 2>&1; then
  ACTUAL=$(sha256sum "$TMPBIN" | awk '{print $1}')
elif command -v shasum > /dev/null 2>&1; then
  ACTUAL=$(shasum -a 256 "$TMPBIN" | awk '{print $1}')
else
  die "Neither sha256sum nor shasum found — cannot verify download"
fi

[ "$ACTUAL" = "$EXPECTED" ] || die "Checksum mismatch (expected $EXPECTED, got $ACTUAL)"

# ---- Install -----------------------------------------------------------------
chmod +x "$TMPBIN"
mkdir -p "$DEST"

if [ -w "$DEST" ]; then
  mv "$TMPBIN" "${DEST}/${NAME}"
else
  sudo mv "$TMPBIN" "${DEST}/${NAME}"
fi

printf 'Installed to %s/%s\n' "$DEST" "$NAME"

# ---- Warn if DEST is not in PATH ---------------------------------------------
case ":${PATH}:" in
  *":${DEST}:"*) ;;
  *) printf "Warning: %s is not in your PATH. Add it with:\n  export PATH=\"%s:\$PATH\"\n" "$DEST" "$DEST" ;;
esac

# ---- Optional: create default config -----------------------------------------
CFG="${HOME}/.config/craftops/config.toml"
if [ ! -f "$CFG" ]; then
  printf 'Creating default config at %s...\n' "$CFG"
  mkdir -p "$(dirname "$CFG")"
  "${DEST}/${NAME}" init-config -o "$CFG" > /dev/null 2>&1 || true
fi

printf 'Done. Run '\''%s --help'\'' to get started.\n' "$NAME"
