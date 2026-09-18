package parser

import (
	ts "github.com/tree-sitter/go-tree-sitter"
)

// Adapter for tree-sitter for parsing.
type Adapter interface {
	Language() *ts.Language
	IsFunctionDecl(node *ts.Node) bool
	GetFuncCall(callExpr *ts.Node) *ts.Node
}
