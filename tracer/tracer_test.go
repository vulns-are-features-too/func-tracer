package tracer_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vulns-are-features-too/func-tracer/logging"
	"github.com/vulns-are-features-too/func-tracer/model"
	"github.com/vulns-are-features-too/func-tracer/testutil/fakes"
	"github.com/vulns-are-features-too/func-tracer/tracer"
)

func TestTraceCallers(t *testing.T) {
	t.Parallel()

	logger := logging.Test(t)
	src := fakes.Source()
	tracer := tracer.New(logger, src, "", 1)

	target := symbol("target", 0)
	fn1 := symbol("fn1", 1)
	fn11 := symbol("fn11", 11)
	fn12 := symbol("fn12", 12)
	fn2 := symbol("fn2", 2)
	fn21 := symbol("fn21", 21)
	fn22 := symbol("fn22", 22)

	/*
		callees v

		          target
		         /      \
		  		fn1        fn2
				 /   \      /   \
				fn11 fn12  fn21 fn22

		callers ^
	*/
	_target := src.Func(target)
	_fn1 := src.Func(fn1).Calls(_target, 100)
	_fn2 := src.Func(fn2).Calls(_target, 200)
	src.Func(fn11).Calls(_fn1, 110)
	src.Func(fn12).Calls(_fn1, 120)
	src.Func(fn21).Calls(_fn2, 210)
	src.Func(fn22).Calls(_fn2, 220)

	graph, err := tracer.TraceCallers(t.Context(), target, 0)
	require.NoError(t, err)
	assert.NotNil(t, graph)

	assert.Equal(t, ids(fn1, fn2), ids(graph.Callers(target)...))
	assert.Equal(t, ids(fn11, fn12), ids(graph.Callers(fn1)...))
	assert.Equal(t, ids(fn21, fn22), ids(graph.Callers(fn2)...))
	assert.Empty(t, ids(graph.Callers(fn11)...))
	assert.Empty(t, ids(graph.Callers(fn12)...))
	assert.Empty(t, ids(graph.Callers(fn21)...))
	assert.Empty(t, ids(graph.Callers(fn22)...))

	hits := make([]string, 0)

	graph.WalkCallers(target, func(curr *model.Symbol, _ int) {
		hits = append(hits, curr.ID)
	})

	assert.Equal(t, ids(
		fn1, fn11, fn12,
		fn2, fn21, fn22,
	), hits)
}

func TestTraceCallees(t *testing.T) {
	t.Parallel()

	logger := logging.Test(t)
	src := fakes.Source()
	tracer := tracer.New(logger, src, "", 1)

	target := symbol("target", 0)
	fn1 := symbol("fn1", 1)
	fn11 := symbol("fn11", 11)
	fn12 := symbol("fn12", 12)
	fn2 := symbol("fn2", 2)
	fn21 := symbol("fn21", 21)
	fn22 := symbol("fn22", 22)

	/*
		callees v

				fn11 fn12  fn21 fn22
				 \   /      \   /
		  		fn1        fn2
		         \      /
		          target

		callers ^
	*/
	_fn11 := src.Func(fn11)
	_fn12 := src.Func(fn12)
	_fn21 := src.Func(fn21)
	_fn22 := src.Func(fn22)
	_fn1 := src.Func(fn1).Calls(_fn11, 110).Calls(_fn12, 120)
	_fn2 := src.Func(fn2).Calls(_fn21, 210).Calls(_fn22, 220)
	src.Func(target).Calls(_fn1, 100).Calls(_fn2, 200)

	graph, err := tracer.TraceCallees(t.Context(), target, 0)
	require.NoError(t, err)
	assert.NotNil(t, graph)

	assert.Equal(t, ids(fn1, fn2), ids(graph.Callees(target)...))
	assert.Equal(t, ids(fn11, fn12), ids(graph.Callees(fn1)...))
	assert.Equal(t, ids(fn21, fn22), ids(graph.Callees(fn2)...))
	assert.Empty(t, ids(graph.Callees(fn11)...))
	assert.Empty(t, ids(graph.Callees(fn12)...))
	assert.Empty(t, ids(graph.Callees(fn21)...))
	assert.Empty(t, ids(graph.Callees(fn22)...))

	hits := make([]string, 0)

	graph.WalkCallees(target, func(curr *model.Symbol, _ int) {
		hits = append(hits, curr.ID)
	})

	assert.Equal(t, ids(
		fn1, fn11, fn12,
		fn2, fn21, fn22,
	), hits)
}

