#!/bin/bash

pwd=`pwd`
echo "running golangci validation..."
go run  -mod=mod github.com/golangci/golangci-lint/cmd/golangci-lint run --verbose
echo "running semgrep validation..."
docker run --rm -v "${pwd}:/src:ro" --workdir /src returntocorp/semgrep -c "/src/.semgrep.yml"
