package jsend

import (
    "errors"
)

const (
    StatusSuccess = "success"
    StatusError = "error"
    StatusFailed = "failed"
)

var (
    ErrNoDataType = errors.New("data-type cannot be empty")
)


type ResponseDTO[T any] struct {
	Status  string         `json:"status"`
	Message string         `json:"message,omitmepty"`
	Data    map[string][]T `json:"data,omitempty"`
	Code    string         `json:"code,omitempty"`
}

func NewSuccessResponse[T any](dataType string, data []T, message string) (*ResponseDTO[T], error) {
	if dataType == "" {
	    return nil, ErrNoDataType
	}

	return &ResponseDTO[T]{
	    Status: StatusSuccess,
	    Message: message,
	    Data: map[string][]T{dataType: data},
	}, nil
}
