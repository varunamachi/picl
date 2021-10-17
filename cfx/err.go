package cfx

import "fmt"

type Error struct {
	Err error  `json:"err"`
	Msg string `json:"msg"`
}

func (fxErr *Error) Error() string {
	return fxErr.Err.Error()
}

func (fxErr *Error) Unwrap() error {
	return fxErr.Err
}

func (cfx *Error) String() string {
	return cfx.Err.Error() + ": " + cfx.Msg
}

func Errf(inner error, msg string, args ...interface{}) *Error {
	msg = fmt.Sprintf(msg, args...)
	return &Error{Err: inner, Msg: msg}
}
