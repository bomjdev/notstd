package notstd

import "sync"

type Store[K comparable, V any] struct {
	m map[K]V
	sync.RWMutex
}

func NewStore[K comparable, V any](m map[K]V) *Store[K, V] {
	if m == nil {
		m = make(map[K]V)
	}
	return &Store[K, V]{
		m:       m,
		RWMutex: sync.RWMutex{},
	}
}

func (s *Store[K, V]) Get(key K) (V, bool) {
	s.RLock()
	defer s.RUnlock()
	v, ok := s.m[key]
	return v, ok
}

func (s *Store[K, V]) Set(key K, value V) {
	s.Lock()
	defer s.Unlock()
	s.m[key] = value
}

func (s *Store[K, V]) Delete(key K) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, key)
}

func (s *Store[K, V]) SetNoLock(key K, value V) {
	s.m[key] = value
}

func (s *Store[K, V]) GetNoLock(key K) (V, bool) {
	v, ok := s.m[key]
	return v, ok
}

func (s *Store[K, V]) DeleteNoLock(key K) {
	delete(s.m, key)
}

func (s *Store[K, V]) GetMap() map[K]V {
	s.RLock()
	defer s.RUnlock()
	return s.m
}

func (s *Store[K, V]) GetMapNoLock() map[K]V {
	return s.m
}
