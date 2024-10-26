package notstd

func Must[T any](v T, err error) T {
	if err != nil {
		panic(NewErrSrcSkip(err, 2))
	}
	return v
}
