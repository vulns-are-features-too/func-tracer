// Package lsp_adapters provides concrete LSP integrations
package lsp_adapters

import (
	"github.com/vulns-are-features-too/func-tracer/lang"
)

// GoPlsLsp provides gopls.
type GoPlsLsp struct{}

// GoPls LSP adapter.
func GoPls() *GoPlsLsp {
	return &GoPlsLsp{}
}

// Command gopls.
func (*GoPlsLsp) Command() string {
	return "gopls"
}

// Args for gopls.
func (*GoPlsLsp) Args() []string {
	return []string{"serve"}
}

// Language go.
func (*GoPlsLsp) Language() lang.Language {
	return lang.Go
}
