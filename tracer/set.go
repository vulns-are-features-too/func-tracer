package tracer

import "sync"

type set struct {
	mu   sync.Mutex
	data map[string]struct{}
}

func newSet() *set {
	return &set{
		sync.Mutex{},
		make(map[string]struct{}),
	}
}

func (s *set) add(key string) {
	s.mu.Lock()
	s.data[key] = struct{}{}
	s.mu.Unlock()
}

func (s *set) has(key string) bool {
	s.mu.Lock()
	_, ok := s.data[key]
	s.mu.Unlock()

	return ok
}
