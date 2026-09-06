// Package lsp provides LSP integration
package lsp

import (
	"encoding/json"

	"github.com/vulns-are-features-too/func-tracer/lang"
)

// Adapter for LSP integration.
type Adapter interface {
	Command() string
	Args() []string
	Language() lang.Language
}

// HasAdditionalCapabilities marks
// LSP server with additional capabilities.
type HasAdditionalCapabilities interface {
	GetAdditionalCapabilities() map[string]any
}

// WaitServerInit marks
// LSP servers whose init we need to wait for.
type WaitServerInit interface {
	WaitServerNotificationMethod() string
	IsServerReady(params json.RawMessage) (bool, error)
}
