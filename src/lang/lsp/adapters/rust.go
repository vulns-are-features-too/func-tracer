package lsp_adapters

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/vulns-are-features-too/func-tracer/src/lang"
)

var errServerInit = errors.New("IsServerReady failed")

// RustLsp provides rust-analyzer.
type RustLsp struct{}

// Rust LSP adapter.
func Rust() *RustLsp {
	return &RustLsp{}
}

// Command rust-analyzer.
func (*RustLsp) Command() string {
	return "rust-analyzer"
}

// Args for rust-analyzer.
func (*RustLsp) Args() []string {
	return nil
}

// Language rust.
func (*RustLsp) Language() lang.Language {
	return lang.Rust
}

// GetAdditionalCapabilities provides
// LSP capabilities needed for later usage.
func (*RustLsp) GetAdditionalCapabilities() map[string]any {
	return map[string]any{
		"experimental": map[string]any{
			"serverStatusNotification": true,
		},
	}
}

// WaitServerNotificationMethod returns the
// "experimental/serverStatus" method for rust-analyzer.
func (*RustLsp) WaitServerNotificationMethod() string {
	return "experimental/serverStatus"
}

// IsServerReady checks `quiescent` in the JSON.
func (*RustLsp) IsServerReady(params json.RawMessage) (bool, error) {
	p := struct {
		Quiescent bool `json:"quiescent"`
	}{Quiescent: false}

	err := json.Unmarshal(params, &p)
	if err != nil {
		return false, fmt.Errorf("%w: %w", errServerInit, err)
	}

	return p.Quiescent, nil
}
