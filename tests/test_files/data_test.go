package test_files_test

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vulns-are-features-too/func-tracer/tests/test_files"
)

// filename => lines.
type fileCache map[string][]string

func TestValidateData(t *testing.T) {
	t.Parallel()

	validateDir(t, test_files.GoFiles, func(line string) bool {
		return strings.HasPrefix(line, "func")
	})

	validateDir(t, test_files.RustFiles, func(line string) bool {
		return strings.HasPrefix(line, "fn ") || strings.Contains(line, " fn ")
	})
}

func validateDir(t *testing.T, files test_files.TestFileList, isLineFunc func(string) bool) {
	t.Helper()

	cache := cacheFiles(t, files)
	dir := files.TestDir()
	t.Run(dir, func(t *testing.T) {
		t.Parallel()

		for _, file := range files.TestFiles() {
			t.Run(file.Name(), func(t *testing.T) {
				t.Parallel()
				validateFile(t, cache, file, isLineFunc)
			})
		}
	})
}

func validateFile(
	t *testing.T,
	cache fileCache,
	file test_files.TestFile,
	isLineFunc func(string) bool,
) {
	t.Helper()

	funcCount := countFunctions(file)
	fns := file.ListFunctions()
	assert.Len(t, fns, funcCount, "Not all functions returned in ListFunctions")

	lines := cache[file.Name()]

	for _, fn := range fns {
		s := getTextAtFuncDecl(t, lines, fn)
		assert.Equal(t, fn.Name, s, "Wrong function postition in declared data")

		for _, ref := range fn.Refs {
			s := getTextAtRef(t, cache, lines, fn, &ref)
			assert.Equalf(
				t,
				fn.Name,
				s,
				"Wrong reference postition in declared data @ %s[%d:%d]",
				ref.File,
				ref.Line,
				ref.Char,
			)
		}
	}

	funcLines := 0

	for _, line := range lines {
		if isLineFunc(line) {
			funcLines++
		}
	}

	assert.Equal(t, funcCount, funcLines, "Function count mismatch between declared data and file")
}

func getTextAtFuncDecl(
	t *testing.T,
	lines []string,
	fn *test_files.FuncInfo,
) string {
	t.Helper()

	line := lines[fn.Line]
	start := fn.Char
	end := start + uint(len(fn.Name))

	if len(line) < int(end) {
		t.Logf("Wrong line: %d", fn.Line)

		if fn.Line > 0 {
			t.Log(fmt.Printf("Above line %d = %s\n", fn.Line, lines[fn.Line-1]))
		}

		t.Log(fmt.Printf("Line %d:[%d:%d] = %s\n", fn.Line, start, end, line))

		if int(fn.Line) < len(lines) {
			t.Log(fmt.Printf("Below line %d = %s\n", fn.Line, lines[fn.Line+1]))
		}

		t.FailNow()
	}

	return line[start:end]
}

func getTextAtRef(
	t *testing.T,
	cache fileCache,
	lines []string,
	fn *test_files.FuncInfo,
	ref *test_files.FuncRef,
) string {
	t.Helper()

	f := cache[ref.File]
	line := f[ref.Line]
	start := ref.Char
	end := start + uint(len(fn.Name))

	if len(line) < int(end) {
		t.Logf("Wrong line: %d", ref.Line)

		if ref.Line > 0 {
			t.Log(fmt.Printf("Above ref %d = %s\n", fn.Line, f[ref.Line-1]))
		}

		t.Log(fmt.Printf("Ref %d:[%d:%d] = %s\n", fn.Line, start, end, line))

		if int(ref.Line) < len(lines) {
			t.Log(fmt.Printf("Below ref %d = %s\n", fn.Line, f[ref.Line+1]))
		}

		t.FailNow()
	}

	return line[start:end]
}

func countFunctions(file test_files.TestFile) int {
	fns := 0

	for _, field := range reflect.ValueOf(file).Fields() {
		if field.Type() == reflect.TypeFor[test_files.FuncInfo]() {
			fns++
		}
	}

	return fns
}

func cacheFiles(t *testing.T, files test_files.TestFileList) fileCache {
	t.Helper()

	dir := files.TestDir()
	cache := make(fileCache)

	for _, file := range files.TestFiles() {
		content, err := os.ReadFile(fmt.Sprintf("%s/%s", dir, file.Name()))
		require.NoError(t, err)

		lines := strings.Split(string(content), "\n")
		cache[file.Name()] = lines
	}

	return cache
}
