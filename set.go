package notstd

type Set[T comparable] map[T]struct{}

type KeyFn[T any, K comparable] func(T) K

func NewSet[T comparable](slice Slice[T]) Set[T] {
	set := make(Set[T], len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	return set
}

func NewSetFunc[T any, K comparable](slice Slice[T], keyFn KeyFn[T, K]) Set[K] {
	s := make(Set[K], len(slice))
	for _, v := range slice {
		s.Add(keyFn(v))
	}
	return s
}

func (s Set[T]) Add(v T) {
	s[v] = struct{}{}
}

func (s Set[T]) Contains(v T) bool {
	_, ok := s[v]
	return ok
}

func (s Set[T]) Delete(v T) {
	delete(s, v)
}

func (s Set[T]) Copy() Set[T] {
	ret := make(Set[T], len(s))
	for k := range s {
		ret[k] = struct{}{}
	}
	return ret
}

func (s Set[T]) Len() int {
	return len(s)
}

func (s Set[T]) Intersect(s2 Set[T]) Set[T] {
	ret := make(Set[T], len(s))
	for k := range s {
		if s2.Contains(k) {
			ret[k] = struct{}{}
		}
	}
	return ret
}

func (s Set[T]) Subtract(s2 Set[T]) Set[T] {
	ret := make(Set[T], len(s))
	for k := range s {
		if !s2.Contains(k) {
			ret[k] = struct{}{}
		}
	}
	return ret
}

func (s Set[T]) Union(s2 Set[T]) Set[T] {
	ret := make(Set[T], len(s)+len(s2))
	for k := range s {
		ret[k] = struct{}{}
	}
	for k := range s2 {
		ret[k] = struct{}{}
	}
	return ret
}

func (s Set[T]) Difference(s2 Set[T]) Set[T] {
	ret := make(Set[T], len(s))
	for k := range s {
		if !s2.Contains(k) {
			ret[k] = struct{}{}
		}
	}
	return ret
}

func (s Set[T]) Slice() Slice[T] {
	ret := make(Slice[T], 0, len(s))
	for k := range s {
		ret = append(ret, k)
	}
	return ret
}
