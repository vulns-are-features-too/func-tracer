// Package model provides common models for func tracing
package model

import "fmt"

// Position in a file.
type Position struct {
	Line      uint `json:"line"`
	Character uint `json:"character"`
}

// Range of 2 positions.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location in a file.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// Symbol of function.
type Symbol struct {
	ID       string
	Name     string
	Location Location
}

// SymbolID for map keys & hasing.
func SymbolID(loc Location, name string) string {
	return fmt.Sprintf(
		"%s:%d:%d:%s",
		loc.URI,
		loc.Range.Start.Line,
		loc.Range.Start.Character,
		name,
	)
}
