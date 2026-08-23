//go:build !race

package lsp_adapters_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vulns-are-features-too/func-tracer/src/lang/lsp"
	lsp_adapters "github.com/vulns-are-features-too/func-tracer/src/lang/lsp/adapters"
)

func TestImplAdapter(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*lsp.Adapter)(nil), new(lsp_adapters.GoLsp))
	assert.Implements(t, (*lsp.Adapter)(nil), new(lsp_adapters.RustLsp))
}
