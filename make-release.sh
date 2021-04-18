#!/usr/bin/env bash

git fetch --all

echo "Make sure you are on the right branch:"
git branch -vv

echo "Existing tags:"
git --no-pager tag

echo -n "GitHub token for release (hidden output):"
read -s github_token
export GITHUB_TOKEN="${github_token}"

echo "Which version do you want to release? Eg: 0.0.1"
read version

git tag -a "v${version}" -m "Release v${version}"
git push origin "v${version}"

# Check the GoRelease locally
goreleaser --skip-publish --rm-dist

echo "Check the local release if fine. Press [Enter] to do the real release..."
read

# If all good, do the real release
goreleaser --rm-dist
