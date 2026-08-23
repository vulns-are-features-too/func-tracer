package parser_adapters

import (
	"slices"

	ts "github.com/tree-sitter/go-tree-sitter"
	ts_rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	"github.com/vulns-are-features-too/func-tracer/src/lang/parser"
)

//nolint:gochecknoglobals
var rustFuncKinds = []string{
	"function_item",
}

type rustAdapter struct {
	language *ts.Language
}

// Rust parser adapter.
//
//nolint:ireturn
func Rust() parser.Adapter {
	return &rustAdapter{
		language: ts.NewLanguage(ts_rust.Language()),
	}
}

func (a *rustAdapter) Language() *ts.Language {
	return a.language
}

func (a *rustAdapter) IsFunctionKind(kind string) bool {
	return slices.Contains(rustFuncKinds, kind)
}
