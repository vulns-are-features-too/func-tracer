//go:build !race

package integration_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vulns-are-features-too/func-tracer/src/lang/parser"
	parser_adapters "github.com/vulns-are-features-too/func-tracer/src/lang/parser/adapters"
	"github.com/vulns-are-features-too/func-tracer/src/model"
	"github.com/vulns-are-features-too/func-tracer/tests/test_files"
)

func TestFindFunctionInFiles(t *testing.T) {
	t.Parallel()
	testFindFunctionInFiles(t, test_files.GoFiles, parser_adapters.Go())
	testFindFunctionInFiles(t, test_files.RustFiles, parser_adapters.Rust())
}

func testFindFunctionInFiles(
	t *testing.T,
	files test_files.TestFileList,
	adapter parser.Adapter,
) {
	t.Helper()

	dir := files.TestDir()
	t.Run(dir, func(t *testing.T) {
		t.Parallel()

		for _, file := range files.TestFiles() {
			testFindFunctionInFile(t, adapter, dir, file)
		}
	})
}

func testFindFunctionInFile(
	t *testing.T,
	adapter parser.Adapter,
	dir string,
	file test_files.TestFile,
) {
	t.Helper()
	t.Run(file.Name(), func(t *testing.T) {
		t.Parallel()

		tree := parseFile(t, adapter, testFile(dir, file.Name()))
		defer tree.Close()

		for _, fn := range file.ListFunctions() {
			t.Run(fn.Name, func(t *testing.T) {
				testFindFunction(t, tree, fn)
			})
		}
	})
}

func testFindFunction(
	t *testing.T,
	tree *parser.ParseTree,
	fn *test_files.FuncInfo,
) {
	t.Helper()

	pos := pos(fn.Line, fn.Char)
	loc := model.Location{Range: model.Range{Start: pos, End: pos}}

	symbolFromLocation, ok := tree.FindFunction(loc)
	assert.True(t, ok)
	assert.Equal(t, fn.Name, symbolFromLocation.Name)
	assert.Equal(t, pos, symbolFromLocation.Location.Range.Start)

	symbolFromName, ok := tree.FindFunctionByName(fn.Name)
	assert.True(t, ok)
	assert.Equal(t, fn.Name, symbolFromName.Name)
	assert.Equal(t, pos, symbolFromName.Location.Range.Start)

	assert.Equal(t, symbolFromLocation, symbolFromName)
}

func parseFile(t *testing.T, adapter parser.Adapter, path string) *parser.ParseTree {
	t.Helper()

	//nolint:gosec // G304 just test files
	source, err := os.ReadFile(path)
	require.NoError(t, err)

	tree, ok := parser.New(adapter).Parse(path, source)
	assert.True(t, ok)
	assert.NotNil(t, tree)

	return tree
}
