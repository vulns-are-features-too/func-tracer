// Package index provides an index of parsed files
package index

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/vulns-are-features-too/func-tracer/lang/parser"
	"github.com/vulns-are-features-too/func-tracer/logging"
	"github.com/vulns-are-features-too/func-tracer/model"
)

var (
	errIndexing   = errors.New("indexing failed")
	errAddingFile = errors.New("failed to index file")
	errParsing    = errors.New("file parsing failed")
)

// Index is an index of parsed files.
type Index interface {
	Build(paths []string) error
	Close()
	FindFunction(location model.Location) (model.Symbol, bool)
	FindFunctionByName(uri string, name string) (model.Symbol, bool)
}

type index struct {
	logger  logging.Logger
	adapter parser.Adapter
	mu      sync.RWMutex
	files   map[string]*parser.ParseTree
}

// New index.
//
//nolint:revive // unexported-return
func New(logger logging.Logger, adapter parser.Adapter) *index {
	return &index{
		logger:  logger,
		adapter: adapter,
		files:   make(map[string]*parser.ParseTree),
	}
}

// Build index.
func (i *index) Build(paths []string) error {
	for _, path := range paths {
		if err := i.addFile(path); err != nil {
			return fmt.Errorf("%w: %w", errIndexing, err)
		}
	}

	return nil
}

// Close index and its file handlers.
func (i *index) Close() {
	i.mu.Lock()
	defer i.mu.Unlock()

	for _, file := range i.files {
		file.Close()
	}
}

// FindFunction by location.
func (i *index) FindFunction(
	location model.Location,
) (model.Symbol, bool) {
	file := i.getFile(location.URI)
	if file == nil {
		return model.Symbol{}, false
	}

	return file.FindFunction(location)
}

// FindFunctionByName file & function name.
func (i *index) FindFunctionByName(
	uri string,
	name string,
) (model.Symbol, bool) {
	file := i.getFile(uri)
	if file == nil {
		return model.Symbol{}, false
	}

	return file.FindFunctionByName(name)
}

func (i *index) addFile(path string) error {
	//nolint:gosec // G304 all files are from project, up to user to control this
	source, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%w: %w", errAddingFile, err)
	}

	uri, err := fileURI(path)
	if err != nil {
		return fmt.Errorf("%w: %w", errAddingFile, err)
	}

	parser := parser.New(i.adapter)
	tree, ok := parser.Parse(uri, source)

	if !ok {
		tree.Close()

		return fmt.Errorf("%w: %w", errParsing, err)
	}

	i.mu.Lock()
	i.files[uri] = tree
	i.mu.Unlock()

	return nil
}

func (i *index) getFile(uri string) *parser.ParseTree {
	i.mu.RLock()
	file, ok := i.files[uri]
	i.mu.RUnlock()

	if !ok {
		i.logger.Errorf("File not found: %s", uri)

		return nil
	}

	return file
}

func fileURI(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		//nolint:wrapcheck
		return "", err
	}

	abs = "file://" + filepath.ToSlash(abs)

	return abs, nil
}
