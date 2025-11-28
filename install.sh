#!/usr/bin/env bash
set -euo pipefail

REPO="dacrab/craftops"
NAME="craftops"

die() { echo "Error: $1" >&2; exit 1; }

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in x86_64|amd64) ARCH=amd64 ;; arm64|aarch64) ARCH=arm64 ;; *) die "Unsupported arch: $ARCH" ;; esac
[[ "$OS" == "linux" || "$OS" == "darwin" ]] || die "Unsupported OS: $OS"

VERSION=${VERSION:-$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep -o '"tag_name": "[^"]*' | cut -d'"' -f4)}
[[ -n "$VERSION" ]] || die "Failed to get latest version"

DEST="${HOME}/.local/bin"
[[ "$(id -u)" -eq 0 ]] && DEST="/usr/local/bin"
mkdir -p "$DEST"

URL="https://github.com/${REPO}/releases/download/${VERSION}/${NAME}-${OS}-${ARCH}"
echo "Installing ${NAME} ${VERSION} (${OS}/${ARCH}) to ${DEST}"

TMP=$(mktemp)
trap 'rm -f "$TMP"' EXIT
curl -fsSL -o "$TMP" "$URL" || die "Download failed"
chmod +x "$TMP"

if [[ -w "$DEST" ]]; then
  mv "$TMP" "${DEST}/${NAME}"
else
  sudo mv "$TMP" "${DEST}/${NAME}"
fi

echo "Installed: ${DEST}/${NAME}"
command -v "$NAME" >/dev/null || echo "Add to PATH: export PATH=\"${DEST}:\$PATH\""

# Create default config if missing
CFG="${HOME}/.config/craftops/config.toml"
if [[ ! -f "$CFG" ]]; then
  mkdir -p "$(dirname "$CFG")"
  "${DEST}/${NAME}" init-config -o "$CFG" 2>/dev/null && echo "Created: $CFG" || true
fi

echo "Done. Run: ${NAME} --help"
