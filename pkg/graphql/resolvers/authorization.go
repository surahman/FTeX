package graphql

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/logger"
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
func AuthorizationCheck(ctx context.Context, auth auth.Auth, logger *logger.Logger, authHeaderKey string) (
	uuid.UUID, int64, error) {
	var (
		clientID   uuid.UUID
		expiresAt  int64
		err        error
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

	return clientID, expiresAt, nil
}
