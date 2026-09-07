package errors

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrEmptyData      = errors.New("empty response data")
	ErrNilValue       = errors.New("nil value")
	ErrNilResponse    = errors.New("nil response")
	ErrUserNotFound   = errors.New("user not found")
	ErrTooManyUsers   = errors.New("too many users")
	ErrHistoryExpired = errors.New("stream history expired")
	ErrKeyNotFound    = errors.New("key not found")
	ErrInvalidAction  = errors.New("invalid action")
	ErrUnknownAction  = errors.New("unknown action")
)

const StreamErrorCodeHistoryExpired int64 = 112

type DBActionError struct {
	StatusCode int
	ErrorKey   string
	Message    string
	Detail     string
}

func NewDBActionError(statusCode int, errorKey, message, detail string) *DBActionError {
	return &DBActionError{
		StatusCode: statusCode,
		ErrorKey:   errorKey,
		Message:    message,
		Detail:     detail,
	}
}

func (e *DBActionError) Error() string {
	return fmt.Sprintf("db returned status %d error %s: %s; detail: %s", e.StatusCode, e.ErrorKey, e.Message, e.Detail)
}

type InteractiveDialogValidationError struct {
	Errors map[string]string
}

func NewInteractiveDialogValidationError(validationErrors map[string]string) *InteractiveDialogValidationError {
	return &InteractiveDialogValidationError{Errors: validationErrors}
}

func (e *InteractiveDialogValidationError) Error() string {
	messages := make([]string, 0, len(e.Errors))
	for key, val := range e.Errors {
		if val == "" {
			continue
		}
		messages = append(messages, fmt.Sprintf("%s - %s", key, val))
	}
	return strings.Join(messages, "; ")
}

type StreamError struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

func NewStreamError(code int64, message string) error {
	return &StreamError{
		Code:    code,
		Message: message,
	}
}

func (c StreamError) Error() string {
	return fmt.Sprintf("%d: %s", c.Code, c.Message)
}

type SubmissionFieldValidationError struct {
	ValidationError string
}

func NewSubmissionFieldValidationError(validationError string) *SubmissionFieldValidationError {
	return &SubmissionFieldValidationError{ValidationError: validationError}
}

func (s *SubmissionFieldValidationError) Error() string {
	return s.ValidationError
}
