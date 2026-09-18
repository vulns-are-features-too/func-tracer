package main

type (
	ReadOnlyDB  struct{}
	ReadWriteDB struct{}
)

func (db ReadOnlyDB) query(q string) string {
	return q
}

func (db ReadWriteDB) query(q string) string {
	return q
}

func (db ReadWriteDB) update(q string) string {
	return q
}

type readonlyDB interface {
	query(q string) string
}

type writeableDB interface {
	readonlyDB
	update(q string) string
}
