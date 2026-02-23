#!/usr/bin/env bash
# release.sh — bump the semver tag and push it to trigger the GitHub Actions release.
# Usage: ./scripts/release.sh [patch|minor|major] ["optional release message"]
set -euo pipefail

VERSION_TYPE="${1:-patch}"
RELEASE_MESSAGE="${2:-}"

# ---- Determine current version -----------------------------------------------
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "0.0.0")
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"

# ---- Bump --------------------------------------------------------------------
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
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || true)

# ---- Build changelog into a temp file (avoids subshell variable loss) --------
TMPFILE=$(mktemp)
trap 'rm -f "$TMPFILE"' EXIT

printf '## What'\''s New\n\n' >> "$TMPFILE"
if [ -n "$RELEASE_MESSAGE" ]; then
  printf '%s\n\n' "$RELEASE_MESSAGE" >> "$TMPFILE"
fi

if [ -n "$LAST_TAG" ]; then
  printf '### Changes since %s\n\n' "$LAST_TAG" >> "$TMPFILE"
  git log "${LAST_TAG}..HEAD" --oneline --no-decorate | while IFS= read -r line; do
    printf -- '- %s\n' "$line" >> "$TMPFILE"
  done
else
  printf '### Changes since initial release\n\n' >> "$TMPFILE"
  git log --oneline --no-decorate | while IFS= read -r line; do
    printf -- '- %s\n' "$line" >> "$TMPFILE"
  done
fi

CHANGELOG=$(cat "$TMPFILE")

# ---- Confirm -----------------------------------------------------------------
printf 'Current version : %s\n' "$CURRENT_VERSION"
printf 'New version     : %s\n' "$NEW_VERSION"
printf '\nChangelog:\n%s\n\n' "$CHANGELOG"
read -r -p "Create and push tag ${NEW_VERSION}? (y/N) " REPLY
printf '\n'

if [[ ! "$REPLY" =~ ^[Yy]$ ]]; then
  printf 'Cancelled\n'
  exit 1
fi

# ---- Tag and push ------------------------------------------------------------
git tag -a "$NEW_VERSION" -m "$CHANGELOG"
git push origin "$NEW_VERSION"

printf '✅ Tag %s created and pushed!\n' "$NEW_VERSION"
printf 'GitHub Actions will now build and create the release automatically.\n'
