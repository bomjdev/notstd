package notstd

import "errors"

func ErrorAs[T error](err error) (T, bool) {
	var v T
	if errors.As(err, &v) {
		return v, true
	}
	return v, false
}
