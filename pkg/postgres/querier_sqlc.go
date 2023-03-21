// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.0

package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type Querier interface {
	CreateUser(ctx context.Context, arg CreateUserParams) (pgtype.UUID, error)
	DeleteUser(ctx context.Context, username string) error
	GetClientIdUser(ctx context.Context, username string) (pgtype.UUID, error)
	GetCredentialsUser(ctx context.Context, username string) (GetCredentialsUserRow, error)
}

var _ Querier = (*Queries)(nil)