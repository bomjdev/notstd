package notstd

type FilterFn[T any] func(v T) bool

func Filter[T any](slice Slice[T], fn FilterFn[T]) Slice[T] {
	ret := make([]T, 0, len(slice))
	for _, s := range slice {
		if fn(s) {
			ret = append(ret, s)
		}
	}
	return ret
}

func NewSliceFilter[T any](fn FilterFn[T]) func(Slice[T]) Slice[T] {
	return func(slice Slice[T]) Slice[T] {
		return Filter(slice, fn)
	}
}
