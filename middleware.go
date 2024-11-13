package notstd

type Middleware[T any] func(T) T

func WithMiddleware[T any](v T, middleware ...Middleware[T]) T {
	for _, mw := range middleware {
		v = mw(v)
	}
	return v
}
