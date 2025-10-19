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

func NilOrFilter[T any](fn FilterFn[T]) FilterFn[*T] {
	return func(v *T) bool {
		if v == nil {
			return true
		}
		return fn(*v)
	}
}

func NotNilAndFilter[T any](fn FilterFn[T]) FilterFn[*T] {
	return func(v *T) bool {
		if v == nil {
			return false
		}
		return fn(*v)
	}
}

func GetterFilter[T, V any](getter func(T) V, fn FilterFn[V]) FilterFn[T] {
	return func(v T) bool {
		return fn(getter(v))
	}
}

func GetterFilterFactory[T, V any](getter func(T) V) func(FilterFn[V]) FilterFn[T] {
	return func(fn FilterFn[V]) FilterFn[T] {
		return func(v T) bool {
			return fn(getter(v))
		}
	}
}

func NilGetterFilter[T, V any](getter func(T) *V, fn FilterFn[V]) FilterFn[T] {
	return func(v T) bool {
		return NilOrFilter(fn)(getter(v))
	}
}

type FilterFactory[T, V any] func(fn FilterFn[V]) FilterFn[T]
