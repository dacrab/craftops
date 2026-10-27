#!/usr/bin/env bash
# Print the CHANGELOG.md section for a release tag.
# Usage: ./scripts/changelog.sh v2.8.2
set -euo pipefail

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
  printf 'Usage: %s <version>\n' "${BASH_SOURCE[0]}" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HEADER="## ${VERSION}"

NOTES="$(awk -v header="$HEADER" '
  /^## / {
    if (index($0, header) == 1 && (length($0) == length(header) || substr($0, length(header) + 1, 1) == " ")) {
      found = 1
      print
      next
    }
    if (found) exit
  }
  found { print }
' "${ROOT}/CHANGELOG.md")"

if [ -z "$NOTES" ]; then
  printf 'Error: add a "%s" section to CHANGELOG.md first.\n' "$HEADER" >&2
  exit 1
fi

printf '%s\n' "$NOTES"
