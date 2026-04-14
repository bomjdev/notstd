package notstd

func NewMapFunc[T any, K comparable, S ~[]T](slice S, keyFn KeyFn[T, K]) map[K]T {
	return NewMapKVFunc(slice, keyFn, func(v T) T { return v })
}

type ValueFn[T, V any] func(T) V

func NewMapKVFunc[T any, K comparable, V any, S ~[]T](slice S, keyFn KeyFn[T, K], valueFn ValueFn[T, V]) map[K]V {
	s := make(map[K]V, len(slice))
	for _, v := range slice {
		s[keyFn(v)] = valueFn(v)
	}
	return s
}

func MapAppend[T any, K comparable, S ~[]T](m map[K]S, k K, v T) {
	s := m[k]
	m[k] = append(s, v)
}

type DefaultMap[T any, K comparable] struct {
}
