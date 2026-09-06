// Package parser_adapters provides tree-sitter adapters
package parser_adapters

import (
	"slices"

	ts "github.com/tree-sitter/go-tree-sitter"
	ts_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	"github.com/vulns-are-features-too/func-tracer/lang/parser"
)

//nolint:gochecknoglobals
var goFuncKinds = []string{
	"function_declaration",
	"generator_function_declaration",
	"method_declaration",
}

type goAdapter struct {
	language *ts.Language
}

// Go parser adapter.
//
//nolint:ireturn
func Go() parser.Adapter {
	return &goAdapter{
		language: ts.NewLanguage(ts_go.Language()),
	}
}

func (a *goAdapter) Language() *ts.Language {
	return a.language
}

func (a *goAdapter) IsFunctionKind(kind string) bool {
	return slices.Contains(goFuncKinds, kind)
}
