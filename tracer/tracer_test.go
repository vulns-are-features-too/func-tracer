package tracer_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vulns-are-features-too/func-tracer/lang/index"
	"github.com/vulns-are-features-too/func-tracer/lang/lsp"
	"github.com/vulns-are-features-too/func-tracer/logging"
	"github.com/vulns-are-features-too/func-tracer/model"
	"github.com/vulns-are-features-too/func-tracer/testutil/fakes"
	"github.com/vulns-are-features-too/func-tracer/tracer"
)

func TestTraceCallers(t *testing.T) {
	t.Parallel()

	logger := logging.Test(t)
	index := fakes.Index()
	session := fakes.Session()
	tracer := tracer.New(logger, session, index, 1)

	target := symbol("target", 0)
	fn1 := symbol("fn1", 1)
	fn11 := symbol("fn11", 11)
	fn12 := symbol("fn12", 12)
	fn2 := symbol("fn2", 2)
	fn21 := symbol("fn21", 21)
	fn22 := symbol("fn22", 22)

	index.Add(target, fn1, fn11, fn12, fn2, fn21, fn22)

	/*
		          target
		         /      \
		  		fn1        fn2
				 /   \      /   \
				fn11 fn12  fn21 fn22
	*/
	session.AddCall(fn1, target)
	session.AddCall(fn2, target)
	session.AddCall(fn11, fn1)
	session.AddCall(fn12, fn1)
	session.AddCall(fn21, fn2)
	session.AddCall(fn22, fn2)

	graph, err := tracer.TraceCallers(t.Context(), target, 0)
	require.NoError(t, err)
	assert.NotNil(t, graph)

	assert.Equal(t, []*model.Symbol{fn1, fn2}, graph.Callers(target))
	assert.Equal(t, []*model.Symbol{fn11, fn12}, graph.Callers(fn1))
	assert.Equal(t, []*model.Symbol{fn21, fn22}, graph.Callers(fn2))
	assert.Equal(t, []*model.Symbol{}, graph.Callers(fn11))
	assert.Equal(t, []*model.Symbol{}, graph.Callers(fn12))
	assert.Equal(t, []*model.Symbol{}, graph.Callers(fn21))
	assert.Equal(t, []*model.Symbol{}, graph.Callers(fn22))

	hits := make([]string, 0)

	graph.WalkCallers(target, func(curr *model.Symbol, _ int) {
		hits = append(hits, curr.Name)
	})

	assert.Equal(t, []string{
		fn1.Name, fn11.Name, fn12.Name,
		fn2.Name, fn21.Name, fn22.Name,
	}, hits)
}

func TestTraceCallersWithMaxDepth(t *testing.T) {
	t.Parallel()

	const maxDepth = 2

	logger := logging.Test(t)
	index := fakes.Index()
	session := fakes.Session()
	tracer := tracer.New(logger, session, index, 1)

	target := symbol("target", 0)
	fn1 := symbol("fn1", 1)
	fn2 := symbol("fn2", 2)
	fn3 := symbol("fn3", 3)

	index.Add(target, fn1, fn2, fn3)

	session.AddCall(fn1, target)
	session.AddCall(fn2, fn1)
	session.AddCall(fn3, fn2)

	graph, err := tracer.TraceCallers(t.Context(), target, maxDepth)
	require.NoError(t, err)
	assert.NotNil(t, graph)

	assert.Equal(t, []*model.Symbol{fn1}, graph.Callers(target))
	assert.Equal(t, []*model.Symbol{fn2}, graph.Callers(fn1))
	assert.Equal(t, []*model.Symbol{}, graph.Callers(fn2))

	hits := make([]string, 0)

	graph.WalkCallers(target, func(curr *model.Symbol, _ int) {
		hits = append(hits, curr.Name)
	})

	assert.Equal(t, []string{fn1.Name, fn2.Name}, hits)
}

func TestRace(t *testing.T) {
	const (
		workers   = 8
		functions = 1000
		calls     = 10_000
	)

	t.Parallel()

	logger := logging.Test(t)
	index, session, target := generateFunctions(functions, calls)
	tracer := tracer.New(logger, session, index, workers)

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
}

//nolint:ireturn
func generateFunctions(functions int, calls int) (index.Index, lsp.Session, *model.Symbol) {
	index := fakes.Index().WithCapacity(functions + 1)
	session := fakes.Session()
	target := symbol("target", 0)
	index.Add(target)

	for i := range functions {
		sym := symbol(fmt.Sprintf("fn%03d", i), uint(i))
		index.Add(sym)
	}

	for range calls {
		session.AddCall(index.Random(), index.Random())
	}

	for _, c := range index.All() {
		if !session.HasCallee(c) {
			session.AddCall(target, c)
		}

		if !session.HasCaller(c) {
			session.AddCall(c, target)
		}
	}

	return index, session, target
}

func symbol(name string, line uint) *model.Symbol {
	pos := model.Position{Line: line, Character: line}

	return &model.Symbol{
		ID:   name,
		Name: name,
		Location: model.Location{
			URI: name,
			Range: model.Range{
				Start: pos,
				End:   pos,
			},
		},
	}
}
