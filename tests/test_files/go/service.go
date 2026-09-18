package main

type service struct{ db readonlyDB }

func runService(write bool) {
	svc := newService(write)
	svc.foo()
	svc.bar()
	svc.save()
}

func newService(write bool) *service {
	if write {
		return &service{db: ReadWriteDB{}}
	}
	return &service{db: ReadOnlyDB{}}
}

func (svc *service) foo() {
	svc.db.query("foo")
}

func (svc *service) bar() {
	svc.db.query("bar")
}

func (svc *service) save() {
	if w, ok := svc.db.(writeableDB); ok {
		w.update("save")
	}
}
