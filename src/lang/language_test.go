//go:build !race

package lang_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vulns-are-features-too/func-tracer/src/lang"
)

func TestNoDuplicate(t *testing.T) {
	t.Parallel()

	l := lang.SupportedLanguages
	slices.Sort(l)
	l = slices.Compact(l)
	assert.Len(t, lang.SupportedLanguages, len(l))
}
