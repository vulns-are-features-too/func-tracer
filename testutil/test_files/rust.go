package test_files

// RustFiles contains all files and their functions
// in the test Rust project.
var RustFiles = initRustFiles()

//nolint:funlen,mnd
func initRustFiles() *rustFiles {
	f := &rustFiles{
		Main: mainFile{
			name:  "main.rs",
			Main:  newFunc("main", 7, 3),
			Start: newFunc("start", 25, 3),
			End:   newFunc("end", 27, 3),
		},
		Service: serviceFile{
			name:       "service.rs",
			RunService: newFunc("run", 7, 7),
			NewService: newFunc("new", 15, 11),
			Foo:        newFunc("foo", 22, 11),
			Bar:        newFunc("bar", 26, 11),
			Save:       newFunc("save", 31, 11),
		},
		Db: dbFile{
			name:    "db.rs",
			QueryRO: newFunc("query", 4, 7),
			QueryRW: newFunc("query", 10, 7),
			Update:  newFunc("update", 16, 7),
		},
		External: externalFile{
			name:     "external.rs",
			ExtHello: newFunc("ext_hello", 2, 7),
			ExtPrint: newFunc("ext_print", 6, 7),
		},
		Recurse: recurseFile{
			name:            "recurse.rs",
			RecurseSelf:     newFunc("recurse_self", 0, 7),
			RecurseOther:    newFunc("recurse_other", 6, 7),
			RecurseOrReturn: newFunc("recurse_or_return", 10, 7),
		},
		Nested: nestedFile{
			name:    "nested.rs",
			Nested0: newFunc("nested0", 0, 7),
			Nested1: newFunc("nested1", 4, 3),
			Nested2: newFunc("nested2", 8, 3),
			Nested3: newFunc("nested3", 12, 3),
			Nested4: newFunc("nested4", 16, 3),
			Nested5: newFunc("nested5", 20, 3),
		},
		Chain: chainFile{
			name:   "chain.rs",
			Chain:  newFunc("chain", 0, 7),
			Chain1: newFunc("chain1", 4, 3),
			Chain2: newFunc("chain2", 8, 3),
			Chain3: newFunc("chain3", 12, 3),
		},
	}

	call(f.Main, &f.Main.Main, &f.Main.Start, 8, 4)
	call(f.Main, &f.Main.Main, &f.Main.End, 22, 4)
	call(f.Main, &f.Main.Main, &f.Service.RunService, 10, 13)
	call(f.Main, &f.Main.Main, &f.External.ExtHello, 15, 14)
	call(f.Main, &f.Main.Main, &f.External.ExtPrint, 16, 14)
	call(f.Main, &f.Main.Main, &f.Recurse.RecurseSelf, 12, 13)
	call(f.Main, &f.Main.Main, &f.Recurse.RecurseOther, 13, 13)
	call(f.Main, &f.Main.Main, &f.Nested.Nested0, 18, 12)
	call(f.Main, &f.Main.Main, &f.Chain.Chain, 20, 15)
	call(f.Main, &f.Main.Main, &f.Chain.Chain, 20, 37)

	call(f.Service, &f.Service.RunService, &f.Service.NewService, 8, 27)
	call(f.Service, &f.Service.RunService, &f.Service.Foo, 9, 8)
	call(f.Service, &f.Service.RunService, &f.Service.Bar, 10, 8)
	call(f.Service, &f.Service.RunService, &f.Service.Save, 11, 8)
	call(f.Service, &f.Service.Foo, &f.Db.QueryRO, 23, 19)
	call(f.Service, &f.Service.Bar, &f.Db.QueryRO, 27, 19)
	call(f.Service, &f.Service.Bar, &f.Db.QueryRW, 28, 19)
	call(f.Service, &f.Service.Save, &f.Db.Update, 32, 19)

	call(f.Recurse, &f.Recurse.RecurseSelf, &f.Recurse.RecurseSelf, 2, 8)
	call(f.Recurse, &f.Recurse.RecurseOther, &f.Recurse.RecurseOrReturn, 7, 4)
	call(f.Recurse, &f.Recurse.RecurseOrReturn, &f.Recurse.RecurseOther, 14, 4)

	call(f.Nested, &f.Nested.Nested0, &f.Nested.Nested1, 1, 4)
	call(f.Nested, &f.Nested.Nested1, &f.Nested.Nested2, 5, 4)
	call(f.Nested, &f.Nested.Nested2, &f.Nested.Nested3, 9, 4)
	call(f.Nested, &f.Nested.Nested3, &f.Nested.Nested4, 13, 4)
	call(f.Nested, &f.Nested.Nested4, &f.Nested.Nested5, 17, 4)

	call(f.Chain, &f.Chain.Chain, &f.Chain.Chain1, 1, 4)
	call(f.Chain, &f.Chain.Chain, &f.Chain.Chain2, 1, 16)
	call(f.Chain, &f.Chain.Chain, &f.Chain.Chain3, 1, 28)

	fnStdout := extFunc("stdout")
	fnWrite := extFunc("write")
	fnAsBytes := extFunc("as_bytes")

	call(f.External, &f.External.ExtHello, fnStdout, 3, 17)
	call(f.External, &f.External.ExtHello, fnWrite, 3, 26)
	call(f.External, &f.External.ExtPrint, fnStdout, 7, 17)
	call(f.External, &f.External.ExtPrint, fnWrite, 7, 26)
	call(f.External, &f.External.ExtPrint, fnAsBytes, 7, 34)

	sortData(f)

	return f
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
