package notstd

type EqualFn[T any] func(T, T) bool

func (fn EqualFn[T]) NewFilter(target T) FilterFn[T] {
	return func(v T) bool {
		return fn(target, v)
	}
}
