package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/vulns-are-features-too/func-tracer/src/lang"
	"github.com/vulns-are-features-too/func-tracer/src/logging"
	"github.com/vulns-are-features-too/func-tracer/src/model"
)

var (
	errStart      = errors.New("failed to start LSP server")
	errInit       = errors.New("failed to initialize LSP server")
	errReferences = errors.New("References() failed")
	errJSONDecode = errors.New("JSON decoding failed")
)

//nolint:gochecknoglobals
var (
	defaultCapabilities = map[string]any{
		"workspace": map[string]any{
			"configuration": true,
		},
	}
	clientInfo = map[string]any{
		"name":    "func-tracer",
		"version": "0.1.0",
	}
)

// Session with LSP server.
type Session interface {
	References(
		ctx context.Context,
		location model.Location,
	) ([]model.Location, error)

	Close(ctx context.Context)
}

type session struct {
	logger   logging.Logger
	client   *client
	adapter  Adapter
	rootURI  string
	language lang.Language
}

// Start LSP server.
//
//nolint:revive // unexported-return
func Start(
	ctx context.Context,
	logger logging.Logger,
	adapter Adapter,
	root string,
) (*session, error) {
	client, err := startClient(ctx, logger, adapter.Command(), adapter.Args())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errStart, err)
	}

	rootURI, err := pathToURI(root)
	if err != nil {
		client.close()

		return nil, fmt.Errorf("%w: %w", errStart, err)
	}

	s := &session{
		logger:   logger,
		client:   client,
		adapter:  adapter,
		rootURI:  rootURI,
		language: adapter.Language(),
	}

	if err := s.initialize(ctx); err != nil {
		s.Close(ctx)

		return nil, fmt.Errorf("%w: %w", errStart, err)
	}

	return s, nil
}

// References of function at location.
func (s *session) References(
	ctx context.Context,
	location model.Location,
) ([]model.Location, error) {
	params := map[string]any{
		"textDocument": map[string]any{
			"uri": location.URI,
		},
		"position": map[string]any{
			"line":      location.Range.Start.Line,
			"character": location.Range.Start.Character,
		},
		"context": map[string]any{
			"includeDeclaration": false,
		},
	}

	result, err := s.client.request(
		ctx,
		"textDocument/references",
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errReferences, err)
	}

	if string(result) == "null" || len(result) == 0 {
		return nil, nil
	}

	var locations []model.Location
	if err := json.Unmarshal(result, &locations); err != nil {
		return nil, fmt.Errorf("%w: %w", errJSONDecode, err)
	}

	references := make([]model.Location, 0, len(locations))
	references = append(references, locations...)

	return references, nil
}

// Close signals LSP server to shutdown and closes connection.
func (s *session) Close(ctx context.Context) {
	if s.client == nil {
		return
	}

	_, _ = s.client.request(
		ctx,
		"shutdown",
		nil,
	)

	s.client.close()
}

func (s *session) initialize(ctx context.Context) error {
	capabilities := defaultCapabilities

	if init, ok := s.adapter.(HasAdditionalCapabilities); ok {
		maps.Copy(capabilities, init.GetAdditionalCapabilities())
	}

	params := map[string]any{
		"processId":  os.Getpid(),
		"clientInfo": clientInfo,
		"rootUri":    s.rootURI,
		"workspaceFolders": []map[string]any{
			{
				"uri":  s.rootURI,
				"name": filepath.Base(strings.TrimPrefix(s.rootURI, "file://")),
			},
		},
		"capabilities": capabilities,
	}

	sendInit := func() error {
		_, err := s.client.request(ctx, "initialize", params)
		if err != nil {
			return fmt.Errorf("%w: %w", errInit, err)
		}

		err = s.client.notify("initialized", map[string]any{})
		if err != nil {
			return fmt.Errorf("%w: %w", errInit, err)
		}

		return nil
	}

	if init, ok := s.adapter.(WaitServerInit); ok {
		err := s.initWithWait(init, sendInit)
		if err != nil {
			return err
		}
	} else {
		if err := sendInit(); err != nil {
			return fmt.Errorf("%w: %w", errInit, err)
		}
	}

	s.logger.Infof("LSP server initialized")

	return nil
}

func (s *session) initWithWait(
	init WaitServerInit,
	sendInit func() error,
) error {
	s.logger.Infof("Waiting for server to initialize")

	wait := s.client.subscribeNotification(init.WaitServerNotificationMethod())

	if err := sendInit(); err != nil {
		return fmt.Errorf("%w: %w", errInit, err)
	}

	ready := false
	for !ready {
		noti := <-wait
		r, err := init.IsServerReady(noti)
		ready = r

		if err != nil {
			s.logger.Errorf("%w: %w", errInit, err)
		}
	}

	s.client.unsubscribeNotification(init.WaitServerNotificationMethod())

	return nil
}

func pathToURI(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		//nolint:wrapcheck
		return "", err
	}

	absolute = filepath.ToSlash(absolute)

	return (&url.URL{
		Scheme: "file",
		Path:   absolute,
	}).String(), nil
}
