// Package test_files provides details about test files
package test_files

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
	Name string
	Line uint
	Char uint
	Refs []FuncRef
}

// FuncRef is a reference to the current function.
type FuncRef struct {
	File string
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
	name                 string
	NewService, Foo, Bar FuncInfo
}

func (f serviceFile) Name() string {
	return f.name
}

func (f serviceFile) ListFunctions() []*FuncInfo {
	return []*FuncInfo{
		&f.NewService,
		&f.Foo,
		&f.Bar,
	}
}

func (f dbFile) Name() string {
	return f.name
}

type dbFile struct {
	name  string
	Query FuncInfo
}

func (f dbFile) ListFunctions() []*FuncInfo {
	return []*FuncInfo{
		&f.Query,
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
