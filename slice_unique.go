package notstd

func Unique[T comparable](slice Slice[T]) Slice[T] {
	return NewSet(slice).Slice()
}
