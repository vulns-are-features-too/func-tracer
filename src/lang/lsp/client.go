package lsp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/vulns-are-features-too/func-tracer/src/logging"
)

var (
	errLspServerNotInstalled = errors.New("LSP server not installed")
	errStdinFailed           = errors.New("failed to get stdin")
	errStderrFailed          = errors.New("failed to get stderr")
	errLspStartFailed        = errors.New("failed to start LSP server")
	errRequestFailed         = errors.New("request failed")
	errMessageWriteFailed    = errors.New("failed to write message")
	errMessageReadFailed     = errors.New("failed to read message")
	errNotificationFailed    = errors.New("notification failed")
	errCtxCancelled          = errors.New("context cancelled")
	errClientStopped         = errors.New("lsp client stopped")
	errLsp                   = errors.New("LSP server error")
)

const jsonRPCVer = "2.0"

type client struct {
	logger        logging.Logger
	cmd           *exec.Cmd
	stdin         io.WriteCloser
	stdout        io.ReadCloser
	writeMu       sync.Mutex
	nextID        atomic.Int64
	pendingMu     sync.Mutex
	pending       map[int64]chan response
	done          chan struct{}
	notifications map[string]chan json.RawMessage
}

type response struct {
	result json.RawMessage
	err    error
}

type envelope struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func startClient(
	ctx context.Context,
	log logging.Logger,
	command string,
	args []string,
) (*client, error) {
	if _, err := exec.LookPath(command); err != nil {
		return nil, fmt.Errorf("%w: %s", errLspServerNotInstalled, command)
	}

	//nolint:gosec // G204 cmd and args are hard-coded
	cmd := exec.CommandContext(ctx, command, args...)

	cmd.Stderr = cmd.Stdout

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errStdinFailed, err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()

		return nil, fmt.Errorf("%w: %w", errStderrFailed, err)
	}

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()

		return nil, fmt.Errorf("%w: %w", errLspStartFailed, err)
	}

	c := &client{
		logger:        log,
		cmd:           cmd,
		stdin:         stdin,
		stdout:        stdout,
		pending:       make(map[int64]chan response),
		done:          make(chan struct{}),
		notifications: make(map[string]chan json.RawMessage),
	}

	go c.readLoop()

	return c, nil
}

// Listen for notifications from server (not necessary paired with requests).
func (c *client) subscribeNotification(method string) chan json.RawMessage {
	if ch, ok := c.notifications[method]; ok {
		return ch
	}

	ch := make(chan json.RawMessage)
	c.notifications[method] = ch

	return ch
}

func (c *client) unsubscribeNotification(method string) {
	delete(c.notifications, method)
}

func (c *client) request(
	ctx context.Context,
	method string,
	params any,
) (json.RawMessage, error) {
	id := c.nextID.Add(1)

	wait := make(chan response, 1)

	c.pendingMu.Lock()
	c.pending[id] = wait
	c.pendingMu.Unlock()

	message, err := json.Marshal(struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int64  `json:"id"`
		Method  string `json:"method"`
		Params  any    `json:"params,omitempty"`
	}{
		JSONRPC: jsonRPCVer,
		ID:      id,
		Method:  method,
		Params:  params,
	})
	if err != nil {
		c.removePending(id)

		return nil, fmt.Errorf("%w: %w", errRequestFailed, err)
	}

	if err := c.write(message); err != nil {
		c.removePending(id)

		return nil, fmt.Errorf("%w: %w", errRequestFailed, err)
	}

	select {
	case result := <-wait:
		return result.result, result.err

	case <-ctx.Done():
		c.removePending(id)

		return nil, fmt.Errorf("%w: %w", errCtxCancelled, ctx.Err())

	case <-c.done:
		c.removePending(id)

		return nil, errClientStopped
	}
}

func (c *client) notify(method string, params any) error {
	message, err := json.Marshal(struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  any    `json:"params,omitempty"`
	}{
		JSONRPC: jsonRPCVer,
		Method:  method,
		Params:  params,
	})
	if err != nil {
		return fmt.Errorf("%w: %w", errNotificationFailed, err)
	}

	return c.write(message)
}

