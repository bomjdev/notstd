package notstd

import "sync"

type Pool[T any] func() T

func RoundRobin[T any](items []T) func() T {
	var idx int
	mu := new(sync.Mutex)
	return func() T {
		mu.Lock()
		item := items[idx]
		idx = (idx + 1) % len(items)
		mu.Unlock()
		return item
	}
}
