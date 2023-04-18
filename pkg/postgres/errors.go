package postgres

import (
	"errors"
	"net/http"
)

// Error is the base error type. The builder pattern is used to add specialization codes to the errors.
type Error struct {
	Message string
	Code    int
}

// Check to ensure the error interface is implemented.
var _ error = &Error{}

// Error get human readable error message.
func (e *Error) Error() string {
	return e.Message
}

// Is will return whether the input err is an instance of expected error.
func (e *Error) Is(err error) bool {
	var target *Error
	if !errors.As(err, &target) {
		return false
	}

	return e.Code == target.Code
}

// NewError is a base error message with no special code.
func NewError(message string) *Error {
	return &Error{Message: message, Code: 0}
}

// SetStatus will configure the status code within the error message.
func (e *Error) SetStatus(code int) *Error {
	e.Code = code

	return e
}

// Generic error variables.
// Errors to be returned by the Postgres Queries exposed through the interface for various failure conditions.
var (
	ErrRegisterUser = errorRegisterUser() // ErrorRegisterUser is returned is user registration fails.
)

func errorRegisterUser() error {
	return &Error{
		Message: "username is already registered",
		Code:    http.StatusConflict,
	}
}
