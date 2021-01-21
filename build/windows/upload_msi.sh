#!/bin/bash
set -e
#
#
# Upload msi artifacts to GH Release assets
#
#
INTEGRATION=$1
ARCH=$2
TAG=$3

gh release upload "$TAG" "build/package/windows/nri-${ARCH}-installer/bin/Release/nri-${INTEGRATION}-${ARCH}.${TAG:1}.msi" --repo "$REPO_FULL_NAME"
