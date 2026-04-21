package notstd

import "errors"

func ErrorAs[T error](err error) (T, bool) {
	var v T
	if err == nil {
		return v, false
	}
	if errors.As(err, &v) {
		return v, true
	}
	return v, false
}
