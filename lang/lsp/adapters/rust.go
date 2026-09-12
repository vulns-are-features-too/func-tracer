package lsp_adapters

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/vulns-are-features-too/func-tracer/lang"
)

var errServerInit = errors.New("IsServerReady failed")

// RustAnalyzerLsp provides rust-analyzer.
type RustAnalyzerLsp struct{}

// RustAnalyzer LSP adapter.
func RustAnalyzer() *RustAnalyzerLsp {
	return &RustAnalyzerLsp{}
}

// Command rust-analyzer.
func (*RustAnalyzerLsp) Command() string {
	return "rust-analyzer"
}

// Args for rust-analyzer.
func (*RustAnalyzerLsp) Args() []string {
	return nil
}

// Language rust.
func (*RustAnalyzerLsp) Language() lang.Language {
	return lang.Rust
}

// GetAdditionalCapabilities provides
// LSP capabilities needed for later usage.
func (*RustAnalyzerLsp) GetAdditionalCapabilities() map[string]any {
	return map[string]any{
		"experimental": map[string]any{
			"serverStatusNotification": true,
		},
	}
}

// WaitServerNotificationMethod returns the
// "experimental/serverStatus" method for rust-analyzer.
func (*RustAnalyzerLsp) WaitServerNotificationMethod() string {
	return "experimental/serverStatus"
}

// IsServerReady checks `quiescent` in the JSON.
func (*RustAnalyzerLsp) IsServerReady(params json.RawMessage) (bool, error) {
	p := struct {
		Quiescent bool `json:"quiescent"`
	}{Quiescent: false}

	err := json.Unmarshal(params, &p)
	if err != nil {
		return false, fmt.Errorf("%w: %w", errServerInit, err)
	}

	return p.Quiescent, nil
}
