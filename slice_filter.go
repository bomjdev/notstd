package notstd

func Filter[T any, S ~[]T](slice S, fn FilterFn[T]) S {
	ret := make([]T, 0, len(slice))
	for _, s := range slice {
		if fn(s) {
			ret = append(ret, s)
		}
	}
	return ret
}

func NewSliceFilter[T any, S ~[]T](fn FilterFn[T]) func(S) S {
	return func(slice S) S {
		return Filter(slice, fn)
	}
}
