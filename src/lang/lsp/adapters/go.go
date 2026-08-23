// Package lsp_adapters provides concrete LSP integrations
package lsp_adapters

import (
	"github.com/vulns-are-features-too/func-tracer/src/lang"
)

// GoLsp provides gopls.
type GoLsp struct{}

// Go LSP adapter.
func Go() *GoLsp {
	return &GoLsp{}
}

// Command gopls.
func (*GoLsp) Command() string {
	return "gopls"
}

// Args for gopls.
func (*GoLsp) Args() []string {
	return []string{"serve"}
}

// Language go.
func (*GoLsp) Language() lang.Language {
	return lang.Go
}
