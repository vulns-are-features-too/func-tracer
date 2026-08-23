package main

type db struct{}

func (db db) query(q string) string {
	return q
}
