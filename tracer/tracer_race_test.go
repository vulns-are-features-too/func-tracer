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
