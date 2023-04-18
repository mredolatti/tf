package jsend

import (
	"errors"
)

const (
	StatusSuccess = "success"
	StatusError   = "error"
	StatusFailed  = "fail"
)

var (
	ErrNoDataType = errors.New("data-type cannot be empty")
)

type ResponseDTO[T any] struct {
	Status  string       `json:"status"`
	Message string       `json:"message,omitmepty"`
	Data    map[string]T `json:"data,omitempty"`
	Code    string       `json:"code,omitempty"`
}

func NewSuccessResponse[T any](dataType string, data T, message string) *ResponseDTO[T] {
	return &ResponseDTO[T]{
		Status:  StatusSuccess,
		Message: message,
		Data:    map[string]T{dataType: data},
	}
}

func NewCustomFailResponse(message string, key string, value string) *ResponseDTO[string] {
	return &ResponseDTO[string]{
		Status:  StatusFailed,
		Message: message,
		Data:    map[string]string{key: value},
	}
}

func NewReadBodyFailResponse(err error) *ResponseDTO[string] {
	return NewCustomFailResponse("", "error", err.Error())
}

func NewErrorResponse(why string) *ResponseDTO[string] {
	return &ResponseDTO[string]{
		Status:  StatusError,
		Message: why,
	}
}

var (
	ResponseEmptySuccess   = &ResponseDTO[string]{Status: StatusSuccess}
	ResponseErrorInSession = NewErrorResponse("internal authentication error")
	ResponseFailToReadBody = NewCustomFailResponse("", "request", "unable to read valid json from request body")
)
