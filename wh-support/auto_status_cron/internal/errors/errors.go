package internalerrors

import (
	"errors"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

var (
	ErrUnprocessableEntity = errors.New("unprocessable entity")

	ErrInternalServerError = errors.New("internal server error")

	ErrUrlsEmpty = errors.New("urls empty")

	ErrResponseNil = errors.New("response is nil")
)

const CommentErrorDuringPerformTicket = "Во время исполнения заявки возникла ошибка"

type Errors struct {
	Data []ErrorWithMsg `json:"errors"`
}

type ExternalError struct {
	Data ErrorWithMsg `json:"error"`
}

type ErrorWithMsg struct {
	ErrKey        string            `json:"error"`
	Msg           string            `json:"message"`
	MsgFmt        string            `json:"message_fmt"`
	MessageValues map[string]string `json:"message_values"`
	Detail        string            `json:"detail"`
}

var defaultErr = ErrorWithMsg{
	ErrKey: support_err_keys.KeyErrorExecutionFailed,
	Msg:    CommentErrorDuringPerformTicket,
}

func (e ErrorWithMsg) Error() string {
	return fmt.Sprintf("%s: %v", e.Msg, e.ErrKey)
}

func ParseErrorMessageFromResponse(response *fasthttp.Response) (ErrorWithMsg, error) {
	if response == nil {
		return defaultErr, errors.New("response is nil")
	}

	var errs Errors
	err := jsoniter.Unmarshal(response.Body(), &errs)
	if err != nil {
		return defaultErr, fmt.Errorf("error unmarshaling: %w", err)
	}

	if len(errs.Data) == 0 || len(errs.Data[0].ErrKey) == 0 {
		return defaultErr, nil
	}

	return errs.Data[0], nil
}

func ParseExternalErrorMessageFromResponse(response *fasthttp.Response) (ErrorWithMsg, error) {
	if response == nil {
		return defaultErr, errors.New("response is nil")
	}

	var errs ExternalError
	err := jsoniter.Unmarshal(response.Body(), &errs)
	if err != nil {
		return defaultErr, fmt.Errorf("error unmarshaling: %w", err)
	}

	return errs.Data, nil
}
