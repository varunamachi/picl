package xcutr

import "fmt"

type SshError struct {
	Err error  `json:"err"`
	Msg string `json:"msg"`
}

func (fxErr *SshError) Error() string {
	return fxErr.Err.Error() + ": " + fxErr.Msg
}

func (fxErr *SshError) Unwrap() error {
	return fxErr.Err
}

func (cfx *SshError) String() string {
	return cfx.Err.Error() + ": " + cfx.Msg
}

func NewErrf(inner error, msg string, args ...interface{}) *SshError {
	msg = fmt.Sprintf(msg, args...)
	return &SshError{Err: inner, Msg: msg}
}
