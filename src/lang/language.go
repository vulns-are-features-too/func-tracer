// Package lang provides a list of supported languages
package lang

// Language type.
type Language string

// Supported languages.
const (
	Go   Language = "go"
	Rust Language = "rust"
)

// SupportedLanguages is the list of supported languages.
//
//nolint:gochecknoglobals
var SupportedLanguages = []Language{Go, Rust}
