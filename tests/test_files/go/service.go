package main

type service struct{ db }

func newService() *service {
	return &service{}
}

func (svc *service) foo() {
	svc.db.query("foo")
}

func (svc *service) bar() {
	svc.db.query("bar")
}
