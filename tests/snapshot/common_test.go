// Package snapshot_test is for snapshot testing
package snapshot_test

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vulns-are-features-too/func-tracer/cli"
	"github.com/vulns-are-features-too/func-tracer/tests/test_files"
)

const (
	snapshotsEnv      = "UPDATE_SNAPSHOT"
	testFilesDir      = "../test_files"
	snapshotsDir      = "./_snapshots"
	snapshotDirPerms  = 0o700
	snapshotFilePerms = 0o600
)

func shouldUpdateSnapshot() bool {
	_, ok := os.LookupEnv(snapshotsEnv)

	return ok
}

func testDataDir(dir string) string {
	return path.Join(testFilesDir, dir)
}

func testFile(dir string, file string) string {
	return path.Join(testFilesDir, dir, file)
}

func snapshotDir(dataDir string, testFile string) string {
	return path.Join(snapshotsDir, dataDir, testFile)
}

func snapshotFile(dataDir string, testFile string, fn string, desc string) string {
	return path.Join(snapshotDir(dataDir, testFile), fmt.Sprintf("%s.%s.txt", fn, desc))
}

func compareSnapshot(t *testing.T, file string, result string) {
	t.Helper()

	//nolint:gosec // G304
	snap, err := os.ReadFile(file)
	require.NoError(t, err)
	assert.Equal(t, string(snap), result)
}

func updateSnapshot(dir string, file string, result string) {
	err := os.MkdirAll(dir, snapshotDirPerms)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(file, []byte(result), snapshotFilePerms)
	if err != nil {
		panic(err)
	}
}

func runTest(
	t *testing.T,
	dir string,
	file string,
	fn test_files.FuncInfo,
	snapshotSuffix string,
	args []string,
) {
	t.Helper()

	result, err := runCmd(args...)
	require.NoError(t, err)

	sdir := snapshotDir(dir, file)
	sfile := snapshotFile(dir, file, fn.Name, snapshotSuffix)

	if shouldUpdateSnapshot() {
		updateSnapshot(sdir, sfile, result)
	} else {
		compareSnapshot(t, sfile, result)
	}
}

func runCmd(args ...string) (string, error) {
	var out bytes.Buffer

	cmd := cli.RootCmd()
	cmd.SetArgs(args)
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()

	return out.String(), err
}

func argsToFilename(args ...string) string {
	if len(args) == 0 {
		return ""
	}

	sb := strings.Builder{}
	sb.WriteString(args[0])

	for _, a := range args[1:] {
		sb.WriteRune('_')
		sb.WriteString(strings.TrimPrefix(a, "-"))
	}

	return sb.String()
}
