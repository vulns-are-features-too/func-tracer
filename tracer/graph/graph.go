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
}

// New graph.
func New() *Graph {
	return &Graph{
		callers: make(map[string]symbolMap),
	}
}

// AddEdge for caller-callee.
func (g *Graph) AddEdge(caller, callee *model.Symbol) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.callers[callee.ID] == nil {
		g.callers[callee.ID] = make(symbolMap)
	}

	g.callers[callee.ID][caller.ID] = caller
}

// Callers returns callers of the specified function.
func (g *Graph) Callers(symbol *model.Symbol) []*model.Symbol {
	g.mu.RLock()
	defer g.mu.RUnlock()

	callersMap := g.callers[symbol.ID]

	callers := make([]*model.Symbol, 0, len(callersMap))
	for _, caller := range callersMap {
		callers = append(callers, caller)
	}

	slices.SortFunc(callers, func(l, r *model.Symbol) int {
		return cmp.Compare(l.ID, r.ID)
	})

	return callers
}

// WalkFunc is for what to do at each node when walking the call graph.
type WalkFunc func(curr *model.Symbol, level int)

// WalkCallers does DFS walk of
// the tree whose root is the target
// and branches are callers of said target,
// calling fn on each node.
func (g *Graph) WalkCallers(target *model.Symbol, fn WalkFunc) {
	const level = 1

	walkedIDs := make(set)
	walkedIDs.add(target.ID)

	for _, c := range g.Callers(target) {
		g.walkCallersInner(c, fn, walkedIDs, level)
	}
}

func (g *Graph) walkCallersInner(curr *model.Symbol, fn WalkFunc, walkedIDs set, level int) {
	if _, ok := walkedIDs[curr.ID]; ok {
		return
	}

	walkedIDs.add(curr.ID)
	fn(curr, level)

	for _, c := range g.Callers(curr) {
		g.walkCallersInner(c, fn, walkedIDs, level+1)
	}
}
