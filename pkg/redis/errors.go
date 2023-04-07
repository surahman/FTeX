package redis

import "errors"

// Error codes.
const (
	ErrorUnknown = iota
	ErrorCacheMiss
	ErrorCacheSet
	ErrorCacheDel
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
	return &Error{Message: message, Code: ErrorUnknown}
}

// errorCacheMiss will specialize the error as a cache miss.
func (e *Error) errorCacheMiss() *Error {
	e.Code = ErrorCacheMiss

	return e
}

// errorCacheSet will specialize the error as a cache set failure.
func (e *Error) errorCacheSet() *Error {
	e.Code = ErrorCacheSet

	return e
}

// errorCacheDel will specialize the error as a cache delete failure.
func (e *Error) errorCacheDel() *Error {
	e.Code = ErrorCacheDel

	return e
}
