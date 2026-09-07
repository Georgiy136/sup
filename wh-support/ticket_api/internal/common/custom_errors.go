package common

import (
	"errors"
	"fmt"
	"net/http"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
)

var (
	ErrResponseNil            = errors.New("response is nil")
	ErrTicketNotFound         = errors.New("ticket not found")
	ErrNoCategoryStatusAccess = errors.New("no category status access")
)

type DetailErrors struct {
	Errors []CustomErrorWrapper `json:"errors"`
}

type CustomErrorWrapper struct {
	rest_data.CustomError
}

func (e CustomErrorWrapper) Error() string {
	return fmt.Sprintf("%s: %v", e.ErrorKey, e.Message)
}

func AsBusinessError(err error) *CustomErrorWrapper {
	if bizErr, ok := errors.AsType[*CustomErrorWrapper](err); ok {
		return bizErr
	}
	return nil
}

func ToRestData(err error) *rest_data.RestData {
	if bizErr := AsBusinessError(err); bizErr != nil {
		return &rest_data.RestData{
			Errors:         []rest_data.CustomError{bizErr.CustomError},
			HttpResultCode: http.StatusUnprocessableEntity,
		}
	}
	return nil
}
