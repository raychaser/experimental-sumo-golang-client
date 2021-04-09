#!/usr/bin/env bash
set -e

OUTPUT=$(git status --untracked-files=no --porcelain)
if [ -z "$OUTPUT" ]; then
  echo "Working directory is clean"
  echo
else
  echo "Working directory dirty. Exiting..."
  echo
  echo "$OUTPUT"
  echo
  git status
  exit 1
fi

OUTPUT=$(git status -sb)
if [[ "$OUTPUT" =~ .*"ahead".* ]]; then
  echo "Local branch ahead of remote. Exiting..."
  echo
  echo "$OUTPUT"
  exit 1
fi

echo Previous tags
git describe --tags
echo

TAG=$1
if [ -z "$TAG" ]; then
  echo "No tag specified. Exiting..."
  exit 1
fi
if [[ "$TAG" =~ [0-9]+\.[0-9]+\.[0-9]+ ]]; then
  echo New tag: $TAG
else
  echo "Tag doesn't look like a semver. Exiting..."
  exit 1
fi

git tag -a "v$TAG" -m "Release $TAG"
echo "Pushing new tag to origin..."
git push origin "v$TAG"

echo
echo Invoking goreleaser
echo
goreleaser --debug --rm-dist
