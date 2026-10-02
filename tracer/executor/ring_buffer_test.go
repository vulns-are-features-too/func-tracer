package executor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vulns-are-features-too/func-tracer/tracer/executor"
	"pgregory.net/rapid"
)

func TestRingBufferEmpty(t *testing.T) {
	t.Parallel()

	rb := rb()

	assert.Equal(t, 0, rb.Len())

	_, ok := rb.Pop()
	assert.False(t, ok)
}

func TestRingBufferPushPop(t *testing.T) {
	t.Parallel()

	vals := []int{1, 2, 3}

	t.Run("1-by-1", func(t *testing.T) {
		t.Parallel()

		rb := rb()

		for _, val := range vals {
			pushed := rb.Push(val)
			assert.Equal(t, 1, pushed)
		}

		assertPops(t, rb, vals)
	})

	t.Run("variadic", func(t *testing.T) {
		t.Parallel()

		rb := rb()
		pushed := rb.Push(vals...)
		assert.Equal(t, len(vals), pushed)

		assertPops(t, rb, vals)
	})
}

func TestRingBufferPushEmpty(t *testing.T) {
	t.Parallel()

	rb := rb()

	pushed := rb.Push()

	assert.Equal(t, 0, pushed)
	assert.Equal(t, 0, rb.Len())
}

func TestRingBufferGrowInPushLoop(t *testing.T) {
	t.Parallel()

	sut := runner(t)
	rb := sut.rb()

	prevCap := rb.Capacity()

	sut.push1by1(3)

	assert.Greater(t, rb.Capacity(), prevCap)

	prevCap = rb.Capacity()

	sut.push1by1(20)

	assert.Greater(t, rb.Capacity(), prevCap)
}

func TestRingBufferGrowInBatchPush(t *testing.T) {
	t.Parallel()

	sut := runner(t)
	rb := sut.rb()

	prevCap := rb.Capacity()

	sut.pushBatch(3)

	assert.Greater(t, rb.Capacity(), prevCap)

	prevCap = rb.Capacity()

	sut.pushBatch(20)

	assert.Greater(t, rb.Capacity(), prevCap)
}

func TestRingBufferWrapWithoutGrow(t *testing.T) {
	t.Parallel()

	const capacity = 16

	sut := runner(t)
	rb := sut.rb()

	// fill to a capacity
	sut.push1by1(capacity)
	assert.Equal(t, capacity, rb.Capacity())

	// move head to middle
	sut.pop(8)
	assert.Equal(t, capacity, rb.Capacity())

	// trigger wrap
	sut.pushBatch(4)
	assert.Equal(t, capacity, rb.Capacity())

	// verify
	sut.pop(12)
	assert.Equal(t, 0, rb.Len())
	assert.Equal(t, capacity, rb.Capacity())
}

func TestRingBufferWrapAndGrow(t *testing.T) {
	t.Parallel()

	const initCapacity = 16

	sut := runner(t)
	rb := sut.rb()

	// fill to a capacity
	sut.push1by1(initCapacity)
	assert.Equal(t, initCapacity, rb.Capacity())

	// move head to middle
	sut.pop(8)

	// wrap & grow
	sut.pushBatch(24)
	assert.Greater(t, rb.Capacity(), initCapacity)

	// verify
	sut.pop(32)
	assert.Equal(t, 0, rb.Len())
	assert.Greater(t, rb.Capacity(), initCapacity)
}

func TestRingBufferDontGrowIfPopMakesSpace(t *testing.T) {
	t.Parallel()

	const (
		initCap = 4
		iters   = 100
	)

	rb := rb()

	for i := range initCap {
		rb.Push(i)
	}

	for i := range iters {
		_, ok := rb.Pop()
		assert.True(t, ok)

		pushed := rb.Push(i)
		assert.Equal(t, 1, pushed)

		assert.Equal(t, initCap, rb.Capacity())
	}

	rb.Push(9999)
	assert.Equal(t, initCap*2, rb.Capacity())
}

