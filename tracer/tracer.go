// Package tracer for tracing functions
package tracer

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/vulns-are-features-too/func-tracer/logging"
	"github.com/vulns-are-features-too/func-tracer/model"
	"github.com/vulns-are-features-too/func-tracer/tracer/cache"
	"github.com/vulns-are-features-too/func-tracer/tracer/executor"
	"github.com/vulns-are-features-too/func-tracer/tracer/graph"
	"github.com/vulns-are-features-too/func-tracer/tracer/source"
)

var (
	errLspReferences  = errors.New("failed to get LSP references")
	errLspDefinitions = errors.New("failed to get LSP definitions")
	errFindCallers    = errors.New("findCallers failed")
	errFindCallees    = errors.New("findCallees failed")
)

type (
	traceFunc         func(ctx context.Context, symbol *model.Symbol) ([]*model.Symbol, error)
	collectResultFunc func(input *model.Symbol, output *model.Symbol)
)

// Tracer of functions.
type Tracer struct {
	logger  logging.Logger
	src     source.Provider
	cache   *cache.Cache[[]*model.Symbol]
	root    string
	workers int
}

// New tracer.
func New(
	logger logging.Logger,
	src source.Provider,
	root string,
	workers int,
) *Tracer {
	if workers < 1 {
		workers = 1
	}

	return &Tracer{
		logger:  logger,
		src:     src,
		cache:   cache.New[[]*model.Symbol](),
		root:    root,
		workers: workers,
	}
}

// TraceCallers traces callers of the specified target.
func (t *Tracer) TraceCallers(
	ctx context.Context,
	target *model.Symbol,
	maxDepth int,
) (*graph.Graph, error) {
	g := graph.CallersOnly()
	collect := func(in *model.Symbol, out *model.Symbol) {
		g.AddEdge(out, in)
	}

	return t.trace(ctx, g, t.findCallers, collect, target, maxDepth)
}

// TraceCallees traces callees of the specified target.
func (t *Tracer) TraceCallees(
	ctx context.Context,
	target *model.Symbol,
	maxDepth int,
) (*graph.Graph, error) {
	g := graph.CalleesOnly()
	collect := func(in *model.Symbol, out *model.Symbol) {
		g.AddEdge(in, out)
	}

	return t.trace(ctx, g, t.findCallees, collect, target, maxDepth)
}

func (t *Tracer) trace(
	ctx context.Context,
	graph *graph.Graph,
	fnTrace traceFunc,
	collectResult collectResultFunc,
	target *model.Symbol,
	maxDepth int,
) (*graph.Graph, error) {
	visited := newSet()
	visited.add(target.ID)

	workerFactory := func() worker {
		return worker{
			ctx,
			t.logger,
			visited,
			maxDepth,
			fnTrace,
			collectResult,
		}
	}

	executor := executor.New(ctx, t.workers, workerFactory)
	executor.Enqueue(newTask(target))
	executor.Run()

	return graph, nil
}

func (t *Tracer) findCallers(
	ctx context.Context,
	funcDef *model.Symbol,
) ([]*model.Symbol, error) {
	value, err := t.cache.GetOrExec(
		funcDef.ID,
		func() ([]*model.Symbol, error) { return t.findCallersInner(ctx, funcDef) },
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFindCallers, err)
	}

	return value, nil
}

func (t *Tracer) findCallersInner(
	ctx context.Context,
	funcDef *model.Symbol,
) ([]*model.Symbol, error) {
	refs, err := t.src.References(ctx, funcDef.Location)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errLspReferences, err)
	}

	callers := make([]*model.Symbol, 0, len(refs))
	seen := newSet()

	for _, ref := range refs {
		caller, ok := t.src.FindFunction(ref)
		if !ok || caller.ID == funcDef.ID {
			continue
		}

		if seen.has(caller.ID) {
			continue
		}

		seen.add(caller.ID)
		callers = append(callers, &caller)
	}

	return callers, nil
}

func (t *Tracer) findCallees(
	ctx context.Context,
	funcDef *model.Symbol,
) ([]*model.Symbol, error) {
	value, err := t.cache.GetOrExec(
		funcDef.ID,
		func() ([]*model.Symbol, error) { return t.findCalleesInner(ctx, funcDef) },
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFindCallees, err)
	}

	return value, nil
}

func (t *Tracer) findCalleesInner(
	ctx context.Context,
	funcDef *model.Symbol,
) ([]*model.Symbol, error) {
	results := make([]*model.Symbol, 0)

	seen := newSet()
	seen.add(funcDef.ID)

	for _, callee := range t.src.FindFunctionCalls(funcDef.Location) {
		t.logger.Debugf("Found callee `%s` at %s", callee.Name, callee.Location)

		if seen.has(callee.ID) {
			continue
		}

		seen.add(callee.ID)

		defs, err := t.src.Definitions(ctx, callee.Location)
		if err != nil {
			t.logger.Errorf("%w: %w", errLspDefinitions, err)

			continue
		}

		for _, def := range defs {
			t.logger.Debugf("%s defined @ %s", callee.Name, def.String())

			if !t.isFileInProject(def.URI) {
				t.logger.Infof("skipping `%s` due to being outside project root", def.URI)

				continue
			}

			results = append(results, &model.Symbol{
				ID:       model.SymbolID(def, callee.Name),
				Name:     callee.Name,
				Location: def,
			})
		}
	}

	return results, nil
}

func (t *Tracer) isFileInProject(uri string) bool {
	if t.root == "" {
		return true
	}

	u, err := url.Parse(uri)
	if err != nil || u.Scheme != "file" {
		return false
	}

	filePath, err := url.PathUnescape(u.Path)
	if err != nil {
		return false
	}

	root, err := filepath.Abs(t.root)
	if err != nil {
		return false
	}

	filePath, err = filepath.Abs(filePath)
	if err != nil {
		return false
	}

	rel, err := filepath.Rel(root, filePath)
	if err != nil {
		return false
	}

	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
