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

// Is will return whether the input error is an instance of expected error.
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
//
//nolint:lll
var (
	ErrRegisterUser          = errorRegisterUser()             // ErrorRegisterUser is returned if user registration fails.
	ErrLoginUser             = errorLoginUser()                // ErrLoginUser is returned if user credentials are not found.
	ErrNotFoundUser          = errorNotFoundUser()             // ErrNotFoundUser is returned if a user account is not found.
	ErrCreateFiat            = errorCreateFiat()               // ErrCreateFiat is returned if a Fiat account could not be opened.
	ErrTransactFiat          = errorTransactionFiat()          // ErrTransactFiat is returned if a Fiat transaction fails.
	ErrNotFound              = errorNotFound()                 // ErrNotFound is returned as a generic not found error.
	ErrUnhealthy             = errorUnhealthy()                // ErrUnhealthy is returned if the database cannot be pinged.
	ErrTransactCrypto        = errorTransactionCrypto()        // ErrTransactCrypto is returned if a Crypto transaction fails.
	ErrTransactCryptoDetails = errorTransactionCryptoDetails() // ErrTransactCryptoDetails is returned if a Crypto transaction succeeds, but transaction retrieval fails.
)

func errorRegisterUser() error {
	return &Error{
		Message: "username is already registered",
		Code:    http.StatusNotFound,
	}
}

func errorLoginUser() error {
	return &Error{
		Message: "invalid username or password",
		Code:    http.StatusForbidden,
	}
}

func errorNotFoundUser() error {
	return &Error{
		Message: "invalid username or password",
		Code:    http.StatusNotFound,
	}
}

func errorCreateFiat() error {
	return &Error{
		Message: "could not open Fiat account",
		Code:    http.StatusConflict,
	}
}

func errorTransactionFiat() error {
	return &Error{
		Message: "could not complete Fiat transaction",
		Code:    http.StatusInternalServerError,
	}
}

func errorNotFound() error {
	return &Error{
		Message: "records not found",
		Code:    http.StatusNotFound,
	}
}

func errorUnhealthy() error {
	return &Error{
		Message: "unhealthy",
		Code:    http.StatusServiceUnavailable,
	}
}

func errorTransactionCrypto() error {
	return &Error{
		Message: "could not complete Crypto transaction",
		Code:    http.StatusInternalServerError,
	}
}

func errorTransactionCryptoDetails() error {
	return &Error{
		Message: "could not retrieve transaction details for successful Crypto transaction",
		Code:    http.StatusInternalServerError,
	}
}
