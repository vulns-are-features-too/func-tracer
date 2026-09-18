// Package parser_adapters provides tree-sitter adapters
package parser_adapters

import (
	"slices"

	ts "github.com/tree-sitter/go-tree-sitter"
	ts_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	"github.com/vulns-are-features-too/func-tracer/lang/parser"
)

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

func (a *goAdapter) IsFunctionDecl(node *ts.Node) bool {
	kinds := []string{
		"function_declaration",
		"generator_function_declaration",
		"method_declaration",
	}

	return slices.Contains(kinds, node.Kind())
}

func (a *goAdapter) GetFuncCall(callExpr *ts.Node) *ts.Node {
	if callExpr == nil {
		return nil
	}

	switch callExpr.Kind() {
	case "identifier", "field_identifier": // foo()
		return callExpr

	case "selector_expression": // foo.bar()
		return a.GetFuncCall(callExpr.ChildByFieldName("field"))

	default:
		return nil
	}
}
