package parser_adapters

import (
	ts "github.com/tree-sitter/go-tree-sitter"
	ts_rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	"github.com/vulns-are-features-too/func-tracer/lang/parser"
)

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

func (a *rustAdapter) IsFunctionDecl(node *ts.Node) bool {
	return node.Kind() == "function_item"
}

func (a *rustAdapter) GetFuncCall(callExpr *ts.Node) *ts.Node {
	if callExpr == nil {
		return nil
	}

	switch callExpr.Kind() {
	case "identifier", "field_identifier": // foo()
		return callExpr

	case "field_expression": // foo.bar()
		return a.GetFuncCall(callExpr.ChildByFieldName("field"))

	case "scoped_identifier": // foo::bar()
		return a.GetFuncCall(callExpr.ChildByFieldName("name"))

	case "generic_function": // foo::<T>() / foo.bar::<T>()
		return a.GetFuncCall(callExpr.ChildByFieldName("function"))

	default:
		return nil
	}
}
