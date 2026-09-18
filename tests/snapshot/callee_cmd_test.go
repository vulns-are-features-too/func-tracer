package snapshot_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/vulns-are-features-too/func-tracer/tests/test_files"
)

func TestCalleeCmdOnGoFiles(t *testing.T) {
	files := test_files.GoFiles
	dir := files.TestDir()

	testCalleeByPosition(t, dir, files.Main.Name(), files.Main.Main, "-d", "2")
	testCalleeByName(t, dir, files.Main.Name(), files.Main.Main)
}

func TestCalleeCmdOnRustFiles(t *testing.T) {
	files := test_files.RustFiles
	dir := files.TestDir()

	testCalleeByPosition(t, dir, files.Main.Name(), files.Main.Main, "-d", "2")
	testCalleeByName(t, dir, files.Main.Name(), files.Main.Main)
}

func testCalleeByPosition(
	t *testing.T,
	dir string,
	file string,
	fn test_files.FuncInfo,
	extraArgs ...string,
) {
	t.Helper()

	//nolint:prealloc
	args := []string{
		"callee",
		"-r", testDataDir(dir),
		"-f", testFile(dir, file),
		"-l", strconv.Itoa(int(fn.Line)),
		"-c", strconv.Itoa(int(fn.Char)),
	}
	args = append(args, extraArgs...)

	desc := strings.Join(args, " ")

	snapshotSuffix := fmt.Sprintf("callee_l%d_c%d", fn.Line, fn.Char)
	if len(extraArgs) > 0 {
		snapshotSuffix = fmt.Sprintf("%s_%s", snapshotSuffix, argsToFilename(extraArgs...))
	}

	t.Run(desc, func(t *testing.T) {
		runTest(t, dir, file, fn, snapshotSuffix, args)
	})
}

func testCalleeByName(
	t *testing.T,
	dir string,
	file string,
	fn test_files.FuncInfo,
	extraArgs ...string,
) {
	t.Helper()

	//nolint:prealloc
	args := []string{
		"callee",
		"-r", testDataDir(dir),
		"-f", testFile(dir, file),
		"-n", fn.Name,
	}
	args = append(args, extraArgs...)

	desc := strings.Join(args, " ")

	snapshotSuffix := "callee_n_" + fn.Name
	if len(extraArgs) > 0 {
		snapshotSuffix = fmt.Sprintf("%s_%s", snapshotSuffix, argsToFilename(extraArgs...))
	}

	t.Run(desc, func(t *testing.T) {
		runTest(t, dir, file, fn, snapshotSuffix, args)
	})
}