//nolint:dupl
func TestTraceCallersWithMaxDepth(t *testing.T) {
	t.Parallel()

	const maxDepth = 2

	logger := logging.Test(t)
	src := fakes.Source()
	tracer := tracer.New(logger, src, "", 1)

	target := symbol("target", 0)
	fn1 := symbol("fn1", 1)
	fn2 := symbol("fn2", 2)
	fn3 := symbol("fn3", 3)

	// target <- fn1 <- fn2 <- fn3
	src.Func(fn3).Calls(
		src.Func(fn2).Calls(
			src.Func(fn1).Calls(
				src.Func(target),
				10),
			20),
		30)

	graph, err := tracer.TraceCallers(t.Context(), target, maxDepth)
	require.NoError(t, err)
	assert.NotNil(t, graph)

	assert.Equal(t, ids(fn1), ids(graph.Callers(target)...))
	assert.Equal(t, ids(fn2), ids(graph.Callers(fn1)...))
	assert.Empty(t, ids(graph.Callers(fn2)...))

	hits := make([]string, 0)

	graph.WalkCallers(target, func(curr *model.Symbol, _ int) {
		hits = append(hits, curr.ID)
	})

	assert.Equal(t, ids(fn1, fn2), hits)
}

//nolint:dupl
func TestTraceCalleesWithMaxDepth(t *testing.T) {
	t.Parallel()

	const maxDepth = 2

	logger := logging.Test(t)
	src := fakes.Source()
	tracer := tracer.New(logger, src, "", 1)

	target := symbol("target", 0)
	fn1 := symbol("fn1", 1)
	fn2 := symbol("fn2", 2)
	fn3 := symbol("fn3", 3)

	// target -> fn1 -> fn2 -> fn3
	src.Func(target).Calls(
		src.Func(fn1).Calls(
			src.Func(fn2).Calls(
				src.Func(fn3),
				30),
			20),
		10)

	graph, err := tracer.TraceCallees(t.Context(), target, maxDepth)
	require.NoError(t, err)
	assert.NotNil(t, graph)

	assert.Equal(t, ids(fn1), ids(graph.Callees(target)...))
	assert.Equal(t, ids(fn2), ids(graph.Callees(fn1)...))
	assert.Empty(t, ids(graph.Callees(fn2)...))

	hits := make([]string, 0)

	graph.WalkCallees(target, func(curr *model.Symbol, _ int) {
		hits = append(hits, curr.ID)
	})

	assert.Equal(t, ids(fn1, fn2), hits)
}

func TestRace(t *testing.T) {
	const (
		workers   = 8
		functions = 1000
		calls     = 10_000
	)

	t.Parallel()

	src, target := generateFunctions(functions, calls)

	t.Run("callers", func(t *testing.T) {
		t.Parallel()
		logger := logging.Test(t)
		tracer := tracer.New(logger, src, "", workers)
		graph, err := tracer.TraceCallers(t.Context(), target, 0)

		require.NoError(t, err)
		assert.NotEmpty(t, graph.Callers(target))

		hits := 0
		maxLevel := -1

		graph.WalkCallers(target, func(_ *model.Symbol, level int) {
			hits++
			maxLevel = max(level, maxLevel)
		})

		assert.Equal(t, functions, hits)
		assert.LessOrEqual(t, maxLevel, functions)
	})

	t.Run("callees", func(t *testing.T) {
		t.Parallel()
		logger := logging.Test(t)
		tracer := tracer.New(logger, src, "", workers)
		graph, err := tracer.TraceCallees(t.Context(), target, 0)

		require.NoError(t, err)
		assert.NotEmpty(t, graph.Callees(target))

		hits := 0
		maxLevel := -1

		graph.WalkCallees(target, func(_ *model.Symbol, level int) {
			hits++
			maxLevel = max(level, maxLevel)
		})

		assert.Equal(t, functions, hits)
		assert.LessOrEqual(t, maxLevel, functions)
	})
}

// generateFunctions generates functions & links them with calls
// param functions is the number of functions (besides the target)
// param calls is the min number of calls.
func generateFunctions(functions int, calls uint) (*fakes.FakeSourceProvider, *model.Symbol) {
	src := fakes.Source().WithCapacity(functions + 1)
	target := src.Func(symbol("target", 0))

	for i := range functions {
		src.Func(symbol(
			fmt.Sprintf("fn%03d", i),
			uint(i),
		))
	}

	src.AddRandomCalls(calls)
	src.LinkFuncsWithoutCalls(target)

	return src, target.Sym
}

func symbol(name string, line uint) *model.Symbol {
	pos := model.Position{Line: line, Character: line}
	loc := model.Location{
		URI: name + ".txt",
		Range: model.Range{
			Start: pos,
			End:   pos,
		},
	}

	return &model.Symbol{
		ID:       model.SymbolID(loc, name),
		Name:     name,
		Location: loc,
	}
}

func ids(symbols ...*model.Symbol) []string {
	results := make([]string, 0, len(symbols))
	for _, s := range symbols {
		results = append(results, s.ID)
	}

	return results
}
