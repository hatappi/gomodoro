//go:build tools
// +build tools

package tools

//go:generate go install github.com/golangci/golangci-lint/cmd/golangci-lint

// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
import (
	_ "github.com/99designs/gqlgen"
	_ "github.com/99designs/gqlgen/graphql/introspection"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
)
