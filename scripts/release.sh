#!/usr/bin/env bash
set -euo pipefail

# Helper script to create a new release tag with changelog generation
# Usage: ./scripts/release.sh [patch|minor|major] [message]

VERSION_TYPE="${1:-patch}"
RELEASE_MESSAGE="${2:-}"

# Get current version from latest tag
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "0.0.0")
IFS='.' read -r -a VERSION_PARTS <<< "$CURRENT_VERSION"
MAJOR="${VERSION_PARTS[0]}"
MINOR="${VERSION_PARTS[1]}"
PATCH="${VERSION_PARTS[2]}"

# Bump version
case "$VERSION_TYPE" in
  major)
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
    ;;
  minor)
    MINOR=$((MINOR + 1))
    PATCH=0
    ;;
  patch)
    PATCH=$((PATCH + 1))
    ;;
  *)
    echo "Error: Version type must be 'major', 'minor', or 'patch'"
    exit 1
    ;;
esac

NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"

# Generate changelog from commits since last tag
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
if [ -z "$LAST_TAG" ]; then
  COMMITS=$(git log --oneline --no-decorate)
else
  COMMITS=$(git log ${LAST_TAG}..HEAD --oneline --no-decorate)
fi

# Create changelog
CHANGELOG="## What's New\n\n"
if [ -n "$RELEASE_MESSAGE" ]; then
  CHANGELOG+="${RELEASE_MESSAGE}\n\n"
fi

if [ -n "$COMMITS" ]; then
  CHANGELOG+="### Changes since ${LAST_TAG:-initial release}\n\n"
  echo "$COMMITS" | while read -r line; do
    CHANGELOG+="- ${line}\n"
  done
fi

# Confirm
echo "Current version: ${CURRENT_VERSION}"
echo "New version: ${NEW_VERSION}"
echo ""
echo "Changelog:"
echo -e "$CHANGELOG"
echo ""
read -p "Create and push tag ${NEW_VERSION}? (y/N) " -n 1 -r
echo

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "Cancelled"
  exit 1
fi

# Create and push tag
git tag -a "${NEW_VERSION}" -m "$(echo -e "$CHANGELOG")"
git push origin "${NEW_VERSION}"

echo "✅ Tag ${NEW_VERSION} created and pushed!"
echo "GitHub Actions will now build and create the release automatically."

