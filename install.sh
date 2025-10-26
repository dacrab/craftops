#!/usr/bin/env bash
set -Eeuo pipefail

REPO="dacrab/craftops"
NAME="craftops"

color() { printf '%b%s%b\n' "$1" "$2" "\033[0m"; }
info() { color "\033[0;34m" "$1"; }
ok() { color "\033[0;32m" "$1"; }
warn() { color "\033[1;33m" "$1"; }
err() { color "\033[0;31m" "$1"; }

need() { command -v "$1" >/dev/null 2>&1 || { err "$1 is required"; exit 1; }; }

detect_os() {
  case "$(uname -s | tr '[:upper:]' '[:lower:]')" in
    linux) echo linux ;;
    darwin) echo darwin ;;
    *) err "unsupported OS"; exit 1;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo amd64 ;;
    arm64|aarch64) echo arm64 ;;
    *) err "unsupported architecture"; exit 1;;
  esac
}

latest_tag() {
  need curl
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | \
    sed -n 's/  "tag_name": "\([^"]\+\)".*/\1/p' | head -n1
}

install_dir() {
  if [ "${EUID:-$(id -u)}" -eq 0 ]; then echo /usr/local/bin; else echo "$HOME/.local/bin"; fi
}

main() {
  local OS ARCH VERSION ASSET URL TMP DEST
  OS=$(detect_os); ARCH=$(detect_arch)
  VERSION=${VERSION:-"$(latest_tag)"}
  [ -n "$VERSION" ] || { err "failed to resolve version"; exit 1; }
  ASSET="${NAME}-${OS}-${ARCH}"
  URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"
  DEST=$(install_dir)

  info "Installing ${NAME} ${VERSION} for ${OS}/${ARCH}"
  info "Destination: ${DEST}"
  mkdir -p "${DEST}"

  TMP=$(mktemp)
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "${TMP}" "${URL}"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "${TMP}" "${URL}"
  else
    err "curl or wget required"
    exit 1
  fi

  [ -s "${TMP}" ] || { err "download failed"; rm -f "${TMP}"; exit 1; }
  chmod +x "${TMP}"

  if [ -w "${DEST}" ]; then
    mv "${TMP}" "${DEST}/${NAME}"
  else
    need sudo
    sudo mv "${TMP}" "${DEST}/${NAME}"
  fi
  ok "Installed ${DEST}/${NAME}"

  if ! command -v "${NAME}" >/dev/null 2>&1; then
    warn "${DEST} may not be on PATH. Add: export PATH=\"${DEST}:\$PATH\""
  fi

  # Initialize config if missing
  local CFG_DIR CFG
  CFG_DIR="$HOME/.config/craftops"; CFG="${CFG_DIR}/config.toml"
  if [ ! -f "${CFG}" ]; then
    mkdir -p "${CFG_DIR}"
    "${DEST}/${NAME}" init-config -o "${CFG}" >/dev/null 2>&1 || true
    if [ -f "${CFG}" ]; then
      ok "Created ${CFG}"
    else
      warn "Could not create default config"
    fi
  fi

  ok "Done. Try: ${NAME} --version && ${NAME} health-check"
}

main "$@"