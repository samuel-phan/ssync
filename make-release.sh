#!/usr/bin/env bash
# Copyright 2021 Yahoo Inc.
# Licensed under the terms of the Apache 2.0 License. See LICENSE file in project root for terms.

git fetch --all

echo "Make sure you are on the right branch:"
git branch -vv
echo -n "If you're on the right branch, press [Enter]..."
read

echo "Existing tags:"
git --no-pager tag

echo -n "GitHub token for release (hidden output):"
read -s github_token
echo
export GITHUB_TOKEN="${github_token}"

echo -n "Enter the version you want to release (eg: 0.0.1): "
read version

git push
git tag -a "v${version}" -m "Release v${version}"
git push origin "v${version}"

# Check the GoRelease locally
goreleaser --skip-publish --clean || exit $?

echo "Check the local release if fine. Type \"yes\" to do the real release..."
read answer
if [ "${answer}" != "yes" ]; then
    echo "Abort."
    exit 1
fi

# If all good, do the real release
goreleaser --clean
