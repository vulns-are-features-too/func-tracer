package executor

// RingBuffer as a queue that auto-grows to accommodate items.
type RingBuffer[T any] struct {
	buf  []T
	head int
	tail int
	size int
}

// NewRingBuffer a ring buffer with some initial capacity.
func NewRingBuffer[T any](initialCapacity int) RingBuffer[T] {
	return RingBuffer[T]{
		buf: make([]T, initialCapacity),
	}
}

// Len returns the number of elements in the buffer.
func (rb *RingBuffer[T]) Len() int {
	if rb == nil {
		return 0
	}

	return rb.size
}

// Capacity of the buffer.
func (rb *RingBuffer[T]) Capacity() int {
	if rb == nil {
		return 0
	}

	return len(rb.buf)
}

// Push adds elements & returns the number of pushed elements.
// Should return len(values) unless the buffer is full,
// in which case nothing is pushed and 0 is returned.
func (rb *RingBuffer[T]) Push(values ...T) int {
	if rb == nil {
		return 0
	}

	vLen := len(values)
	if vLen == 0 {
		return 0
	}

	required := rb.size + vLen
	if required > rb.Capacity() {
		ok := rb.grow(required)
		if !ok {
			return 0
		}
	}

	// Copy as much as possible in 1 go.
	// If the values don't fit in the tail space,
	// the remaining values wrap around the ring.
	//
	// BUFFER:
	//      unused        used     tail space
	// |--------------|+++++++++++|-----------|
	//           head ^           ^ tail
	// VALUES to be added:                    v split
	//                            |++++++++++++++++++++|
	// RESULT:
	// |+++++++++|----|+++++++++++++++++++++++|
	//  ^ wrapped around
	capa := rb.Capacity()
	split := min(vLen, capa-rb.tail)
	copy(rb.buf[rb.tail:rb.tail+split], values[:split])

	remaining := vLen - split
	if remaining > 0 {
		copy(rb.buf[:remaining], values[split:])
	}

	rb.tail = (rb.tail + vLen) % capa
	rb.size += vLen

	rb.assertInvariants()

	return vLen
}

// Pop an item from the buffer,
// return (T, false) if empty
//
//nolint:ireturn
func (rb *RingBuffer[T]) Pop() (T, bool) {
	var zero T

	if rb == nil {
		return zero, false
	}

	if rb.size == 0 {
		return zero, false
	}

	value := rb.buf[rb.head]

	// not strictly needed but useful in mem dump
	rb.buf[rb.head] = zero

	rb.head++
	if rb.head == rb.Capacity() {
		rb.head = 0
	}

	rb.size--

	rb.assertInvariants()

	return value, true
}

// Clear buffer.
func (rb *RingBuffer[T]) Clear() {
	if rb.size == 0 {
		return
	}

	rb.buf = rb.buf[:0]
	rb.head = 0
	rb.tail = 0
	rb.size = 0
}

// grow to accommodate at least `required` values
// size increases in powers of 2
//
// returns whether or not it actually grew
// (false if capacity is already max int).
func (rb *RingBuffer[T]) grow(required int) bool {
	const growthFactor = 2

	oldCap := rb.Capacity()
	newCap := max(oldCap*growthFactor, required)
	newBuf := make([]T, newCap)

	if rb.size > 0 {
		split := min(rb.size, oldCap-rb.head)
		copy(newBuf, rb.buf[rb.head:rb.head+split])

		remaining := rb.size - split
		if remaining > 0 {
			copy(newBuf[split:], rb.buf[:remaining])
		}
	}

	rb.buf = newBuf
	rb.head = 0
	rb.tail = rb.size

	return true
}

func (rb *RingBuffer[T]) assertInvariants() {
	if rb.head < 0 {
		panic("negative head")
	}

	if rb.tail < 0 {
		panic("negative tail")
	}

	if rb.size < 0 {
		panic("negative size")
	}

	if rb.tail > len(rb.buf) {
		panic("tail > capacity")
	}

	if rb.size > len(rb.buf) {
		panic("size > capacity")
	}
}
