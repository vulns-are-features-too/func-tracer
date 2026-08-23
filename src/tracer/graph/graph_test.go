//go:build !race

package graph_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vulns-are-features-too/func-tracer/src/model"
	"github.com/vulns-are-features-too/func-tracer/src/tracer/graph"
)

type visitEntry struct {
	name  string
	level int
}

var (
	target = symbol("target")
	fn1    = symbol("fn1")
	fn11   = symbol("fn11")
	fn12   = symbol("fn12")
	fn2    = symbol("fn2")
	fn21   = symbol("fn21")
	fn22   = symbol("fn22")
	fn3    = symbol("fn3")
)

func TestWalkCallersLinear(t *testing.T) {
	t.Parallel()

	// arrange
	g := graph.New()
	g.AddEdge(fn1, target)
	g.AddEdge(fn2, fn1)

	visits := make([]visitEntry, 0)

	// act
	g.WalkCallers(target, func(curr *model.Symbol, level int) {
		visits = append(visits, visitEntry{curr.Name, level})
	})

	// assert
	assert.Equal(t, []visitEntry{
		{fn1.Name, 1},
		{fn2.Name, 2},
	}, visits)
}

func TestWalkCallersTree(t *testing.T) {
	t.Parallel()

	// arrange
	/*
		          target
		         /      \
		  		fn1        fn2
				 /   \      /   \
				fn11 fn12  fn21 fn22
	*/
	g := graph.New()
	g.AddEdge(fn1, target)
	g.AddEdge(fn11, fn1)
	g.AddEdge(fn12, fn1)
	g.AddEdge(fn2, target)
	g.AddEdge(fn22, fn2)
	g.AddEdge(fn21, fn2)

	visits := make([]visitEntry, 0)

	// act
	g.WalkCallers(target, func(curr *model.Symbol, level int) {
		visits = append(visits, visitEntry{curr.Name, level})
	})

	// assert
	assert.Equal(t, []visitEntry{
		{fn1.Name, 1},
		{fn11.Name, 2},
		{fn12.Name, 2},
		{fn2.Name, 1},
		{fn21.Name, 2},
		{fn22.Name, 2},
	}, visits)
}

func TestWalkCallersDiamond(t *testing.T) {
	t.Parallel()

	// arrange
	left := fn1
	right := fn2
	bottom := fn3
	/*
			  target
			 /      \
		left      right
			 \      /
				bottom
	*/
	g := graph.New()
	g.AddEdge(left, target)
	g.AddEdge(right, target)
	g.AddEdge(bottom, right)
	g.AddEdge(bottom, left)

	visits := make([]visitEntry, 0)

	// act
	g.WalkCallers(target, func(curr *model.Symbol, level int) {
		visits = append(visits, visitEntry{curr.Name, level})
	})

	// assert
	assert.Equal(t, []visitEntry{
		{left.Name, 1},
		{bottom.Name, 2},
		{right.Name, 1},
	}, visits)
}

func TestWalkCallersCycleIncludingTarget(t *testing.T) {
	t.Parallel()

	// arrange
	// target <- fn1 <- fn2 <- target
	g := graph.New()
	g.AddEdge(fn1, target)
	g.AddEdge(fn2, fn1)
	g.AddEdge(target, fn2)

	visits := make([]visitEntry, 0)

	// act
	g.WalkCallers(target, func(curr *model.Symbol, level int) {
		visits = append(visits, visitEntry{curr.Name, level})
	})

	// assert
	assert.Equal(t, []visitEntry{
		{fn1.Name, 1},
		{fn2.Name, 2},
	}, visits)
}

func TestWalkCallersCycleSelfCall(t *testing.T) {
	t.Parallel()

	// arrange
	g := graph.New()
	g.AddEdge(fn1, target)
	g.AddEdge(fn2, fn1)
	g.AddEdge(fn1, fn1)
	g.AddEdge(fn2, fn2)
	g.AddEdge(target, target)

	visits := make([]visitEntry, 0)

	// act
	g.WalkCallers(target, func(curr *model.Symbol, level int) {
		visits = append(visits, visitEntry{curr.Name, level})
	})

	// assert
	assert.Equal(t, []visitEntry{
		{fn1.Name, 1},
		{fn2.Name, 2},
	}, visits)
}

func TestWalkCallersCycle2Callers(t *testing.T) {
	t.Parallel()

	// arrange
	g := graph.New()
	g.AddEdge(fn1, target)
	g.AddEdge(fn2, fn1)
	g.AddEdge(fn1, fn2)

	visits := make([]visitEntry, 0)

	// act
	g.WalkCallers(target, func(curr *model.Symbol, level int) {
		visits = append(visits, visitEntry{curr.Name, level})
	})

	// assert
	assert.Equal(t, []visitEntry{
		{fn1.Name, 1},
		{fn2.Name, 2},
	}, visits)
}

func TestWalkCallersCycleAll(t *testing.T) {
	t.Parallel()

	// arrange
	symbols := []*model.Symbol{target, fn1, fn2, fn3}

	g := graph.New()
	g.AddEdge(fn1, target)
	g.AddEdge(fn2, fn1)
	g.AddEdge(fn3, fn2)

	for _, caller := range symbols {
		for _, callee := range symbols {
			g.AddEdge(caller, callee)
		}
	}

	visits := make([]visitEntry, 0)

	// act
	g.WalkCallers(target, func(curr *model.Symbol, level int) {
		visits = append(visits, visitEntry{curr.Name, level})
	})

	// assert
	assert.Equal(t, []visitEntry{
		{fn1.Name, 1},
		{fn2.Name, 2},
		{fn3.Name, 3},
	}, visits)
}

func symbol(name string) *model.Symbol {
	return &model.Symbol{
		ID:   name,
		Name: name,
	}
}
