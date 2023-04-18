package postgres

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	modelsPostgres "github.com/surahman/FTeX/pkg/models/postgres"
)

func (p *postgresImpl) CreateUser(userDetails *modelsPostgres.UserAccount) (uuid.UUID, error) {
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
