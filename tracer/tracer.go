// Package tracer for tracing functions
package tracer

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	"github.com/vulns-are-features-too/func-tracer/logging"
	"github.com/vulns-are-features-too/func-tracer/model"
	"github.com/vulns-are-features-too/func-tracer/tracer/cache"
	"github.com/vulns-are-features-too/func-tracer/tracer/graph"
	"github.com/vulns-are-features-too/func-tracer/tracer/source"
)

var (
	errLspReferences  = errors.New("failed to get LSP references")
	errLspDefinitions = errors.New("failed to get LSP definitions")
	errFindCallers    = errors.New("findCallers failed")
	errFindCallees    = errors.New("findCallees failed")
	errTrace          = errors.New("trace failed")
	errCtxCancelled   = errors.New("context cancelled")
)

type (
	traceFunc         func(ctx context.Context, symbol *model.Symbol) ([]*model.Symbol, error)
	collectResultFunc func(graph *graph.Graph, input *model.Symbol, output *model.Symbol)
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
	collect := func(g *graph.Graph, in *model.Symbol, out *model.Symbol) {
		g.AddEdge(out, in)
	}
	g := graph.CallersOnly()

	return t.trace(ctx, g, t.findCallers, collect, target, maxDepth)
}

// TraceCallees traces callees of the specified target.
func (t *Tracer) TraceCallees(
	ctx context.Context,
	target *model.Symbol,
	maxDepth int,
) (*graph.Graph, error) {
	collect := func(g *graph.Graph, in *model.Symbol, out *model.Symbol) {
		g.AddEdge(in, out)
	}
	g := graph.CalleesOnly()

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
	tasks := []task{newTask(target)}
	visited := set{target.ID: {}}

	for len(tasks) > 0 {
		t.logger.Debugf("Remaining tasks: %d", len(tasks))

		newTasks, err := t.traceBatch(ctx, fnTrace, collectResult, graph, visited, tasks, maxDepth)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errTrace, err)
		}

		tasks = newTasks
	}

	return graph, nil
}

func (t *Tracer) traceBatch(
	ctx context.Context,
	fnTrace traceFunc,
	collectResult collectResultFunc,
	graph *graph.Graph,
	visited set,
	tasks []task,
	maxDepth int,
) ([]task, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("%w: %w", errCtxCancelled, ctx.Err())
	}

	results := t.runTasks(ctx, fnTrace, tasks, maxDepth)

	newTasks, err := nextTasksFromResults(results, collectResult, graph, visited, maxDepth)
	if err != nil {
		return nil, err
	}

	return newTasks, nil
}

func (t *Tracer) runTasks(
	ctx context.Context,
	fnTrace traceFunc,
	tasks []task,
	maxDepth int,
) chan result {
	sem := make(chan struct{}, t.workers)

	var wg sync.WaitGroup

	results := make(chan result, len(tasks))
	for _, currTask := range tasks {
		if maxDepth != 0 && currTask.depth > maxDepth {
			continue
		}

		wg.Go(func() {
			select {
			case <-ctx.Done():
				return
			case sem <- struct{}{}:
				defer func() { <-sem }()

				traceResults, err := fnTrace(ctx, currTask.symbol)
				results <- result{currTask, traceResults, err}
			}
		})
	}

	wg.Wait()
	close(results)

	return results
}

func nextTasksFromResults(
	results chan result,
	collectResult collectResultFunc,
	graph *graph.Graph,
	visited set,
	maxDepth int,
) ([]task, error) {
	newTasks := make([]task, 0)

	for result := range results {
		res, err := nextTasksFromResult(result, collectResult, graph, visited, maxDepth)
		if err != nil {
			return nil, err
		}

		newTasks = append(newTasks, res...)
	}

	return newTasks, nil
}

func nextTasksFromResult(
	res result,
	collectResult collectResultFunc,
	graph *graph.Graph,
	visited set,
	maxDepth int,
) ([]task, error) {
	if res.err != nil {
		return nil, res.err
	}

	newTasks := make([]task, 0)

	for _, fn := range res.funcs {
		collectResult(graph, res.task.symbol, fn)

		if maxDepth != 0 && res.task.depth >= maxDepth {
			continue
		}

		if _, ok := visited[fn.ID]; ok {
			continue
		}

		visited.add(fn.ID)
		newTasks = append(newTasks, res.task.next(fn))
	}

	return newTasks, nil
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
	seen := make(set)

	for _, ref := range refs {
		caller, ok := t.src.FindFunction(ref)
		if !ok || caller.ID == funcDef.ID {
			continue
		}

		if _, ok := seen[caller.ID]; ok {
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

	seen := make(set)
	seen.add(funcDef.ID)

	for _, callee := range t.src.FindFunctionCalls(funcDef.Location) {
		t.logger.Debugf("Found callee `%s` at %s", callee.Name, callee.Location)

		if _, ok := seen[callee.ID]; ok {
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
