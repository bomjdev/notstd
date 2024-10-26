package notstd

type MapperFn[In, Out any] func(v In) Out

func Map[In, Out any](slice Slice[In], fn MapperFn[In, Out]) Slice[Out] {
	ret := make([]Out, 0, len(slice))
	for _, s := range slice {
		ret = append(ret, fn(s))
	}
	return ret
}

func NewSliceMapper[In, Out any](fn MapperFn[In, Out]) func(Slice[In]) Slice[Out] {
	return func(slice Slice[In]) Slice[Out] {
		return Map(slice, fn)
	}
}
