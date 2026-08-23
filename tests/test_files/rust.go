package test_files

// RustFiles contains all files and their functions
// in the test Rust project
//
//nolint:goconst
var RustFiles = &rustFiles{
	Main: mainFile{
		name:  "main.rs",
		Main:  FuncInfo{"main", 9, 3, []FuncRef{}},
		Start: FuncInfo{"start", 29, 3, []FuncRef{{"main.rs", 10, 4}}},
		End:   FuncInfo{"end", 31, 3, []FuncRef{{"main.rs", 26, 4}}},
	},
	Service: serviceFile{
		name:       "service.rs",
		NewService: FuncInfo{"new", 8, 11, []FuncRef{{"main.rs", 12, 27}}},
		Foo:        FuncInfo{"foo", 12, 11, []FuncRef{{"main.rs", 13, 8}}},
		Bar:        FuncInfo{"bar", 16, 11, []FuncRef{{"main.rs", 14, 8}}},
	},
	Db: dbFile{
		name: "db.rs",
		Query: FuncInfo{"query", 3, 11, []FuncRef{
			{"service.rs", 13, 16},
			{"service.rs", 18, 16},
		}},
	},
	External: externalFile{
		name:     "external.rs",
		ExtHello: FuncInfo{"ext_hello", 0, 7, []FuncRef{{"main.rs", 19, 14}}},
		ExtPrint: FuncInfo{"ext_print", 4, 7, []FuncRef{{"main.rs", 20, 14}}},
	},
	Recurse: recurseFile{
		name: "recurse.rs",
		RecurseSelf: FuncInfo{"recurse_self", 0, 7, []FuncRef{
			{"main.rs", 16, 13},
			{"recurse.rs", 2, 8},
		}},
		RecurseOther: FuncInfo{"recurse_other", 6, 7, []FuncRef{
			{"main.rs", 17, 13},
			{"recurse.rs", 14, 4},
		}},
		RecurseOrReturn: FuncInfo{"recurse_or_return", 10, 7, []FuncRef{
			{"recurse.rs", 7, 4},
		}},
	},
	Nested: nestedFile{
		name:    "nested.rs",
		Nested0: FuncInfo{"nested0", 0, 7, []FuncRef{{"main.rs", 22, 12}}},
		Nested1: FuncInfo{"nested1", 4, 3, []FuncRef{{"nested.rs", 1, 4}}},
		Nested2: FuncInfo{"nested2", 8, 3, []FuncRef{{"nested.rs", 5, 4}}},
		Nested3: FuncInfo{"nested3", 12, 3, []FuncRef{{"nested.rs", 9, 4}}},
		Nested4: FuncInfo{"nested4", 16, 3, []FuncRef{{"nested.rs", 13, 4}}},
		Nested5: FuncInfo{"nested5", 20, 3, []FuncRef{{"nested.rs", 17, 4}}},
	},
	Chain: chainFile{
		name: "chain.rs",
		Chain: FuncInfo{"chain", 0, 7, []FuncRef{
			{"main.rs", 24, 15},
			{"main.rs", 24, 37},
		}},
		Chain1: FuncInfo{"chain1", 4, 3, []FuncRef{{"chain.rs", 1, 4}}},
		Chain2: FuncInfo{"chain2", 8, 3, []FuncRef{{"chain.rs", 1, 16}}},
		Chain3: FuncInfo{"chain3", 12, 3, []FuncRef{{"chain.rs", 1, 28}}},
	},
}

type rustFiles struct {
	Main     mainFile
	Service  serviceFile
	Db       dbFile
	External externalFile
	Recurse  recurseFile
	Nested   nestedFile
	Chain    chainFile
}

func (*rustFiles) TestDir() string {
	return TestDirs.Rust
}

func (l *rustFiles) TestFiles() []TestFile {
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
