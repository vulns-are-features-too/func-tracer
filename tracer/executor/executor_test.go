package executor_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vulns-are-features-too/func-tracer/tracer/executor"
)

type resultCollection struct {
	mu      sync.Mutex
	results []int
}

func resultCollector() *resultCollection {
	return &resultCollection{sync.Mutex{}, []int{}}
}

func (rc *resultCollection) add(i int) {
	rc.mu.Lock()
	rc.results = append(rc.results, i)
	rc.mu.Unlock()
}

func (rc *resultCollection) dump() []int {
	return rc.results
}

type workerFunc func(int) []int

type testWorker struct {
	exec workerFunc
}

func (w testWorker) Exec(task int) []int {
	return w.exec(task)
}

func workerFactory(exec workerFunc) func() testWorker {
	return func() testWorker {
		return testWorker{exec}
	}
}

func TestJustInitialTasks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tasks := makeTasks(5)
	results := resultCollector()

	newWorker := workerFactory(func(task int) []int {
		results.add(task)

		return nil
	})

	e := executor.New(ctx, 2, newWorker)

	e.Enqueue(tasks...)
	e.Run()

	assert.ElementsMatch(t, tasks, results.dump())
	assert.Zero(t, e.RemainingTasks())
}

func TestWorkersReturnNewTasks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	want := []int{0, 1, 2, 3}
	results := resultCollector()

	newWorker := workerFactory(func(task int) []int {
		results.add(task)

		if task < 3 {
			return []int{task + 1}
		}

		return nil
	})

	e := executor.New(ctx, 2, newWorker)

	e.Enqueue(0)
	e.Run()

	assert.Equal(t, want, results.dump())
}

func TestDoesNotStartWorkersBeforeRun(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64

	newWorker := workerFactory(func(int) []int {
		calls.Add(1)

		return nil
	})

	e := executor.New(context.Background(), 2, newWorker)

	e.Enqueue(1, 2, 3)

	assert.Zero(t, calls.Load())
	assert.Equal(t, 3, e.RemainingTasks())

	e.Run()

	assert.Equal(t, int64(3), calls.Load())
}

func TestEmptyExecutorCompletes(t *testing.T) {
	t.Parallel()

	newWorker := workerFactory(func(int) []int {
		panic("shouldn't run")
	})

	e := executor.New(context.Background(), 2, newWorker)

	e.Run()
}

func TestRunIsSingleUse(t *testing.T) {
	t.Parallel()

	newWorker := workerFactory(func(int) []int {
		return []int{}
	})
	e := executor.New(context.Background(), 1, newWorker)

	e.Enqueue(1)
	e.Run()

	assert.Panics(t, e.Run)
}

func TestEnqueueAfterDoneIsIgnored(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64

	newWorker := workerFactory(func(int) []int {
		calls.Add(1)

		return nil
	})

	e := executor.New(context.Background(), 1, newWorker)

	e.Enqueue(1)
	e.Run()

	assert.Equal(t, int64(1), calls.Load())

	e.Enqueue(2, 3, 4)

	assert.Equal(t, int64(1), calls.Load())
	assert.Zero(t, e.RemainingTasks())
}

func TestMaxWorkersNotExceeded(t *testing.T) {
	t.Parallel()

	const (
		maxWorkers = 3
		tasks      = 6
	)

	active := atomic.Int64{}
	maxActive := atomic.Int64{}
	started := make(chan struct{}, tasks+1)
	release := make(chan struct{})

	newWorker := workerFactory(func(int) []int {
		n := active.Add(1)
		defer active.Add(-1)

		for {
			old := maxActive.Load()
			if n <= old || maxActive.CompareAndSwap(old, n) {
				break
			}
		}

		started <- struct{}{}

		<-release

		return nil
	})

	e := executor.New(context.Background(), maxWorkers, newWorker)
	e.Enqueue(makeTasks(tasks)...)

	done := make(chan struct{})

	go func() {
		e.Run()
		close(done)
	}()

	for range maxWorkers {
		select {
		case <-started:
		case <-runTimeout():
			t.Fatal("timed out waiting for workers to start")
		}
	}

	assert.LessOrEqual(t, maxActive.Load(), int64(maxWorkers))

	close(release)

	select {
	case <-done:
	case <-runTimeout():
		t.Fatal("Run did not finish")
	}

	assert.Zero(t, active.Load())
	assert.LessOrEqual(t, maxActive.Load(), int64(maxWorkers))
}

func TestCancelWaitsForExecutingTasks(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. e.Run()
	// 2. make worker wait on started
	// 3. cancel (after start, before release) => worker keeps going
	// 4. release worker to continue
	// 5. worker finished => e.Run() done
	started := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan struct{})

	newWorker := workerFactory(func(task int) []int {
		require.Equal(t, 1, task)
		close(started)
		<-release
		close(finished)

		return nil
	})

	e := executor.New(ctx, 1, newWorker)

	e.Enqueue(1)

	done := make(chan struct{})

	go func() {
		e.Run()
		close(done)
	}()

	<-started
	cancel()

	select {
	case <-done:
		t.Fatal("Run returned before executing task finished")
	case <-time.After(10 * time.Millisecond):
	}

	close(release)

	select {
	case <-finished:
	case <-runTimeout():
		t.Fatal("worker did not finish")
	}

	select {
	case <-done:
	case <-runTimeout():
		t.Fatal("Run did not finish after worker completed")
	}
}

func TestCancelDiscardsExecutordTasks(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tasks := makeTasks(3)
	wantedResult := tasks[:1]
	results := resultCollector()

	started := make(chan struct{})
	release := make(chan struct{})

	newWorker := workerFactory(func(task int) []int {
		results.add(task)
		close(started)
		<-release

		return nil
	})

	e := executor.New(ctx, 1, newWorker)

	e.Enqueue(tasks...)

	done := make(chan struct{})

	go func() {
		e.Run()
		close(done)
	}()

	// 1. start a single worker/task
	<-started
	// 2. cancel queue
	cancel()
	// 3. let previously started task finish
	close(release)
	// expectation: only that 1 task should have finished

	select {
	case <-done:
	case <-runTimeout():
		t.Fatal("Run did not finish")
	}

	assert.Equal(t, wantedResult, results.dump())
	assert.Zero(t, e.RemainingTasks())
}

func TestResultsReturnedAfterCancellationAreDiscarded(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var calls atomic.Int64

	started := make(chan struct{})
	release := make(chan struct{})

	newWorker := workerFactory(func(task int) []int {
		calls.Add(1)
		require.Equal(t, 1, task)

		close(started)
		<-release

		return []int{2, 3}
	})

	e := executor.New(ctx, 1, newWorker)

	e.Enqueue(1)

	done := make(chan struct{})

	go func() {
		e.Run()
		close(done)
	}()

	<-started
	cancel()
	close(release)

	select {
	case <-done:
	case <-runTimeout():
		t.Fatal("Run did not finish")
	}

	assert.Equal(t, int64(1), calls.Load())
	assert.Zero(t, e.RemainingTasks())
}

func makeTasks(n int) []int {
	tasks := make([]int, n)
	for i := range n {
		// ensure value != index
		tasks[i] = i * 10
	}

	return tasks
}

func runTimeout() <-chan time.Time {
	return time.After(10 * time.Millisecond)
}
