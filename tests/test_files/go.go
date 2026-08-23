package test_files

// GoFiles contains all files and their functions
// in the test Go project
//
//nolint:goconst
var GoFiles = &goFiles{
	Main: mainFile{
		name:  "main.go",
		Main:  FuncInfo{"main", 2, 5, []FuncRef{}},
		Start: FuncInfo{"start", 22, 5, []FuncRef{{"main.go", 3, 1}}},
		End:   FuncInfo{"end", 24, 5, []FuncRef{{"main.go", 19, 1}}},
	},
	Service: serviceFile{
		name:       "service.go",
		NewService: FuncInfo{"newService", 4, 5, []FuncRef{{"main.go", 5, 8}}},
		Foo:        FuncInfo{"foo", 8, 20, []FuncRef{{"main.go", 6, 5}}},
		Bar:        FuncInfo{"bar", 12, 20, []FuncRef{{"main.go", 7, 5}}},
	},
	Db: dbFile{
		name: "db.go",
		Query: FuncInfo{"query", 4, 13, []FuncRef{
			{"service.go", 9, 8},
			{"service.go", 13, 8},
		}},
	},
	External: externalFile{
		name:     "external.go",
		ExtHello: FuncInfo{"extHello", 4, 5, []FuncRef{{"main.go", 12, 1}}},
		ExtPrint: FuncInfo{"extPrint", 8, 5, []FuncRef{{"main.go", 13, 1}}},
	},
	Recurse: recurseFile{
		name: "recurse.go",
		RecurseSelf: FuncInfo{"recurseSelf", 2, 5, []FuncRef{
			{"main.go", 9, 1},
			{"recurse.go", 4, 2},
		}},
		RecurseOther: FuncInfo{"recurseOther", 8, 5, []FuncRef{
			{"main.go", 10, 1},
			{"recurse.go", 16, 1},
		}},
		RecurseOrReturn: FuncInfo{"recurseOrReturn", 12, 5, []FuncRef{
			{"recurse.go", 9, 1},
		}},
	},
	Nested: nestedFile{
		name:    "nested.go",
		Nested0: FuncInfo{"nested0", 2, 5, []FuncRef{{"main.go", 15, 1}}},
		Nested1: FuncInfo{"nested1", 6, 5, []FuncRef{{"nested.go", 3, 1}}},
		Nested2: FuncInfo{"nested2", 10, 5, []FuncRef{{"nested.go", 7, 1}}},
		Nested3: FuncInfo{"nested3", 14, 5, []FuncRef{{"nested.go", 11, 1}}},
		Nested4: FuncInfo{"nested4", 18, 5, []FuncRef{{"nested.go", 15, 1}}},
		Nested5: FuncInfo{"nested5", 22, 5, []FuncRef{{"nested.go", 19, 1}}},
	},
	Chain: chainFile{
		name: "chain.go",
		Chain: FuncInfo{"chain", 2, 5, []FuncRef{
			{"main.go", 17, 5},
			{"main.go", 17, 20},
		}},
		Chain1: FuncInfo{"chain1", 6, 5, []FuncRef{{"chain.go", 3, 8}}},
		Chain2: FuncInfo{"chain2", 10, 5, []FuncRef{{"chain.go", 3, 20}}},
		Chain3: FuncInfo{"chain3", 14, 5, []FuncRef{{"chain.go", 3, 32}}},
	},
}

type goFiles struct {
	Main     mainFile
	Service  serviceFile
	Db       dbFile
	External externalFile
	Recurse  recurseFile
	Nested   nestedFile
	Chain    chainFile
}

func (*goFiles) TestDir() string {
	return TestDirs.Go
}

func (l *goFiles) TestFiles() []TestFile {
	return []TestFile{
		l.Main,
		l.Service,
		l.Db,
		l.External,
		l.Recurse,
		l.Nested,
		l.Chain,
	}
}
