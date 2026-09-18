// Package source provides a wrapper for getting source code info
package source

import (
	"context"
	"errors"
	"fmt"

	"github.com/vulns-are-features-too/func-tracer/lang/index"
	"github.com/vulns-are-features-too/func-tracer/lang/lsp"
	"github.com/vulns-are-features-too/func-tracer/model"
)

var errLsp = errors.New("LSP error")

// Provider provides information about source code.
type Provider interface {
	FindFunction(location model.Location) (model.Symbol, bool)
	FindFunctionByName(uri string, name string) []model.Symbol
	FindFunctionCalls(location model.Location) []model.Symbol

	References(
		ctx context.Context,
		funcDef model.Location,
	) ([]model.Location, error)

	Definitions(
		ctx context.Context,
		callee model.Location,
	) ([]model.Location, error)

	Close(ctx context.Context)
}

// LocalSourceProvider for local files.
type LocalSourceProvider struct {
	lsp   lsp.Session
	index index.Index
}

// Local files source provider.
func Local(lsp lsp.Session, index index.Index) *LocalSourceProvider {
	return &LocalSourceProvider{lsp, index}
}

// FindFunction wraps the same index method.
func (src *LocalSourceProvider) FindFunction(location model.Location) (model.Symbol, bool) {
	return src.index.FindFunction(location)
}

// FindFunctionByName wraps the same index method.
func (src *LocalSourceProvider) FindFunctionByName(uri string, name string) []model.Symbol {
	return src.index.FindFunctionByName(uri, name)
}

// FindFunctionCalls wraps the same index method.
func (src *LocalSourceProvider) FindFunctionCalls(location model.Location) []model.Symbol {
	return src.index.FindFunctionCalls(location)
}

// References wraps the same LSP method.
func (src *LocalSourceProvider) References(
	ctx context.Context,
	funcDef model.Location,
) ([]model.Location, error) {
	res, err := src.lsp.References(ctx, funcDef)
	if err != nil {
		return []model.Location{}, fmt.Errorf("%w: %w", errLsp, err)
	}

	return res, nil
}

// Definitions wraps the same LSP method.
func (src *LocalSourceProvider) Definitions(
	ctx context.Context,
	callee model.Location,
) ([]model.Location, error) {
	res, err := src.lsp.Definitions(ctx, callee)
	if err != nil {
		return []model.Location{}, fmt.Errorf("%w: %w", errLsp, err)
	}

	return res, nil
}

// Close index & LSP server.
func (src *LocalSourceProvider) Close(ctx context.Context) {
	src.index.Close()
	src.lsp.Close(ctx)
}
