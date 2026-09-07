package serviceerrors

import "errors"

var (
	ErrObjectNotFound = errors.New("object not found")
	ErrResponseNil    = errors.New("response is nil")
)
