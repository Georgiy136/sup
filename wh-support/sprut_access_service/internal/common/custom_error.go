package common

import (
	"errors"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
)

var (
	MsgErrorBuildingExists = rest_data.CustomError{
		ErrorKey: "std.std.err_biz.building.already_exists",
		Message:  "Строение уже существует",
	}
)

type DetailErrors struct {
	Data []CustomErrorWrapper `json:"errors"`
}

type CustomErrorWrapper struct {
	rest_data.CustomError
}

func (e CustomErrorWrapper) Error() string {
	return fmt.Sprintf("%s: %v", e.ErrorKey, e.Message)
}

func ParseErrorMessageFromResponse(response *fasthttp.Response) (*CustomErrorWrapper, error) {
	var detailErr DetailErrors
	if err := jsoniter.Unmarshal(response.Body(), &detailErr); err != nil {
		return nil, fmt.Errorf("error unmarshaling: %w", err)
	}

	if len(detailErr.Data) == 0 {
		return nil, fmt.Errorf("no errors in response")
	}

	return &detailErr.Data[0], nil
}

func AsCustomError(err error) *CustomErrorWrapper {
	if customErrWrapper, ok := errors.AsType[*CustomErrorWrapper](err); ok {
		return customErrWrapper
	}
	return nil
}