func TestRingBufferZeroValues(t *testing.T) {
	t.Parallel()

	rb := executor.NewRingBuffer[*int](1)

	a := 1
	b := 2

	rb.Push(&a, nil, &b)

	got, ok := rb.Pop()
	require.True(t, ok)
	assert.Equal(t, &a, got)

	got, ok = rb.Pop()
	require.True(t, ok)
	assert.Nil(t, got)

	got, ok = rb.Pop()
	require.True(t, ok)
	assert.Equal(t, &b, got)

	got, ok = rb.Pop()
	assert.False(t, ok)
	assert.Nil(t, got)
}

func TestRingBufferProp(t *testing.T) {
	t.Parallel()

	iterations := 10_000
	maxPerPush := uint8(50)

	if testing.Short() {
		iterations = 100
		maxPerPush = 10
	}

	rapid.Check(t, func(t *rapid.T) {
		sut := runner(t)
		actions := rapid.Uint8Range(0, 2)
		counts := rapid.Uint8Range(0, maxPerPush)

		for range iterations {
			count := int(counts.Draw(t, "count"))
			switch actions.Draw(t, "action") {
			case 0:
				sut.push1by1(count)
			case 1:
				sut.pushBatch(count)
			case 2:
				sut.popIfNotEmpty(count)
			}
		}

		assert.Equal(t, sut.expectedLen(), sut.rb().Len())
	})
}

func rb() executor.RingBuffer[int] {
	return executor.NewRingBuffer[int](1)
}

func assertPops[T any](t *testing.T, rb executor.RingBuffer[T], wants []T) {
	t.Helper()

	assert.Equal(t, len(wants), rb.Len())

	for i, want := range wants {
		got, ok := rb.Pop()

		require.True(t, ok, "Pop() failed at %d=%d", i, want)
		assert.Equal(t, want, got)
	}

	assert.Equal(t, 0, rb.Len())

	_, ok := rb.Pop()
	assert.False(t, ok, "Pop() returned ok=true after being drained")
}

type stepRunner struct {
	t        rapid.TB
	buf      executor.RingBuffer[int]
	nextPush int
	nextPop  int
}

func runner(t rapid.TB) *stepRunner {
	t.Helper()

	return &stepRunner{
		t:        t,
		buf:      rb(),
		nextPush: 1,
		nextPop:  1,
	}
}

func (r *stepRunner) rb() *executor.RingBuffer[int] {
	return &r.buf
}

func (r *stepRunner) push1by1(times int) {
	r.t.Helper()

	for range times {
		pushed := r.buf.Push(r.nextPush)
		if pushed != 1 {
			r.t.Fatalf("push failed val=%d pushed=%d", r.nextPush, pushed)
		}

		r.nextPush++
	}
}

func (r *stepRunner) pushBatch(times int) {
	r.t.Helper()

	vals := make([]int, times)
	for i := range vals {
		vals[i] = r.nextPush
		r.nextPush++
	}

	pushed := r.buf.Push(vals...)
	if pushed != times {
		r.t.Fatalf("push failed from=%d to=%d pushed=%d", r.nextPush-times, r.nextPush, pushed)
	}
}

func (r *stepRunner) popIfNotEmpty(times int) {
	r.t.Helper()

	for range times {
		got, ok := r.buf.Pop()

		if ok {
			if got != r.nextPop {
				r.t.Fatalf("popped wrong result: expected %d, got %d", r.nextPop, got)
			}

			r.nextPop++
		}
	}
}

func (r *stepRunner) pop(times int) {
	r.t.Helper()

	for range times {
		got, ok := r.buf.Pop()
		if !ok {
			r.t.Fatalf("Pop() failed at %d", r.nextPop)
		}

		if got != r.nextPop {
			r.t.Fatalf("popped wrong result: expected %d, got %d", r.nextPop, got)
		}

		r.nextPop++
	}
}

func (r *stepRunner) expectedLen() int {
	return r.nextPush - r.nextPop
}
