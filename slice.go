package notstd

import "slices"

type Slice[T any] []T

func NewSlice[T any](items ...T) Slice[T] {
	return items
}

func (s Slice[T]) Filter(fn FilterFn[T]) Slice[T] {
	return Filter(s, fn)
}

// Reverse returns a reversed copy of the slice
func (s Slice[T]) Reverse() Slice[T] {
	ret := slices.Clone(s)
	slices.Reverse(ret)
	return ret
}
