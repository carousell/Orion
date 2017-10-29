package utils

import "github.com/carousell/go-utils/utils/errors"

type CustomError struct {
	Message    string
	StatusCode int
	Payload    interface{}
	errors.ErrorExt
}

func NewCustomError(msg string, code int, payload interface{}) CustomError {
	return CustomError{
		Message:    msg,
		StatusCode: code,
		Payload:    payload,
		ErrorExt:   errors.NewWithSkip(msg, 1),
	}
}
