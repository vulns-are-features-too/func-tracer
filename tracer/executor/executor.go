// Package executor provides a task executor that runs specified tasks
package executor

import (
	"context"
	"sync"
)

type executorState int8

// Executor states.
const (
	InitState executorState = iota
	RunningState
	DoneState
)

// Worker executes tasks.
type Worker[T any] interface {
	// Exec executes a task and may return additional tasks to enqueue.
	Exec(t T) []T
}

// Executor of tasks with workers to process those tasks.
// - The number of tasks is unbounded.
// - The number of workers is bounded.
// - Each task may result in 0 or more new tasks after processing.
//
// Run is single-use.
//
// Cancellation semantics:
//   - tasks already executing are allowed to finish
//   - queued tasks are discarded
//   - tasks returned by Exec after cancellation are discarded
//   - no new tasks are started after cancellation
type Executor[T any, W Worker[T]] struct {
	ctx           context.Context
	maxWorkers    int
	newWorker     func() W
	tasks         RingBuffer[T]
	state         executorState
	mu            sync.Mutex
	activeWorkers int
	done          chan struct{}
}

// New Executor.
func New[T any, W Worker[T]](
	ctx context.Context,
	maxWorkers int,
	newWorker func() W,
) *Executor[T, W] {
	if maxWorkers <= 0 {
		maxWorkers = 1
	}

	return &Executor[T, W]{
		ctx:        ctx,
		maxWorkers: maxWorkers,
		newWorker:  newWorker,
		tasks:      NewRingBuffer[T](NextCapacity(maxWorkers)),
		done:       make(chan struct{}),
	}
}

// Run starts processing and blocks until all workers have stopped.
//
// Workers stop starting new tasks when ctx is cancelled, but any
// task already inside Exec is allowed to finish.
func (e *Executor[T, W]) Run() {
	e.mu.Lock()

	if e.state != InitState {
		e.mu.Unlock()
		panic("Run called more than once")
	}

	e.state = RunningState

	if e.ctx.Err() != nil {
		e.finishLocked()
		e.mu.Unlock()
		<-e.done

		return
	}

	e.startWorkersLocked()

	if e.activeWorkers == 0 {
		e.finishLocked()
	}

	e.mu.Unlock()

	<-e.done

	if e.state != DoneState {
		panic("Not marked as Done at end of Run")
	}

	if e.activeWorkers != 0 {
		panic("Some worker(s) still active at end of Run")
	}
}

// Enqueue tasks.
//
// Enqueue is:
// - ignored after cancellation or after Run has finished
// - safe to call concurrently.
func (e *Executor[T, W]) Enqueue(tasks ...T) {
	if len(tasks) == 0 {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state == DoneState || e.ctx.Err() != nil {
		return
	}

	if pushed := e.tasks.Push(tasks...); pushed != len(tasks) {
		// not expecting this to reach any time soon,
		// so leaving this error handling for later
		panic("queue full")
	}

	if e.state == RunningState {
		e.startWorkersLocked()
	}
}

func (e *Executor[T, W]) startWorkersLocked() {
	for e.activeWorkers < e.maxWorkers && e.tasks.Len() > 0 {
		e.assertInvariants()

		e.activeWorkers++
		go e.worker()
	}
}

func (e *Executor[T, W]) worker() {
	defer e.workerDone()

	for {
		e.mu.Lock()

		if e.ctx.Err() != nil {
			e.tasks.Clear()
			e.mu.Unlock()

			return
		}

		task, ok := e.tasks.Pop()
		e.mu.Unlock()

		if !ok {
			return
		}

		next := e.newWorker().Exec(task)

		if e.ctx.Err() == nil {
			e.Enqueue(next...)
		}
	}
}

func (e *Executor[T, W]) workerDone() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.activeWorkers--

	if e.ctx.Err() != nil {
		e.tasks.Clear()
	}

	if e.activeWorkers == 0 && e.tasks.Len() == 0 {
		e.finishLocked()
	}
}

func (e *Executor[T, W]) finishLocked() {
	if e.state == DoneState {
		return
	}

	e.state = DoneState
	e.tasks.Clear()
	close(e.done)
}

// RemainingTasks to be executed (may go up or down).
func (e *Executor[T, W]) RemainingTasks() int {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.tasks.Len()
}

func (e *Executor[T, W]) assertInvariants() {
	if e.activeWorkers < 0 {
		panic("activeWorkers < 0")
	}

	if e.activeWorkers > e.maxWorkers {
		panic("activeWorkers > maxWorkers")
	}
}
