package postgres

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	modelsPostgres "github.com/surahman/FTeX/pkg/models/postgres"
)

// UserRegister is the interface through which external methods can create a user.
func (p *postgresImpl) UserRegister(userDetails *modelsPostgres.UserAccount) (uuid.UUID, error) {
	params := userCreateParams{
		Username:  userDetails.Username,
		Password:  userDetails.Password,
		FirstName: userDetails.FirstName,
		LastName:  userDetails.LastName,
		Email:     userDetails.Email,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:gomnd

	defer cancel()

	clientID, err := p.Query.userCreate(ctx, &params)
	if err != nil {
		return uuid.UUID{}, ErrRegisterUser
	}

	return clientID, nil
}

// UserCredentials is the interface through which external methods can retrieve user credentials.
func (p *postgresImpl) UserCredentials(username string) (uuid.UUID, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:gomnd

	defer cancel()

	credentials, err := p.Query.userGetCredentials(ctx, username)
	if err != nil {
		return uuid.UUID{}, "", ErrLoginUser
	}

	return credentials.ClientID, credentials.Password, nil
}
