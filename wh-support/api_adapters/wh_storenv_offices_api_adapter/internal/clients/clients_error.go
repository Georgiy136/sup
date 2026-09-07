package clients

import "errors"

var (
	ErrResponseNil     = errors.New("response is nil")
	ErrWrongStatusCode = errors.New("wrong status code")
)
