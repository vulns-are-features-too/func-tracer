package executor_test

import (
	"context"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vulns-are-features-too/func-tracer/tracer/executor"
)

type taskWithDepth struct {
	id    int64
	depth int
}

type workerWithDepth struct {
	next func(t taskWithDepth) []taskWithDepth
}

func (w workerWithDepth) Exec(t taskWithDepth) []taskWithDepth {
	return w.next(t)
}

func TestExecutorRace_ConcurrentEnqueue(t *testing.T) {
	t.Parallel()

	const (
		producers          = 32
		tasksPerProd       = 1_000
		workers            = 8
		wantedResult int64 = producers*tasksPerProd + 1
	)

	var processed atomic.Int64

	newWorker := workerFactory(func(int) []int {
		processed.Add(1)
		runtime.Gosched()

		return nil
	})

	ctx := context.Background()
	e := executor.New(ctx, workers, newWorker)

	e.Enqueue(1)

	var runWG sync.WaitGroup
	runWG.Go(e.Run)

	var enqueueWG sync.WaitGroup
	enqueueWG.Add(producers)

	for producer := range producers {
		go func(producer int) {
			defer enqueueWG.Done()

			for i := range tasksPerProd {
				e.Enqueue(producer*tasksPerProd + i)
				runtime.Gosched()
			}
		}(producer)
	}

	enqueueWG.Wait()
	runWG.Wait()

	assert.Equal(t, wantedResult, processed.Load())
	assert.Zero(t, e.RemainingTasks())
}

func TestExecutorRace_EnqueueWhileWorkersAreRunning(t *testing.T) {
	t.Parallel()

	const (
		workers    = 8
		initTasks  = workers
		extraTasks = 1000
	)

	var (
		startedCount atomic.Int64
		completed    atomic.Int64
	)

	started := make(chan struct{}, workers)
	release := make(chan struct{})

	newWorker := workerFactory(func(int) []int {
		if startedCount.Add(1) <= workers {
			started <- struct{}{}
		}

		<-release

		completed.Add(1)

		return nil
	})

	e := executor.New(t.Context(), workers, newWorker)
	e.Enqueue(makeTasks(initTasks)...)

	var runWG sync.WaitGroup
	runWG.Go(e.Run)

	for i := range workers {
		select {
		case <-started:
		case <-runTimeout():
			t.Fatalf("worker %d didn't start", i)
		}
	}

	var enqueueWG sync.WaitGroup

	// enqueue more work while workers are running
	enqueueWG.Go(func() {
		for i := range extraTasks {
			e.Enqueue(i)
			runtime.Gosched()
		}
	})

	enqueueWG.Wait()

	close(release)
	runWG.Wait()

	assert.Equal(t, int64(initTasks+extraTasks), completed.Load())
	assert.Zero(t, e.RemainingTasks())
}

func TestExecutorRace_CancelWhileWorkersExecuting(t *testing.T) {
	t.Parallel()

	const workers = 16

	var (
		startedCount atomic.Int64
		completed    atomic.Int64
	)

	started := make(chan struct{}, workers)
	release := make(chan struct{})
	tasks := makeTasks(workers)
	doneTasks := resultCollector()

	newWorker := workerFactory(func(task int) []int {
		startedCount.Add(1)

		started <- struct{}{}

		<-release

		doneTasks.add(task)
		completed.Add(1)

		// These must not be processed after cancellation.
		return []int{task + 1, task + 2}
	})

	ctx, cancel := context.WithCancel(context.Background())
	e := executor.New(ctx, workers, newWorker)
	e.Enqueue(tasks...)

	var runWG sync.WaitGroup
	runWG.Go(e.Run)

	for i := range workers {
		select {
		case <-started:
		case <-runTimeout():
			t.Fatalf("worker %d did not start", i)
		}
	}

	cancel()
	close(release)

	runWG.Wait()

	assert.Equal(t, int64(workers), startedCount.Load())
	assert.Equal(t, int64(workers), completed.Load())
	assert.Zero(t, e.RemainingTasks())

	results := doneTasks.dump()
	slices.Sort(results)
	assert.Equal(t, tasks, results)
}

func TestExecutorRace_ExecGeneratesTasks(t *testing.T) {
	t.Parallel()

	const (
		workers         = 8
		initTasks       = 100
		depth           = 5
		newTasksPerTask = 3
	)

	var (
		currTasks  int64 = initTasks
		totalTasks       = currTasks
	)
	for range depth {
		currTasks *= newTasksPerTask
		totalTasks += currTasks
	}

	var processed atomic.Int64

	next := func(t taskWithDepth) []taskWithDepth {
		nextTasks := make([]taskWithDepth, 0, newTasksPerTask)

		processed.Add(1)

		if t.id%3 != 0 {
			runtime.Gosched()
		}

		if t.depth >= depth {
			return nil
		}

		for i := range newTasksPerTask {
			nextTasks = append(nextTasks, taskWithDepth{
				t.id + int64(i) + 1,
				t.depth + 1,
			})
		}

		return nextTasks
	}
	newWorker := func() workerWithDepth {
		return workerWithDepth{next}
	}

	ctx := context.Background()
	e := executor.New(ctx, workers, newWorker)

	for i := range initTasks {
		e.Enqueue(taskWithDepth{int64(i) * 1000, 0})
	}

	e.Run()

	assert.Equal(t, totalTasks, processed.Load())
	assert.Zero(t, e.RemainingTasks())
}
