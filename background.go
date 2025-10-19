package notstd

import (
	"context"
	"sync"
)

type Result[T any] struct {
	Result T
	Error  error
}

type ResultWaiterFn[T any] func(ctx context.Context) (Result[T], bool)

func ResultWaiterFactory[T any](resultCh <-chan Result[T]) ResultWaiterFn[T] {
	var (
		mu        sync.Mutex
		nextID    int
		subs      = make(map[int]chan<- Result[T])
		done      bool
		hasResult bool
		result    Result[T]
	)

	go func() {
		res, ok := <-resultCh
		mu.Lock()
		done = true
		if ok {
			result = res
			hasResult = true
		}
		for _, ch := range subs {
			if hasResult {
				ch <- result
			}
			close(ch)
		}
		mu.Unlock()
	}()

	return func(ctx context.Context) (Result[T], bool) {
		mu.Lock()
		if done {
			r := result
			ok := hasResult
			mu.Unlock()
			return r, ok
		}
		id := nextID
		nextID++
		ch := make(chan Result[T], 1)
		subs[id] = ch
		mu.Unlock()

		select {
		case <-ctx.Done():
			mu.Lock()
			delete(subs, id)
			mu.Unlock()
			return Result[T]{}, false
		case res, ok := <-ch:
			return res, ok
		}
	}
}

func WaitResult[T any](fn func() (T, error)) ResultWaiterFn[T] {
	resultCh := make(chan Result[T])
	go func() {
		res, err := fn()
		resultCh <- Result[T]{Result: res, Error: err}
		close(resultCh)
	}()
	return ResultWaiterFactory(resultCh)
}

func StoreResultWaiter[K comparable, V any](store *Store[K, ResultWaiterFn[V]], key K, fn func() (V, error), handler func(Result[V], bool)) {
	waiter := WaitResult(fn)

	store.Set(key, waiter)

	go func() {
		res, ok := waiter(context.Background())
		if handler != nil {
			handler(res, ok)
		}
		store.Delete(key)
	}()
}

func GetResultWaiter[K comparable, V any](ctx context.Context, store *Store[K, ResultWaiterFn[V]], key K) (Result[V], bool) {
	waitFn, hasKey := store.Get(key)
	if !hasKey {
		return Result[V]{}, false
	}
	return waitFn(ctx)
}
