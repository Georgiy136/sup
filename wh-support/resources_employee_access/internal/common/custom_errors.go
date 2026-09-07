package common

import "errors"

var (
	ErrEmployeeUnknown = errors.New("can't find employee")
	ErrResponseNil     = errors.New("response is nil")
)
