package parser

import (
	ts "github.com/tree-sitter/go-tree-sitter"
)

// Adapter for tree-sitter for parsing.
type Adapter interface {
	Language() *ts.Language
	IsFunctionKind(kind string) bool
}
