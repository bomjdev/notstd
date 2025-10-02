package notstd

func Reduce[I, O any, S ~[]I](slice S, acc O, fn func(O, I) O) O {
	for _, s := range slice {
		acc = fn(acc, s)
	}
	return acc
}
