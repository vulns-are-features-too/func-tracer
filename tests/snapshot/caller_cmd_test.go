package snapshot_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/vulns-are-features-too/func-tracer/tests/test_files"
)

func TestCallerCmdOnGoFiles(t *testing.T) {
	files := test_files.GoFiles
	dir := files.TestDir()

	testCallerByPosition(t, dir, files.Db.Name(), files.Db.Query, "-d", "1")

	testCallerByName(t, dir, files.Db.Name(), files.Db.Query)
	testCallerByName(t, dir, files.Recurse.Name(), files.Recurse.RecurseSelf)
	testCallerByName(t, dir, files.Recurse.Name(), files.Recurse.RecurseOther)
	testCallerByName(t, dir, files.Nested.Name(), files.Nested.Nested5)
	testCallerByName(t, dir, files.Nested.Name(), files.Nested.Nested5, "-d", "2")
}

func TestCallerCmdOnRustFiles(t *testing.T) {
	files := test_files.RustFiles
	dir := files.TestDir()

	testCallerByPosition(t, dir, files.Db.Name(), files.Db.Query, "-d", "1")

	testCallerByName(t, dir, files.Db.Name(), files.Db.Query)
	testCallerByName(t, dir, files.Recurse.Name(), files.Recurse.RecurseSelf)
	testCallerByName(t, dir, files.Recurse.Name(), files.Recurse.RecurseOther)
	testCallerByName(t, dir, files.Nested.Name(), files.Nested.Nested5)
	testCallerByName(t, dir, files.Nested.Name(), files.Nested.Nested5, "-d", "2")
}

func testCallerByPosition(
	t *testing.T,
	dir string,
	file string,
	fn test_files.FuncInfo,
	extraArgs ...string,
) {
	t.Helper()

	//nolint:prealloc
	args := []string{
		"caller",
		"-r", testDataDir(dir),
		"-f", testFile(dir, file),
		"-l", strconv.Itoa(int(fn.Line)),
		"-c", strconv.Itoa(int(fn.Char)),
	}
	args = append(args, extraArgs...)

	desc := strings.Join(args, " ")

	snapshotSuffix := fmt.Sprintf("caller_l%d_c%d", fn.Line, fn.Char)
	if len(extraArgs) > 0 {
		snapshotSuffix = fmt.Sprintf("%s_%s", snapshotSuffix, argsToFilename(extraArgs...))
	}

	t.Run(desc, func(t *testing.T) {
		runTest(t, dir, file, fn, snapshotSuffix, args)
	})
}

func testCallerByName(
	t *testing.T,
	dir string,
	file string,
	fn test_files.FuncInfo,
	extraArgs ...string,
) {
	t.Helper()

	//nolint:prealloc
	args := []string{
		"caller",
		"-r", testDataDir(dir),
		"-f", testFile(dir, file),
		"-n", fn.Name,
	}
	args = append(args, extraArgs...)

	desc := strings.Join(args, " ")

	snapshotSuffix := "caller_n_" + fn.Name
	if len(extraArgs) > 0 {
		snapshotSuffix = fmt.Sprintf("%s_%s", snapshotSuffix, argsToFilename(extraArgs...))
	}

	t.Run(desc, func(t *testing.T) {
		runTest(t, dir, file, fn, snapshotSuffix, args)
	})
}
