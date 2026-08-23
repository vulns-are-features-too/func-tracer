package caller

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/vulns-are-features-too/func-tracer/src/lang"
	"github.com/vulns-are-features-too/func-tracer/src/lang/registry"
)

var (
	errUnsupportedLanguage    = errors.New("unsupported language")
	errFailedToGetSourceFiles = errors.New("failed to get source files")
)

var blacklistedDirs = []string{
	".git",
}

func isBlacklistedDir(path string) bool {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}

	dir := filepath.Base(filepath.Clean(path))

	return strings.HasPrefix(dir, ".") || slices.Contains(blacklistedDirs, dir)
}

func sourceFiles(root string, l lang.Language) ([]string, error) {
	extensions, ok := registry.Extensions(l)
	if !ok {
		return nil, fmt.Errorf("%w: %s", errUnsupportedLanguage, l)
	}

	var files []string

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if isBlacklistedDir(entry.Name()) {
				return filepath.SkipDir
			}

			return nil
		}

		if slices.Contains(extensions, strings.ToLower(filepath.Ext(path))) {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToGetSourceFiles, err)
	}

	return files, nil
}
