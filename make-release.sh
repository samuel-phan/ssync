#!/usr/bin/env bash

git fetch --all

echo "Make sure you are on the right branch:"
git branch -vv

echo "Existing tags:"
git --no-pager tag

echo -n "GitHub token for release (hidden output):"
read -s github_token
echo
export GITHUB_TOKEN="${github_token}"

echo -n "Enter the version you want to release (eg: 0.0.1): "
read version

git tag -a "v${version}" -m "Release v${version}"
git push origin "v${version}"

# Check the GoRelease locally
goreleaser --skip-publish --rm-dist || exit $?

echo "Check the local release if fine. Type \"yes\" to do the real release..."
read answer
if [ "${answer}" != "yes" ]; then
    echo "Abort."
    exit 1
fi

# If all good, do the real release
goreleaser --rm-dist