func (c *client) write(message []byte) error {
	header := []byte(
		"Content-Length: " + strconv.Itoa(len(message)) + "\r\n\r\n",
	)

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if _, err := c.stdin.Write(header); err != nil {
		return fmt.Errorf("%w: %w", errMessageWriteFailed, err)
	}

	c.logger.Debugf("Writing request: %s", string(message))

	if _, err := c.stdin.Write(message); err != nil {
		return fmt.Errorf("%w: %w", errMessageWriteFailed, err)
	}

	return nil
}

func (c *client) readLoop() {
	reader := bufio.NewReader(c.stdout)

	defer close(c.done)

	for {
		c.readOnce(reader)
	}
}

func (c *client) readOnce(reader *bufio.Reader) {
	message, err := readMessage(reader)
	if err != nil {
		return
	}

	c.logger.Debugf("Reading response: %s", string(message))

	var envelope envelope
	if err := json.Unmarshal(message, &envelope); err != nil {
		return
	}

	method := envelope.Method
	if ch, ok := c.notifications[method]; ok {
		ch <- envelope.Params

		return
	}

	if len(envelope.ID) == 0 {
		return
	}

	var id int64
	if err := json.Unmarshal(envelope.ID, &id); err != nil {
		return
	}

	if envelope.Method == "workspace/configuration" {
		if c.handleWorkspaceConfiguration(envelope, id) {
			return
		}
	}

	c.pendingMu.Lock()
	wait := c.pending[id]
	delete(c.pending, id)
	c.pendingMu.Unlock()

	if wait == nil {
		return
	}

	if envelope.Error != nil {
		wait <- response{
			err: fmt.Errorf("%w %d: %s", errLsp, envelope.Error.Code, envelope.Error.Message),
		}

		return
	}

	wait <- response{
		result: envelope.Result,
	}
}

func readMessage(reader *bufio.Reader) ([]byte, error) {
	contentLength := -1

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errMessageReadFailed, err)
		}

		line = strings.TrimRight(line, "\r\n")

		if line == "" {
			break
		}

		const prefix = "Content-Length:"
		if len(line) >= len(prefix) && line[:len(prefix)] == prefix {
			value := bytes.TrimSpace([]byte(line[len(prefix):]))

			length, err := strconv.Atoi(string(value))
			if err != nil {
				return nil, fmt.Errorf("%w: %w", errMessageReadFailed, err)
			}

			contentLength = length
		}
	}

	if contentLength < 0 {
		return nil, fmt.Errorf("%w: missing Content-Length", errMessageReadFailed)
	}

	message := make([]byte, contentLength)
	if _, err := io.ReadFull(reader, message); err != nil {
		return message, fmt.Errorf("%w: %w", errMessageReadFailed, err)
	}

	return message, nil
}

func (c *client) removePending(id int64) {
	c.pendingMu.Lock()
	delete(c.pending, id)
	c.pendingMu.Unlock()
}

func (c *client) close() {
	_ = c.notify("exit", nil)
	_ = c.stdin.Close()

	if c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}
}

func (c *client) handleWorkspaceConfiguration(e envelope, id int64) bool {
	// currently hard-coded for gopls
	var params struct {
		Items []json.RawMessage `json:"items,omitempty"`
	}
	if err := json.Unmarshal(e.Params, &params); err != nil {
		return false
	}

	results := make([]map[string]any, len(params.Items))
	for i := range results {
		results[i] = map[string]any{}
	}

	msg, err := json.Marshal(struct {
		JSONRPC string           `json:"jsonrpc"`
		ID      int64            `json:"id,omitempty"`
		Result  []map[string]any `json:"result"`
	}{
		JSONRPC: jsonRPCVer,
		ID:      id,
		Result:  results,
	})
	if err != nil {
		return false
	}

	if err := c.write(msg); err != nil {
		c.logger.Errorf("Error sending workspace configuration request: %s", err.Error())
	}

	return true
}
