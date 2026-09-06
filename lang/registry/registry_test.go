//go:build !race

package registry_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vulns-are-features-too/func-tracer/lang"
	"github.com/vulns-are-features-too/func-tracer/lang/registry"
)

func TestAllLanguagesHaveLsps(t *testing.T) {
	t.Parallel()

	langs := registry.ListLanguages()

	assert.Equal(t, langs, mapKeys(registry.ListLSP()))

	for _, l := range langs {
		lsp, err := registry.LSP(l)
		assert.NotNil(t, lsp)
		require.NoError(t, err)
	}
}

func TestAllLanguagesHaveParsers(t *testing.T) {
	t.Parallel()

	for _, l := range registry.ListLanguages() {
		parser, err := registry.Parser(l)
		assert.NotNil(t, parser)
		require.NoError(t, err)
	}
}

func TestAllLanguagesHaveExtensions(t *testing.T) {
	t.Parallel()

	langs := registry.ListLanguages()
	for _, l := range langs {
		ext, ok := registry.Extensions(l)
		assert.NotEmpty(t, ext)
		assert.True(t, ok)
	}
}

func TestListLanguagesIsSorted(t *testing.T) {
	t.Parallel()

	assert.True(t, slices.IsSorted(registry.ListLanguages()))
}

func mapKeys[V any](m map[lang.Language]V) []lang.Language {
	res := slices.Collect(maps.Keys(m))
	slices.Sort(res)

	return res
}
