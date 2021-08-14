package cfx

import "fmt"

type FileError struct {
	Err error  `json:"err"`
	Msg string `json:"msg"`
}

func (fxErr *FileError) Error() string {
	return fxErr.Err.Error()
}

func (fxErr *FileError) Unwrap() error {
	return fxErr.Err
}

func (cfx *FileError) String() string {
	return cfx.Err.Error() + ": " + cfx.Msg
}

func FileErrf(inner error, msg string, args ...interface{}) *FileError {
	msg = fmt.Sprintf(msg, args...)
	return &FileError{Err: inner, Msg: msg}
}
