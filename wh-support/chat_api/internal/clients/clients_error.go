package clients

import (
	"errors"
	"fmt"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
)

var (
	errResponseNil = errors.New("response is nil")
)

type DataError struct {
	Errors []CustomErrorWrapper `json:"errors"`
}

type CustomErrorWrapper struct {
	rest_data.CustomError
}

func (c CustomErrorWrapper) Error() string {
	return fmt.Sprintf("%s: %v", c.ErrorKey, c.Message)
}

func (d DataError) Error() string {
	joinedErr := errors.Join(d.ConvertToErrorSlice()...)
	if joinedErr == nil {
		return "empty data error"
	}
	return joinedErr.Error()
}

func (d DataError) ConvertToErrorSlice() []error {
	errs := make([]error, len(d.Errors))
	for i, e := range d.Errors {
		errs[i] = e
	}
	return errs
}
