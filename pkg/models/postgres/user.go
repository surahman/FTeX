package models

import (
	"github.com/gofrs/uuid"
)

// User represents a users account and is a row in user table.
type User struct {
	*UserAccount
	ClientID  uuid.UUID `json:"clientId,omitempty"`
	IsDeleted bool      `json:"isDeleted"`
}

// UserAccount is the core user account information.
type UserAccount struct {
	UserLoginCredentials
	FirstName string `json:"firstName,omitempty" validate:"required,max=64"`
	LastName  string `json:"lastName,omitempty" validate:"required,max=64"`
	Email     string `json:"email,omitempty" validate:"required,email,max=64"`
}

// UserLoginCredentials will contain the login credentials. This will also be used for login requests.
type UserLoginCredentials struct {
	Username string `json:"username,omitempty" validate:"required,min=8,max=32"`
	Password string `json:"password,omitempty" validate:"required,min=8,max=32"`
}
