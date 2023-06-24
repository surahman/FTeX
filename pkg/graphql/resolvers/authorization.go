package graphql

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/postgres"
	"go.uber.org/zap"
)

type GinContextKey struct{}

// GinContextFromContext will extract the Gin context from the context passed in.
func GinContextFromContext(ctx context.Context, logger *logger.Logger) (*gin.Context, error) {
	ctxValue := ctx.Value(GinContextKey{})
	if ctxValue == nil {
		logger.Error("could not retrieve gin.Context")

		return nil, errors.New("malformed request: authorization information not found")
	}

	ginContext, ok := ctxValue.(*gin.Context)
	if !ok {
		logger.Error("gin.Context has wrong type")

		return nil, errors.New("malformed request: authorization information malformed")
	}

	return ginContext, nil
}

// AuthorizationCheck will validate the JWT payload for valid authorization information.
func AuthorizationCheck(ctx context.Context, auth auth.Auth, db postgres.Postgres, logger *logger.Logger,
	authHeaderKey string) (uuid.UUID, int64, error) {
	var (
		clientID   uuid.UUID
		expiresAt  int64
		err        error
		isDeleted  bool
		ginContext *gin.Context
	)

	if ginContext, err = GinContextFromContext(ctx, logger); err != nil {
		return clientID, -1, err
	}

	tokenString := ginContext.GetHeader(authHeaderKey)
	if tokenString == "" {
		return clientID, -1, errors.New("request does not contain an access token")
	}

	if clientID, expiresAt, err = auth.ValidateJWT(tokenString); err != nil {
		return clientID, expiresAt, fmt.Errorf("failed to validate JWT %w", err)
	}

	// Check for user deleted status.
	if isDeleted, err = db.UserIsDeleted(clientID); err != nil {
		logger.Error("unable to retrieve client account status", zap.Error(err))

		return clientID, expiresAt, errors.New(constants.RetryMessageString())
	}

	if isDeleted {
		return clientID, expiresAt, errors.New("request contains invalid or expired authorization token")
	}

	return clientID, expiresAt, nil
}
