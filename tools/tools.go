// +build tools

package tools

import (
	_ "github.com/axw/gocov/gocov"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/AlekSi/gocov-xml"
	_ "golang.org/x/tools/cmd/goimports"
	_ "mvdan.cc/gofumpt"
	_ "github.com/josephspurrier/goversioninfo"
)