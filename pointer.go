package notstd

func Pointer[T any](v T) *T {
	return &v
}
