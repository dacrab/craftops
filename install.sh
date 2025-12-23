#!/usr/bin/env bash
set -euo pipefail

REPO="dacrab/craftops"
NAME="craftops"
DEST="${HOME}/.local/bin"
[[ "$(id -u)" -eq 0 ]] && DEST="/usr/local/bin"

die() { echo "Error: $1" >&2; exit 1; }

# OS and Arch detection
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in 
    x86_64|amd64)   ARCH=amd64 ;; 
    arm64|aarch64)  ARCH=arm64 ;; 
    *)              die "Unsupported architecture: $ARCH" ;; 
esac

# Version fetching
echo "Checking latest version..."
VERSION=${VERSION:-$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | sed -n 's/.*"tag_name": "\(.*\)".*/\1/p')}
[[ -n "$VERSION" ]] || die "Failed to determine version"

# Download and Install
URL="https://github.com/${REPO}/releases/download/${VERSION}/${NAME}-${OS}-${ARCH}"
echo "Installing ${NAME} ${VERSION} (${OS}/${ARCH}) to ${DEST}"

TMP=$(mktemp)
trap 'rm -f "$TMP"' EXIT
curl -fsSL -o "$TMP" "$URL" || die "Download failed from $URL"
chmod +x "$TMP"

if [[ -w "$DEST" ]]; then
    mv "$TMP" "${DEST}/${NAME}"
else
    sudo mv "$TMP" "${DEST}/${NAME}"
fi

echo "Successfully installed to ${DEST}/${NAME}"

# Optional post-install setup
CFG="${HOME}/.config/craftops/config.toml"
if [[ ! -f "$CFG" ]]; then
    echo "Creating default configuration..."
    mkdir -p "$(dirname "$CFG")"
    "${DEST}/${NAME}" init-config -o "$CFG" &>/dev/null || true
fi

echo "Done. Run '${NAME} --help' to get started."
