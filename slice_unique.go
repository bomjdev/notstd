package notstd

func Unique[T comparable, S ~[]T](slice S) S {
	return NewSet(slice).Slice()
}
