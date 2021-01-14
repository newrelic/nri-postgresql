#!/bin/bash
set -e
#
#
# Upload dist artifacts to GH Release assets
#
#
cd dist
find . -regex ".*\.\(msi\|rpm\|deb\|zip\|tar.gz\)" | while read filename; do
  echo "===> Uploading to GH $TAG: ${filename}"
      gh release upload $TAG $filename
done
