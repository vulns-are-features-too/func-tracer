// Package registry provides a registry of supported languages
package registry

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/vulns-are-features-too/func-tracer/lang"
	"github.com/vulns-are-features-too/func-tracer/lang/lsp"
	lsp_adapters "github.com/vulns-are-features-too/func-tracer/lang/lsp/adapters"
	"github.com/vulns-are-features-too/func-tracer/lang/parser"
	parser_adapters "github.com/vulns-are-features-too/func-tracer/lang/parser/adapters"
)

var (
	errUnsupportedFileExtension = errors.New("unsupported file extension")
	errUnsupportedLanguage      = errors.New("unsupported language")
)

//nolint:gochecknoglobals
var extensions = map[lang.Language][]string{
	lang.Go:   {".go"},
	lang.Rust: {".rs"},
}

// ListLanguages lists supported languages.
func ListLanguages() []lang.Language {
	return lang.SupportedLanguages
}

// Detect language based on file name.
func Detect(path string) (lang.Language, error) {
	ext := strings.ToLower(filepath.Ext(path))

	for lang, exts := range extensions {
		if slices.Contains(exts, ext) {
			return lang, nil
		}
	}

	return "", fmt.Errorf("%w: %s", errUnsupportedFileExtension, ext)
}

// Extensions matching the language.
func Extensions(lang lang.Language) ([]string, bool) {
	res, ok := extensions[lang]

	return res, ok
}

// Parser for the language.
//
//nolint:ireturn
func Parser(l lang.Language) (parser.Adapter, error) {
	switch l {
	case lang.Go:
		return parser_adapters.Go(), nil
	case lang.Rust:
		return parser_adapters.Rust(), nil
	default:
		return nil, fmt.Errorf("%w: %q", errUnsupportedLanguage, l)
	}
}

// ListLSP lists supported languages and their LSP servers.
func ListLSP() map[lang.Language][]lsp.Adapter {
	return map[lang.Language][]lsp.Adapter{
		lang.Go:   {lsp_adapters.Go()},
		lang.Rust: {lsp_adapters.Rust()},
	}
}

// LSP for the language.
//
//nolint:ireturn
func LSP(l lang.Language) (lsp.Adapter, error) {
	lsps, ok := ListLSP()[l]
	if !ok {
		return nil, fmt.Errorf("%w: %q", errUnsupportedLanguage, l)
	}

	return lsps[0], nil
}
