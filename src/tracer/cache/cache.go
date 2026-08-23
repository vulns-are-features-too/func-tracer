// Package cache provides function call caching
package cache

import "sync"

type entry[T any] struct {
	value T
	err   error
	ready chan struct{}
}

// Cache of function calls.
type Cache[T any] struct {
	mu      sync.Mutex
	entries map[string]*entry[T]
}

// New cache.
func New[T any]() *Cache[T] {
	return &Cache[T]{
		entries: make(map[string]*entry[T]),
	}
}

// GetOrExec gets result from cache, or exec func if key not cached.
//
//nolint:ireturn
func (c *Cache[T]) GetOrExec(
	key string,
	fn func() (T, error),
) (T, error) {
	c.mu.Lock()

	if entry, ok := c.entries[key]; ok {
		c.mu.Unlock()
		<-entry.ready

		return entry.value, entry.err
	}

	e := &entry[T]{
		ready: make(chan struct{}),
	}

	c.entries[key] = e
	c.mu.Unlock()

	e.value, e.err = fn()
	close(e.ready)

	return e.value, e.err
}
