package tracer

import (
	"context"

	"github.com/vulns-are-features-too/func-tracer/logging"
)

type worker struct {
	ctx           context.Context
	logger        logging.Logger
	visited       *set
	maxDepth      int
	trace         traceFunc
	collectResult collectResultFunc
}

func (w worker) Exec(t task) []task {
	if w.maxDepth != 0 && t.depth > w.maxDepth {
		w.logger.Infof("max depth reached")

		return nil
	}

	results, err := w.trace(w.ctx, t.symbol)
	if err != nil {
		w.logger.Errorf("%w", err)

		return nil
	}

	nextTasks := make([]task, 0, len(results))

	for _, res := range results {
		w.collectResult(t.symbol, res)

		if w.visited.has(res.ID) {
			continue
		}

		w.visited.add(res.ID)
		nextTasks = append(nextTasks, t.next(res))
	}

	w.logger.Debugf("got %d new tasks from %s", len(nextTasks), t.symbol.ID)

	return nextTasks
}
