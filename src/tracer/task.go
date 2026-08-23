package tracer

import (
	"github.com/vulns-are-features-too/func-tracer/src/model"
)

type task struct {
	symbol *model.Symbol
	depth  int
}

func newTask(symbol *model.Symbol) task {
	return task{symbol: symbol, depth: 1}
}

func (t *task) next(symbol *model.Symbol) task {
	return task{symbol: symbol, depth: t.depth + 1}
}

type result struct {
	task    task
	callers []*model.Symbol
	err     error
}
