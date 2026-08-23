// Package parser provides file parsing
package parser

import (
	ts "github.com/tree-sitter/go-tree-sitter"
)

// Parser is the parser.
type Parser struct {
	adapter Adapter
}

// New parser using provided adapter.
func New(adapter Adapter) Parser {
	return Parser{adapter: adapter}
}

// Parse the file content.
func (p Parser) Parse(uri string, source []byte) (*ParseTree, bool) {
	tsp := ts.NewParser()
	if err := tsp.SetLanguage(p.adapter.Language()); err != nil {
		return nil, false
	}

	tree := tsp.Parse(source, nil)
	tsp.Close()

	if tree == nil {
		return nil, false
	}

	return &ParseTree{
		uri:     uri,
		source:  source,
		tree:    tree,
		adapter: p.adapter,
	}, true
}
