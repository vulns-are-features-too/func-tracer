// Package graph provides the call graph
package graph

import (
	"cmp"
	"slices"
	"sync"

	"github.com/vulns-are-features-too/func-tracer/model"
)

type symbolMap map[string]*model.Symbol

// Graph is the call graph.
type Graph struct {
	mu      sync.RWMutex
	callers map[string]symbolMap
	callees map[string]symbolMap
}

// New graph.
func New(trackCallers bool, trackCallees bool) *Graph {
	g := &Graph{
		callers: nil,
		callees: nil,
	}

	if trackCallers {
		g.callers = make(map[string]symbolMap)
	}

	if trackCallees {
		g.callees = make(map[string]symbolMap)
	}

	return g
}

// CallersOnly creates graph that only tracks callers.
func CallersOnly() *Graph {
	return New(true, false)
}

// CalleesOnly creates graph that only tracks callees.
func CalleesOnly() *Graph {
	return New(false, true)
}

// AddEdge for caller-callee.
func (g *Graph) AddEdge(caller, callee *model.Symbol) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.callers != nil {
		if g.callers[callee.ID] == nil {
			g.callers[callee.ID] = make(symbolMap)
		}

		g.callers[callee.ID][caller.ID] = caller
	}

	if g.callees != nil {
		if g.callees[caller.ID] == nil {
			g.callees[caller.ID] = make(symbolMap)
		}

		g.callees[caller.ID][callee.ID] = callee
	}
}

// Callers returns callers of the specified function.
func (g *Graph) Callers(symbol *model.Symbol) []*model.Symbol {
	return g.listSymbols(g.callers, symbol)
}

// Callees returns functions called by the specified function.
func (g *Graph) Callees(symbol *model.Symbol) []*model.Symbol {
	return g.listSymbols(g.callees, symbol)
}

// VisitFunc is for what to do at each node when walking the call graph.
type VisitFunc func(curr *model.Symbol, level int)

// nextSymbolsFunc is for how to get the next symbols from the current symbol.
type nextSymbolsFunc func(symbol *model.Symbol) []*model.Symbol

// WalkCallers does DFS walk of
// the tree whose root is the target
// and branches are callers of said target,
// calling visit on each node.
func (g *Graph) WalkCallers(target *model.Symbol, visit VisitFunc) {
	g.walk(target, g.Callers, visit)
}

// WalkCallees does DFS walk of
// the tree whose root is the target
// and branches are callees of said target,
// calling visit on each node.
func (g *Graph) WalkCallees(target *model.Symbol, visit VisitFunc) {
	g.walk(target, g.Callees, visit)
}

func (g *Graph) listSymbols(symMap map[string]symbolMap, symbol *model.Symbol) []*model.Symbol {
	if symMap == nil {
		return make([]*model.Symbol, 0)
	}

	g.mu.RLock()
	defer g.mu.RUnlock()

	resultMap := symMap[symbol.ID]

	results := make([]*model.Symbol, 0, len(resultMap))
	for _, result := range resultMap {
		results = append(results, result)
	}

	slices.SortFunc(results, func(l, r *model.Symbol) int {
		return cmp.Compare(l.ID, r.ID)
	})

	return results
}

func (g *Graph) walk(target *model.Symbol, next nextSymbolsFunc, visit VisitFunc) {
	const level = 1

	visited := make(set)
	visited.add(target.ID)

	for _, c := range next(target) {
		g.walkInner(c, next, visit, visited, level)
	}
}

func (g *Graph) walkInner(
	curr *model.Symbol,
	next nextSymbolsFunc,
	visit VisitFunc,
	visited set,
	level int,
) {
	if _, ok := visited[curr.ID]; ok {
		return
	}

	visited.add(curr.ID)
	visit(curr, level)

	for _, c := range next(curr) {
		g.walkInner(c, next, visit, visited, level+1)
	}
}
