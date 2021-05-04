#!/usr/bin/env bash

echo "[format] formatting code"

for pkg in $(GOFLAGS=-mod=vendor go list -f '{{.Dir}}' ./... | grep -v /vendor/ ); do \
    echo "format files in $pkg"
    go run -mod=vendor golang.org/x/tools/cmd/goimports -l -w -e $pkg/*.go; \
    go run mvdan.cc/gofumpt -l -w  $pkg/*.go;
done