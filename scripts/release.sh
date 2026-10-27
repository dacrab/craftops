#!/usr/bin/env bash
# Bump the semver tag and push it to trigger the GitHub Actions release.
# Usage: ./scripts/release.sh [patch|minor|major] ["optional release message"]
set -euo pipefail

VERSION_TYPE="${1:-patch}"
RELEASE_MESSAGE="${2:-}"

CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "0.0.0")
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"

case "$VERSION_TYPE" in
  major) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
  minor) MINOR=$((MINOR + 1)); PATCH=0 ;;
  patch) PATCH=$((PATCH + 1)) ;;
  *)
    printf 'Error: version type must be patch, minor, or major\n' >&2
    exit 1
    ;;
esac

NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"

# Use the curated CHANGELOG.md section for the release.
TMPFILE=$(mktemp)
trap 'rm -f "$TMPFILE"' EXIT

awk -v header="## ${NEW_VERSION}" '
  /^## / {
    if (index($0, header) == 1) { found = 1; print; next }
    if (found) exit
  }
  found { print }
' CHANGELOG.md > "$TMPFILE"

if [ ! -s "$TMPFILE" ]; then
  printf 'Error: add a "%s" section to CHANGELOG.md first.\n' "## ${NEW_VERSION}" >&2
  exit 1
fi

if [ -n "$RELEASE_MESSAGE" ]; then
  printf '\n### Notes\n\n%s\n' "$RELEASE_MESSAGE" >> "$TMPFILE"
fi

CHANGELOG=$(cat "$TMPFILE")

printf 'Current version : %s\n' "$CURRENT_VERSION"
printf 'New version     : %s\n' "$NEW_VERSION"
printf '\nChangelog:\n%s\n\n' "$CHANGELOG"
read -r -p "Create and push tag ${NEW_VERSION}? (y/N) " REPLY
printf '\n'

if [[ ! "$REPLY" =~ ^[Yy]$ ]]; then
  printf 'Cancelled\n'
  exit 1
fi

git tag -a "$NEW_VERSION" -m "$CHANGELOG"
git push origin "$NEW_VERSION"

printf '✅ Tag %s created and pushed!\n' "$NEW_VERSION"
printf 'GitHub Actions will now build and create the release automatically.\n'
