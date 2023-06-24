package postgres

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/surahman/FTeX/pkg/constants"
	modelsPostgres "github.com/surahman/FTeX/pkg/models/postgres"
	"go.uber.org/zap"
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

	ctx, cancel := context.WithTimeout(context.Background(), constants.ThreeSeconds())

	defer cancel()

	clientID, err := p.Query.userCreate(ctx, &params)
	if err != nil {
		p.logger.Error("failed to register user", zap.Error(err))

		return uuid.UUID{}, ErrRegisterUser
	}

	return clientID, nil
}

// UserCredentials is the interface through which external methods can retrieve user credentials.
func (p *postgresImpl) UserCredentials(username string) (uuid.UUID, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.ThreeSeconds())

	defer cancel()

	credentials, err := p.Query.userGetCredentials(ctx, username)
	if err != nil {
		p.logger.Error("failed to retrieve user account credentials", zap.Error(err))

		return uuid.UUID{}, "", ErrLoginUser
	}

	return credentials.ClientID, credentials.Password, nil
}

// UserGetInfo is the interface through which external methods can retrieve user account information.
func (p *postgresImpl) UserGetInfo(clientID uuid.UUID) (modelsPostgres.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.ThreeSeconds())

	defer cancel()

	userAccount, err := p.Query.userGetInfo(ctx, clientID)
	if err != nil {
		p.logger.Error("failed to retrieve user account info", zap.Error(err))

		return modelsPostgres.User{}, ErrNotFoundUser
	}

	return modelsPostgres.User{
			UserAccount: &modelsPostgres.UserAccount{
				UserLoginCredentials: modelsPostgres.UserLoginCredentials{
					Username: userAccount.Username,
					Password: userAccount.Password,
				},
				FirstName: userAccount.FirstName,
				LastName:  userAccount.LastName,
				Email:     userAccount.Email,
			},
			ClientID:  userAccount.ClientID,
			IsDeleted: userAccount.IsDeleted,
		},
		nil
}

// UserDelete is the interface through which external methods can soft-delete a user account.
func (p *postgresImpl) UserDelete(clientID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), constants.ThreeSeconds())

	defer cancel()

	rowsAffected, err := p.Query.userDelete(ctx, clientID)
	if err != nil || rowsAffected != int64(1) {
		p.logger.Error("failed to delete user", zap.Error(err))

		return ErrNotFoundUser
	}

	return nil
}

// UserIsDeleted is the interface through which external methods can check if a user account is soft-deleted.
func (p *postgresImpl) UserIsDeleted(clientID uuid.UUID) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.ThreeSeconds())

	defer cancel()

	isDeleted, err := p.Query.userIsDeleted(ctx, clientID)
	if err != nil {
		p.logger.Error("failed to check deletion status of user", zap.Error(err))

		return false, ErrNotFound
	}

	return isDeleted, nil
}
