package notstd

type MapperFn[In, Out any] func(v In) Out

func Map[I, O any, S ~[]I](slice S, fn MapperFn[I, O]) []O {
	ret := make([]O, 0, len(slice))
	for _, s := range slice {
		ret = append(ret, fn(s))
	}
	return ret
}

func NewSliceMapper[I, O any, S ~[]I](fn MapperFn[I, O]) func(S) []O {
	return func(slice S) []O {
		return Map(slice, fn)
	}
}
