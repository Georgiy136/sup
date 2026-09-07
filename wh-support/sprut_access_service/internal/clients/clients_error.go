package clients

import (
	"errors"
)

var (
	ErrInternalServerError = errors.New("internal server error")
	ErrResponseNil         = errors.New("response is nil")
	ErrWrongStatusCode     = errors.New("wrong status code")
	ErrTranslationNotFound = errors.New("translation not found")
)
