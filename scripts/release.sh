#!/usr/bin/env bash
# Bump the semver tag and push it to trigger the GitHub Actions release.
# Usage: ./scripts/release.sh [patch|minor|major]
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

VERSION_TYPE="${1:-patch}"
case "$VERSION_TYPE" in
  major|minor|patch) ;;
  *)
    printf 'Error: version type must be patch, minor, or major\n' >&2
    exit 1
    ;;
esac

CURRENT_VERSION="$(git describe --tags --abbrev=0 --match 'v[0-9]*' 2>/dev/null || echo '0.0.0')"
CURRENT_VERSION="${CURRENT_VERSION#v}"

if [[ ! "$CURRENT_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  printf 'Error: current tag is not a valid semver version\n' >&2
  exit 1
fi

IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"
case "$VERSION_TYPE" in
  major) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
  minor) MINOR=$((MINOR + 1)); PATCH=0 ;;
  patch) PATCH=$((PATCH + 1)) ;;
esac

NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"

if [ -n "$(git status --porcelain)" ]; then
  printf 'Error: working tree is not clean; commit or stash changes first.\n' >&2
  exit 1
fi

if ! git rev-parse --abbrev-ref --symbolic-full-name '@{u}' >/dev/null 2>&1; then
  printf 'Error: current branch has no upstream.\n' >&2
  exit 1
fi

if [ "$(git rev-list --count '@{u}..HEAD')" -ne 0 ]; then
  printf 'Error: local commits are not pushed; push the branch first.\n' >&2
  exit 1
fi

if git rev-parse -q --verify "refs/tags/${NEW_VERSION}" >/dev/null ||
  git ls-remote --exit-code --tags origin "refs/tags/${NEW_VERSION}" >/dev/null 2>&1; then
  printf 'Error: tag %s already exists.\n' "$NEW_VERSION" >&2
  exit 1
fi

TMPFILE=$(mktemp)
trap 'rm -f "$TMPFILE"' EXIT

./scripts/changelog.sh "$NEW_VERSION" > "$TMPFILE"

printf 'Current version : %s\n' "$CURRENT_VERSION"
printf 'New version     : %s\n' "$NEW_VERSION"
printf '\nChangelog:\n\n'
cat "$TMPFILE"
printf '\n'

read -r -p "Create and push tag ${NEW_VERSION}? (y/N) " REPLY
printf '\n'

case "$REPLY" in
  [yY]) ;;
  *) printf 'Cancelled\n'; exit 1 ;;
esac

git tag -a "$NEW_VERSION" -F "$TMPFILE"
git push origin "$NEW_VERSION"

printf 'Tag %s created and pushed.\n' "$NEW_VERSION"
printf 'GitHub Actions will now build and create the release automatically.\n'
