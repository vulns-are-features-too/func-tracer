//go:build !race

package integration_test

import (
	"cmp"
	"path"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vulns-are-features-too/func-tracer/src/lang/lsp"
	lsp_adapters "github.com/vulns-are-features-too/func-tracer/src/lang/lsp/adapters"
	"github.com/vulns-are-features-too/func-tracer/src/model"
	"github.com/vulns-are-features-too/func-tracer/tests/test_files"
)

func TestLsp(t *testing.T) {
	t.Parallel()
	testLsp(t, test_files.GoFiles, lsp_adapters.Go())
	testLsp(t, test_files.RustFiles, lsp_adapters.Rust())
}

func testLsp(t *testing.T, files test_files.TestFileList, adapter lsp.Adapter) {
	t.Helper()
	requireBin(t, adapter.Command())

	ctx := t.Context()
	logger := logger(t)
	dir := path.Join(testFilesDir, files.TestDir())

	t.Run(files.TestDir(), func(t *testing.T) {
		t.Parallel()

		session, err := lsp.Start(ctx, logger, adapter, dir)
		require.NoError(t, err)

		defer session.Close(ctx)

		for _, file := range files.TestFiles() {
			for _, fn := range file.ListFunctions() {
				t.Run(fn.Name, func(t *testing.T) {
					loc := fnLoc(dir, file, fn)
					t.Log(loc)

					t.Run("References", func(t *testing.T) {
						refs, err := session.References(ctx, loc)
						require.NoError(t, err)
						assertRefs(t, fn, refs)
					})
				})
			}
		}
	})
}

func assertRefs(t *testing.T, expected *test_files.FuncInfo, actual []model.Location) {
	t.Helper()

	refs := expected.Refs
	assert.Len(t, actual, len(refs))

	if len(refs) == 0 {
		return
	}

	if len(refs) > 1 {
		slices.SortFunc(refs, func(l, r test_files.FuncRef) int {
			file := cmp.Compare(l.File, r.File)
			if file != 0 {
				return file
			}

			line := cmp.Compare(l.Line, r.Line)
			if line != 0 {
				return line
			}

			return cmp.Compare(l.Char, r.Char)
		})

		slices.SortFunc(actual, func(l, r model.Location) int {
			file := cmp.Compare(l.URI, r.URI)
			if file != 0 {
				return file
			}

			line := cmp.Compare(l.Range.Start.Line, r.Range.Start.Line)
			if line != 0 {
				return line
			}

			return cmp.Compare(l.Range.Start.Character, r.Range.Start.Character)
		})
	}

	for i, fn := range refs {
		actualRef := actual[i]
		assert.Equal(
			t,
			fn.Line,
			actualRef.Range.Start.Line,
			"Wrong line (file=%s, func=%s)",
			fn.File,
			expected.Name,
		)
		assert.Equal(
			t,
			fn.Char,
			actualRef.Range.Start.Character,
			"Wrong char (file=%s, func=%s)",
			fn.File,
			expected.Name,
		)
	}
}
