// Package integration_test is for integration tests
package integration_test

import (
	"net/url"
	"os/exec"
	"path"
	"path/filepath"
	"testing"

	"github.com/vulns-are-features-too/func-tracer/logging"
	"github.com/vulns-are-features-too/func-tracer/model"
	"github.com/vulns-are-features-too/func-tracer/tests/test_files"
)

const testFilesDir = "../test_files"

func testFile(dir string, file string) string {
	return path.Join(testFilesDir, dir, file)
}

func pos(line uint, char uint) model.Position {
	return model.Position{Line: line, Character: char}
}

func fnLoc(
	dir string,
	file test_files.TestFile,
	fn *test_files.FuncInfo,
) model.Location {
	p := pos(fn.Line, fn.Char)

	uri, err := fileURI(testFile(dir, file.Name()))
	if err != nil {
		panic(err)
	}

	return model.Location{
		URI:   uri,
		Range: model.Range{Start: p, End: p},
	}
}

func requireBin(t *testing.T, bin string) {
	t.Helper()

	if _, err := exec.LookPath(bin); err != nil {
		t.Fatalf("%s not installed", bin)
	}
}

func logger(t *testing.T) *logging.TestLogger {
	t.Helper()

	return logging.Test(t)
}

func fileURI(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	uri := (&url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(absolute),
	}).String()

	return uri, nil
}
