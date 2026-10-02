package test_files

// GoFiles contains all files and their functions
// in the test Go project.
var GoFiles = initGoFiles()

//nolint:funlen,mnd
func initGoFiles() *goFiles {
	f := &goFiles{
		Main: mainFile{
			name:  "main.go",
			Main:  newFunc("main", 2, 5),
			Start: newFunc("start", 20, 5),
			End:   newFunc("end", 22, 5),
		},
		Service: serviceFile{
			name:       "service.go",
			RunService: newFunc("runService", 4, 5),
			NewService: newFunc("newService", 11, 5),
			Foo:        newFunc("foo", 18, 20),
			Bar:        newFunc("bar", 22, 20),
			Save:       newFunc("save", 26, 20),
		},
		Db: dbFile{
			name:    "db.go",
			QueryRO: newFunc("query", 7, 21),
			QueryRW: newFunc("query", 11, 22),
			Update:  newFunc("update", 15, 22),
		},
		External: externalFile{
			name:     "external.go",
			ExtHello: newFunc("extHello", 4, 5),
			ExtPrint: newFunc("extPrint", 8, 5),
		},
		Recurse: recurseFile{
			name:            "recurse.go",
			RecurseSelf:     newFunc("recurseSelf", 2, 5),
			RecurseOther:    newFunc("recurseOther", 8, 5),
			RecurseOrReturn: newFunc("recurseOrReturn", 12, 5),
		},
		Nested: nestedFile{
			name:    "nested.go",
			Nested0: newFunc("nested0", 2, 5),
			Nested1: newFunc("nested1", 6, 5),
			Nested2: newFunc("nested2", 10, 5),
			Nested3: newFunc("nested3", 14, 5),
			Nested4: newFunc("nested4", 18, 5),
			Nested5: newFunc("nested5", 22, 5),
		},
		Chain: chainFile{
			name:   "chain.go",
			Chain:  newFunc("chain", 2, 5),
			Chain1: newFunc("chain1", 6, 5),
			Chain2: newFunc("chain2", 10, 5),
			Chain3: newFunc("chain3", 14, 5),
		},
	}

	call(f.Main, &f.Main.Main, &f.Main.Start, 3, 1)
	call(f.Main, &f.Main.Main, &f.Main.End, 17, 1)
	call(f.Main, &f.Main.Main, &f.Service.RunService, 5, 1)
	call(f.Main, &f.Main.Main, &f.External.ExtHello, 10, 1)
	call(f.Main, &f.Main.Main, &f.External.ExtPrint, 11, 1)
	call(f.Main, &f.Main.Main, &f.Recurse.RecurseSelf, 7, 1)
	call(f.Main, &f.Main.Main, &f.Recurse.RecurseOther, 8, 1)
	call(f.Main, &f.Main.Main, &f.Nested.Nested0, 13, 1)
	call(f.Main, &f.Main.Main, &f.Chain.Chain, 15, 5)
	call(f.Main, &f.Main.Main, &f.Chain.Chain, 15, 20)

	call(f.Service, &f.Service.RunService, &f.Service.NewService, 5, 8)
	call(f.Service, &f.Service.RunService, &f.Service.Foo, 6, 5)
	call(f.Service, &f.Service.RunService, &f.Service.Bar, 7, 5)
	call(f.Service, &f.Service.RunService, &f.Service.Save, 8, 5)
	call(f.Service, &f.Service.Foo, &f.Db.QueryRO, 19, 8)
	call(f.Service, &f.Service.Bar, &f.Db.QueryRO, 23, 8)
	call(f.Service, &f.Service.Foo, &f.Db.QueryRW, 19, 8)
	call(f.Service, &f.Service.Bar, &f.Db.QueryRW, 23, 8)
	call(f.Service, &f.Service.Save, &f.Db.Update, 28, 4)

	call(f.Recurse, &f.Recurse.RecurseSelf, &f.Recurse.RecurseSelf, 4, 2)
	call(f.Recurse, &f.Recurse.RecurseOther, &f.Recurse.RecurseOrReturn, 9, 1)
	call(f.Recurse, &f.Recurse.RecurseOrReturn, &f.Recurse.RecurseOther, 16, 1)

	call(f.Nested, &f.Nested.Nested0, &f.Nested.Nested1, 3, 1)
	call(f.Nested, &f.Nested.Nested1, &f.Nested.Nested2, 7, 1)
	call(f.Nested, &f.Nested.Nested2, &f.Nested.Nested3, 11, 1)
	call(f.Nested, &f.Nested.Nested3, &f.Nested.Nested4, 15, 1)
	call(f.Nested, &f.Nested.Nested4, &f.Nested.Nested5, 19, 1)

	call(f.Chain, &f.Chain.Chain, &f.Chain.Chain1, 3, 8)
	call(f.Chain, &f.Chain.Chain, &f.Chain.Chain2, 3, 20)
	call(f.Chain, &f.Chain.Chain, &f.Chain.Chain3, 3, 32)

	call(f.External, &f.External.ExtHello, extFunc("Println"), 5, 5)
	call(f.External, &f.External.ExtPrint, extFunc("Printf"), 9, 5)

	sortData(f)

	return f
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
