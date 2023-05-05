//go:build tools
// +build tools

package tools

//go:generate go install github.com/golangci/golangci-lint/cmd/golangci-lint
//go:generate go install github.com/reviewdog/reviewdog/cmd/reviewdog

// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/reviewdog/reviewdog/cmd/reviewdog"
)
