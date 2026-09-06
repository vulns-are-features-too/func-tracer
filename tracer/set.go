package tracer

type set map[string]struct{}

func (s set) add(key string) {
	s[key] = struct{}{}
}
