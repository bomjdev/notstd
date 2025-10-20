package notstd

func NewMapFunc[T any, K comparable, S ~[]T](slice S, keyFn KeyFn[T, K]) map[K]T {
	s := make(map[K]T, len(slice))
	for _, v := range slice {
		s[keyFn(v)] = v
	}
	return s
}
