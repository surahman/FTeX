package model_postgres

import "github.com/jackc/pgx/v5/pgtype"

// User represents a users account and is a row in user table.
type User struct {
	*UserAccount
	ClientID  pgtype.UUID `json:"client_id,omitempty"`
	IsDeleted bool        `json:"is_deleted"`
}

// UserAccount is the core user account information.
type UserAccount struct {
	UserLoginCredentials
	FirstName string `json:"first_name,omitempty" validate:"required,max=64"`
	LastName  string `json:"last_name,omitempty" validate:"required,max=64"`
	Email     string `json:"email,omitempty" validate:"required,email,max=64"`
}

// UserLoginCredentials will contain the login credentials. This will also be used for login requests.
type UserLoginCredentials struct {
	Username string `json:"username,omitempty" validate:"required,min=8,max=32"`
	Password string `json:"password,omitempty" validate:"required,min=8,max=32"`
}
