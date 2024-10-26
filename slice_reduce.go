package notstd

func Reduce[In, Out any](slice Slice[In], acc Out, fn func(Out, In) Out) Out {
	for _, s := range slice {
		acc = fn(acc, s)
	}
	return acc
}
