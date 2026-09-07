package clients

import (
	"errors"
	"fmt"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
)

type DataError struct {
	Errors []CustomErrorWrapper `json:"errors"`
}

type CustomErrorWrapper struct {
	rest_data.CustomError
}

func (e CustomErrorWrapper) Error() string {
	return fmt.Sprintf("%s: %v", e.ErrorKey, e.Message)
}

var (
	ErrResponseNil     = errors.New("response is nil")
	ErrWrongStatusCode = errors.New("wrong status code")
)
