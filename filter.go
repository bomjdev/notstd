package notstd

type FilterFn[T any] func(v T) bool

func (fn FilterFn[T]) And(fn2 FilterFn[T]) FilterFn[T] {
	return func(v T) bool {
		return fn(v) && fn2(v)
	}
}

func (fn FilterFn[T]) Or(fn2 FilterFn[T]) FilterFn[T] {
	return func(v T) bool {
		return fn(v) || fn2(v)
	}
}

func (fn FilterFn[T]) Not() FilterFn[T] {
	return func(v T) bool {
		return !fn(v)
	}
}

func AnyFilter[T any](fns ...FilterFn[T]) FilterFn[T] {
	return func(v T) bool {
		for _, fn := range fns {
			if fn(v) {
				return true
			}
		}
		return false
	}
}

func AllFilter[T any](fns ...FilterFn[T]) FilterFn[T] {
	return func(v T) bool {
		for _, fn := range fns {
			if !fn(v) {
				return false
			}
		}
		return true
	}
}
