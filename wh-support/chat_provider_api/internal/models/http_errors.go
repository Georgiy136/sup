package models

type HTTPError struct {
	StatusCode int
	Err        error
}

func (h *HTTPError) Error() string {
	return h.Err.Error()
}

func NewHTTPError(statusCode int, err error) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Err:        err,
	}
}
