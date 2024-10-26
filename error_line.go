package notstd

import (
	"errors"
	"fmt"
	"runtime"
)

type ErrSource struct {
	Err  error
	File string
	Line int
}

func (e ErrSource) Error() string {
	return fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Err)
}

func (e ErrSource) Is(err error) bool {
	return errors.Is(e.Err, err)
}

func NewErrSrc(err error) ErrSource {
	return NewErrSrcSkip(err, 2)
}

func NewErrSrcSkip(err error, skip int) ErrSource {
	_, file, line, _ := runtime.Caller(skip)
	return ErrSource{Err: err, Line: line, File: file}
}

func NewErrSrcSkipFactory(skip int) func(error) ErrSource {
	return func(err error) ErrSource {
		return NewErrSrcSkip(err, skip+1)
	}
}
