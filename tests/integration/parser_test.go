//go:build !race

package integration_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vulns-are-features-too/func-tracer/lang/parser"
	parser_adapters "github.com/vulns-are-features-too/func-tracer/lang/parser/adapters"
	"github.com/vulns-are-features-too/func-tracer/model"
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

	symbolsFromName := tree.FindFunctionByName(fn.Name)
	assertSingle(t, symbolsFromName, func(sym model.Symbol) bool {
		return sym == symbolFromLocation
	})
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

func TestFindFunctionCallsInFiles(t *testing.T) {
	t.Parallel()
	testFindFunctionCallsInFiles(t, test_files.GoFiles, parser_adapters.Go())
	testFindFunctionCallsInFiles(t, test_files.RustFiles, parser_adapters.Rust())
}

func testFindFunctionCallsInFiles(
	t *testing.T,
	files test_files.TestFileList,
	adapter parser.Adapter,
) {
	t.Helper()

	dir := files.TestDir()
	t.Run(dir, func(t *testing.T) {
		t.Parallel()

		for _, file := range files.TestFiles() {
			testFindFunctionCallsInFile(t, adapter, dir, file)
		}
	})
}

func testFindFunctionCallsInFile(
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
				testFindFunctionCalls(t, tree, file.Name(), fn)
			})
		}
	})
}

func testFindFunctionCalls(
	t *testing.T,
	tree *parser.ParseTree,
	filename string,
	fn *test_files.FuncInfo,
) {
	t.Helper()

	pos := pos(fn.Line, fn.Char)
	loc := model.Location{
		URI:   filename,
		Range: model.Range{Start: pos, End: pos},
	}

	actualCalls := tree.FindFunctionCalls(loc)

	if !assert.Len(t, actualCalls, len(fn.Calls)) {
		return
	}

	calls := make([]test_files.FuncCall, 0, len(actualCalls))
	for _, call := range actualCalls {
		calls = append(calls, symbolToCall(call))
	}

	test_files.SortFuncCalls(calls)

	assert.Equal(t, fn.Calls, calls)
}

func symbolToCall(s model.Symbol) test_files.FuncCall {
	return test_files.FuncCall{
		Name: s.Name,
		Line: s.Location.Range.Start.Line,
		Char: s.Location.Range.Start.Character,
	}
}
