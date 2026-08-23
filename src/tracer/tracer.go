// Package tracer for tracing functions
package tracer

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/vulns-are-features-too/func-tracer/src/lang/index"
	"github.com/vulns-are-features-too/func-tracer/src/lang/lsp"
	"github.com/vulns-are-features-too/func-tracer/src/logging"
	"github.com/vulns-are-features-too/func-tracer/src/model"
	"github.com/vulns-are-features-too/func-tracer/src/tracer/cache"
	"github.com/vulns-are-features-too/func-tracer/src/tracer/graph"
)

var (
	errLspReferences = errors.New("failed to get LSP references")
	errFindCallers   = errors.New("findCallers failed")
	errTraceCallers  = errors.New("TraceCallers failed")
	errCtxCancelled  = errors.New("context cancelled")
)

// Tracer of functions.
type Tracer struct {
	logger  logging.Logger
	lsp     lsp.Session
	index   index.Index
	cache   *cache.Cache[[]*model.Symbol]
	workers int
}

// New tracer.
func New(
	logger logging.Logger,
	session lsp.Session,
	index index.Index,
	workers int,
) *Tracer {
	if workers < 1 {
		workers = 1
	}

	return &Tracer{
		logger:  logger,
		lsp:     session,
		index:   index,
		cache:   cache.New[[]*model.Symbol](),
		workers: workers,
	}
}

// TraceCallers traces callers of the specified target.
func (t *Tracer) TraceCallers(
	ctx context.Context,
	target *model.Symbol,
	maxDepth int,
) (*graph.Graph, error) {
	graph := graph.New()
	tasks := []task{newTask(target)}
	visited := set{target.ID: {}}

	for len(tasks) > 0 {
		t.logger.Debugf("Remaining tasks: %d", len(tasks))

		newTasks, err := t.traceCallersBatch(ctx, graph, visited, tasks, maxDepth)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errTraceCallers, err)
		}

		tasks = newTasks
	}

	return graph, nil
}

func (t *Tracer) traceCallersBatch(
	ctx context.Context,
	graph *graph.Graph,
	visited set,
	tasks []task,
	maxDepth int,
) ([]task, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("%w: %w", errCtxCancelled, ctx.Err())
	}

	results := t.runTasks(ctx, tasks, maxDepth)

	newTasks, err := nextTasksFromResults(results, graph, visited, maxDepth)
	if err != nil {
		return nil, err
	}

	return newTasks, nil
}

func (t *Tracer) runTasks(
	ctx context.Context,
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
			sem <- struct{}{}
			defer func() { <-sem }()

			callers, err := t.findCallers(ctx, currTask.symbol)
			results <- result{currTask, callers, err}
		})
	}

	wg.Wait()
	close(results)

	return results
}

func nextTasksFromResults(
	results chan result,
	graph *graph.Graph,
	visited set,
	maxDepth int,
) ([]task, error) {
	newTasks := make([]task, 0)

	for result := range results {
		res, err := nextTasksFromResult(result, graph, visited, maxDepth)
		if err != nil {
			return nil, err
		}

		newTasks = append(newTasks, res...)
	}

	return newTasks, nil
}

func nextTasksFromResult(
	res result,
	graph *graph.Graph,
	visited set,
	maxDepth int,
) ([]task, error) {
	if res.err != nil {
		return nil, res.err
	}

	newTasks := make([]task, 0)

	for _, caller := range res.callers {
		graph.AddEdge(caller, res.task.symbol)

		if maxDepth != 0 && res.task.depth >= maxDepth {
			continue
		}

		if _, ok := visited[caller.ID]; ok {
			continue
		}

		visited.add(caller.ID)
		newTasks = append(newTasks, res.task.next(caller))
	}

	return newTasks, nil
}

func (t *Tracer) findCallers(
	ctx context.Context,
	symbol *model.Symbol,
) ([]*model.Symbol, error) {
	value, err := t.cache.GetOrExec(
		symbol.ID,
		func() ([]*model.Symbol, error) { return t.findCallersInner(ctx, symbol) },
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFindCallers, err)
	}

	return value, nil
}

func (t *Tracer) findCallersInner(
	ctx context.Context,
	symbol *model.Symbol,
) ([]*model.Symbol, error) {
	refs, err := t.lsp.References(ctx, symbol.Location)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errLspReferences, err)
	}

	callers := make([]*model.Symbol, 0, len(refs))
	seen := make(set)

	for _, ref := range refs {
		caller, ok := t.index.FindFunction(ref)
		if !ok || caller.ID == symbol.ID {
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
