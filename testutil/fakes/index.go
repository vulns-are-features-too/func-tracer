package fakes

import (
	"math/rand"
	"slices"

	"github.com/vulns-are-features-too/func-tracer/model"
)

type fakeIndex struct{ symbols []*model.Symbol }

// Index fake for testing.
//
//nolint:revive // unexported-return
func Index() *fakeIndex {
	return &fakeIndex{symbols: make([]*model.Symbol, 0)}
}

// Add symbols to fake index.
func (i *fakeIndex) Add(sym ...*model.Symbol) {
	i.symbols = append(i.symbols, sym...)
}

// WithCapacity ensures the min number of symbols
// can be stored without realloc.
func (i *fakeIndex) WithCapacity(c int) *fakeIndex {
	i.symbols = slices.Grow(i.symbols, c)

	return i
}

// Random symbol previously collected.
func (i *fakeIndex) Random() *model.Symbol {
	//nolint:gosec // G404
	sym := i.symbols[rand.Intn(len(i.symbols))]

	return sym
}

// All symbols.
func (i *fakeIndex) All() []*model.Symbol {
	return i.symbols
}

// Build does nothing.
func (i *fakeIndex) Build([]string) error { return nil }

// Close does nothing.
func (i *fakeIndex) Close() {}

// FindFunction with matching location.
func (i *fakeIndex) FindFunction(
	location model.Location,
) (model.Symbol, bool) {
	for _, sym := range i.symbols {
		if sym.Location == location {
			return *sym, true
		}
	}

	return model.Symbol{}, false
}

// FindFunctionByName with matching uri & name.
func (i *fakeIndex) FindFunctionByName(
	uri string,
	name string,
) (model.Symbol, bool) {
	for _, sym := range i.symbols {
		if sym.Name == name && sym.Location.URI == uri {
			return *sym, true
		}
	}

	return model.Symbol{}, false
}
