package redis

import (
	"errors"
)

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

// Generic error variables.
// Errors to be returned by the Postgres Queries exposed through the interface for various failure conditions.
var (
	ErrCacheUnknown = errorCacheUnknown() // ErrCacheUnknown is returned if an unknown error occurs in the cache.
	ErrCacheMiss    = errorCacheMiss()    // ErrCacheMiss is returned when a cache miss occurs.
	ErrCacheSet     = errorCacheSet()     // ErrCacheSet is returned if a key-value pair cannot be placed in the cache.
	ErrCacheDel     = errorCacheDel()     // ErrCacheDel is returned if a key-value pair cannot be deleted from the cache.
)

func errorCacheUnknown() error {
	return &Error{
		Message: "unknown Redis cache error",
		Code:    ErrorUnknown,
	}
}

func errorCacheMiss() error {
	return &Error{
		Message: "Redis cache miss",
		Code:    ErrorCacheMiss,
	}
}

func errorCacheSet() error {
	return &Error{
		Message: "Redis cache Set failure",
		Code:    ErrorCacheSet,
	}
}

func errorCacheDel() error {
	return &Error{
		Message: "Redis cache Del failure",
		Code:    ErrorCacheDel,
	}
}
