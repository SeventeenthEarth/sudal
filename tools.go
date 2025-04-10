//go:build tools
// +build tools

package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/onsi/ginkgo/v2/ginkgo"
	_ "go.uber.org/mock/mockgen"
)

// This file imports tool dependencies
// It ensures they're tracked in go.mod but not included in the final build
