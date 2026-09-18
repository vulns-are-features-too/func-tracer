package fakes

import (
	"context"
	"math/rand"
	"slices"

	"github.com/vulns-are-features-too/func-tracer/model"
)

// FakeSourceProvider fakes the index & LSP.
type FakeSourceProvider struct {
	funcs []*funcDef
}

type funcDef struct {
	Sym     *model.Symbol
	callees []*funcRef
	callers []*funcDef
}

type funcRef struct {
	callee *model.Symbol
	pos    model.Position
}

// Source create FakeSourceProvider.
func Source() *FakeSourceProvider {
	return &FakeSourceProvider{make([]*funcDef, 0)}
}

// WithCapacity preallocates the underlying slice.
func (src *FakeSourceProvider) WithCapacity(c int) *FakeSourceProvider {
	src.funcs = slices.Grow(src.funcs, c)

	return src
}

// Func registers a new func or gets it from known funcs.
//
//nolint:revive // unexported-return
func (src *FakeSourceProvider) Func(sym *model.Symbol) *funcDef {
	for _, fn := range src.funcs {
		if fn.Sym.ID == sym.ID {
			return fn
		}
	}

	fn := &funcDef{
		sym,
		make([]*funcRef, 0),
		make([]*funcDef, 0),
	}
	src.funcs = append(src.funcs, fn)

	return fn
}

// FindFunction finds the func definition based on URI & Start.
func (src *FakeSourceProvider) FindFunction(location model.Location) (model.Symbol, bool) {
	for _, fn := range src.funcs {
		if fn.Sym.Location.Range.Start == location.Range.Start &&
			fn.Sym.Location.URI == location.URI {
			return *fn.Sym, true
		}
	}

	return model.Symbol{}, false
}

// FindFunctionByName finds the func definition based on URI & name.
func (src *FakeSourceProvider) FindFunctionByName(uri string, name string) []model.Symbol {
	results := make([]model.Symbol, 0)

	for _, fn := range src.funcs {
		sym := fn.Sym
		if sym.Location.URI == uri && sym.Name == name {
			results = append(results, *sym)
		}
	}

	return results
}

// FindFunctionCalls finds calls in the provided function definition.
func (src *FakeSourceProvider) FindFunctionCalls(location model.Location) []model.Symbol {
	for _, fn := range src.funcs {
		if fn.Sym.Location.Range.Start == location.Range.Start {
			results := make([]model.Symbol, 0, len(fn.callees))
			for _, call := range fn.callees {
				results = append(results, call.symbol())
			}

			return results
		}
	}

	return []model.Symbol{}
}

// References finds references to the provided function.
func (src *FakeSourceProvider) References(
	_ context.Context,
	location model.Location,
) ([]model.Location, error) {
	var target *funcDef

	for _, fn := range src.funcs {
		if fn.Sym.Location.Range.Start == location.Range.Start {
			target = fn

			break
		}
	}

	results := make([]model.Location, 0, len(target.callers))
	for _, caller := range target.callers {
		results = append(results, caller.Sym.Location)
	}

	return results, nil
}

// Definitions finds definitions of the provided function.
func (src *FakeSourceProvider) Definitions(
	_ context.Context,
	callee model.Location,
) ([]model.Location, error) {
	results := make([]model.Location, 0)

	for _, fn := range src.funcs {
		for _, call := range fn.callees {
			if call.pos == callee.Range.Start {
				results = append(results, call.callee.Location)
			}
		}
	}

	return results, nil
}

// Close does nothing.
func (src *FakeSourceProvider) Close(context.Context) {}

// AddRandomCalls randomly creates calls between registered functions.
func (src *FakeSourceProvider) AddRandomCalls(calls uint) {
	length := len(src.funcs)

	//nolint:gosec // G404
	for range calls {
		r1 := uint(rand.Intn(length))
		r2 := uint(rand.Intn(length))
		caller := src.funcs[r1]
		callee := src.funcs[r2]
		caller.callsAt(callee, r1, r2)
	}
}

// LinkFuncsWithoutCalls makes sure every function has at least 1 caller/callee
// by making the target call functions with no callers
// and be called by functions with no callee.
func (src *FakeSourceProvider) LinkFuncsWithoutCalls(target *funcDef) {
	const (
		calleeOffset = 500
		callerOffset = 800
	)

	l := len(src.funcs)
	hasCallers := make([]bool, l)
	hasCallees := make([]bool, l)

	for i, fn := range src.funcs {
		hasCallees[i] = len(fn.callees) > 0
		hasCallers[i] = len(fn.callers) > 0
	}

	for i := range l {
		if !hasCallees[i] {
			src.funcs[i].Calls(target, calleeOffset+uint(i))
		}

		if !hasCallers[i] {
			target.Calls(src.funcs[i], callerOffset+uint(i))
		}
	}
}

// Calls registers a function call to callee at fake pos.
func (fn *funcDef) Calls(callee *funcDef, pos uint) *funcDef {
	return fn.callsAt(callee, pos, pos)
}

func (fn *funcDef) callsAt(callee *funcDef, line uint, char uint) *funcDef {
	fn.callees = append(fn.callees, &funcRef{
		callee.Sym,
		model.Position{Line: line, Character: char},
	})

	if !slices.Contains(callee.callers, fn) {
		callee.callers = append(callee.callers, fn)
	}

	return fn
}

func (ref funcRef) symbol() model.Symbol {
	return model.Symbol{
		ID:       ref.callee.ID,
		Name:     ref.callee.Name,
		Location: ref.location(),
	}
}

func (ref funcRef) location() model.Location {
	return model.Location{
		URI: ref.callee.Location.URI,
		Range: model.Range{
			Start: ref.pos,
			End:   ref.pos,
		},
	}
}
