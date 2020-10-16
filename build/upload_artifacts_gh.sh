#!/bin/bash
#
#
# Upload dist artifacts to GH Release assets
#
#
cd dist
for package in $(find  -regex ".*\.\(msi\|rpm\|deb\|zip\|tar.gz\)");do
  echo "===> Uploading to GH $TAG: ${package}"
  hub release edit -a ${package} -m "${TAG}" ${TAG}
done
