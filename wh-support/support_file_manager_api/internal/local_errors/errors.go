package local_errors

import (
	"errors"
	"fmt"
	"net/http"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
)

type BusinessError struct {
	Code    string
	Message string
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

var (
	ErrFileNotFound   = &BusinessError{Code: "file.not_found", Message: "file not found"}
	ErrUploadNotFound = &BusinessError{Code: "upload.not_found", Message: "upload not found"}
)

func AsBusinessError(err error) *BusinessError {
	if bizErr, ok := errors.AsType[*BusinessError](err); ok {
		return bizErr
	}
	return nil
}

func ToRestData(err error) *rest_data.RestData {
	if bizErr := AsBusinessError(err); bizErr != nil {
		return &rest_data.RestData{
			Errors: []rest_data.CustomError{{
				ErrorKey: bizErr.Code,
				Message:  bizErr.Message,
			}},
			HttpResultCode: http.StatusUnprocessableEntity,
		}
	}
	return nil
}

// HandleRestDataError обрабатывает ошибку из RestData.
// Для бизнес-ошибок (422) возвращает BusinessError, для остальных — обычную ошибку.
func HandleRestDataError(rd rest_data.RestData, operation string) error {
	if !rd.HasError() {
		return nil
	}

	if rd.HttpResultCode == http.StatusUnprocessableEntity {
		return &BusinessError{
			Code:    rd.Errors[0].ErrorKey,
			Message: rd.Errors[0].Message,
		}
	}
	return fmt.Errorf("%s: %w", operation, rest_data.ToError(rd))
}
