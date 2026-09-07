package errors

import "fmt"

type CustomError struct {
	StatusCode int
	Message    string
	Body       string
}

func NewCustomError(statusCode int, message string) *CustomError {
	return &CustomError{
		StatusCode: statusCode,
		Message:    message,
	}
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Body)
}
