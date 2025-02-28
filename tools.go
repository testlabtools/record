//go:build tools
// +build tools

package record

import (
	_ "github.com/jstemmer/go-junit-report/v2"
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
	_ "go.uber.org/nilaway/cmd/nilaway"
)
