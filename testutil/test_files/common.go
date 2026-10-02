// Package test_files provides details about test files
package test_files

import (
	"cmp"
	"slices"
)

// TestDirs are test dir names in test_files.
var TestDirs = struct {
	Go   string
	Rust string
}{
	Go:   "go",
	Rust: "rust",
}

// FuncInfo provides name, location, and call info about a function.
type FuncInfo struct {
	Name  string
	Line  uint
	Char  uint
	Refs  []FuncRef
	Calls []FuncCall
}

func newFunc(name string, line uint, char uint) FuncInfo {
	return FuncInfo{name, line, char, []FuncRef{}, []FuncCall{}}
}

func extFunc(name string) *FuncInfo {
	f := newFunc(name, 0, 0)

	return &f
}

func call(file TestFile, caller *FuncInfo, callee *FuncInfo, line uint, char uint) {
	c := FuncCall{callee.Name, line, char}
	if !slices.Contains(caller.Calls, c) {
		caller.Calls = append(caller.Calls, c)
	}

	callee.Refs = append(callee.Refs, FuncRef{file.Name(), line, char})
}

// FuncRef is a reference to the current function.
type FuncRef struct {
	File string
	Line uint
	Char uint
}

// FuncCall is a call to another function.
type FuncCall struct {
	Name string
	Line uint
	Char uint
}

// TestFileList provides the directory and files of a test data project.
type TestFileList interface {
	TestDir() string
	TestFiles() []TestFile
}

// TestFile is a file's name and its functions declarations/definitions.
type TestFile interface {
	Name() string
	ListFunctions() []*FuncInfo
}

type mainFile struct {
	name             string
	Main, Start, End FuncInfo
}

func (f mainFile) Name() string {
	return f.name
}

func (f mainFile) ListFunctions() []*FuncInfo {
	return []*FuncInfo{
		&f.Main,
		&f.Start,
		&f.End,
	}
}

type serviceFile struct {
	name                                   string
	RunService, NewService, Foo, Bar, Save FuncInfo
}

func (f serviceFile) Name() string {
	return f.name
}

func (f serviceFile) ListFunctions() []*FuncInfo {
	return []*FuncInfo{
		&f.RunService,
		&f.NewService,
		&f.Foo,
		&f.Bar,
		&f.Save,
	}
}

func (f dbFile) Name() string {
	return f.name
}

type dbFile struct {
	name                     string
	QueryRO, QueryRW, Update FuncInfo
}

func (f dbFile) ListFunctions() []*FuncInfo {
	return []*FuncInfo{
		&f.QueryRO,
		&f.QueryRW,
		&f.Update,
	}
}

type externalFile struct {
	name               string
	ExtHello, ExtPrint FuncInfo
}

func (f externalFile) Name() string {
	return f.name
}

func (f externalFile) ListFunctions() []*FuncInfo {
	return []*FuncInfo{
		&f.ExtHello,
		&f.ExtPrint,
	}
}

type recurseFile struct {
	name                                       string
	RecurseSelf, RecurseOther, RecurseOrReturn FuncInfo
}

func (f recurseFile) Name() string {
	return f.name
}

func (f recurseFile) ListFunctions() []*FuncInfo {
	return []*FuncInfo{
		&f.RecurseSelf,
		&f.RecurseOther,
		&f.RecurseOrReturn,
	}
}

type nestedFile struct {
	name                                                 string
	Nested0, Nested1, Nested2, Nested3, Nested4, Nested5 FuncInfo
}

func (f nestedFile) Name() string {
	return f.name
}

func (f nestedFile) ListFunctions() []*FuncInfo {
	return []*FuncInfo{
		&f.Nested0,
		&f.Nested1,
		&f.Nested2,
		&f.Nested3,
		&f.Nested4,
		&f.Nested5,
	}
}

type chainFile struct {
	name                          string
	Chain, Chain1, Chain2, Chain3 FuncInfo
}

func (f chainFile) Name() string {
	return f.name
}

func (f chainFile) ListFunctions() []*FuncInfo {
	return []*FuncInfo{
		&f.Chain,
		&f.Chain1,
		&f.Chain2,
		&f.Chain3,
	}
}

func sortData(data TestFileList) {
	for _, file := range data.TestFiles() {
		for _, fn := range file.ListFunctions() {
			SortFuncCalls(fn.Calls)
		}
	}
}

// SortFuncCalls sorts calls by name, line, & char.
func SortFuncCalls(calls []FuncCall) {
	slices.SortFunc(calls, func(l, r FuncCall) int {
		c := cmp.Compare(l.Name, r.Name)
		if c != 0 {
			return c
		}

		c = cmp.Compare(l.Line, r.Line)
		if c != 0 {
			return c
		}

		return cmp.Compare(l.Char, r.Char)
	})
}
