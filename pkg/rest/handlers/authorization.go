package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/postgres"
	"go.uber.org/zap"
)

// AuthMiddleware is the middleware that checks whether a JWT is valid and can access an endpoint.
func AuthMiddleware(auth auth.Auth, db postgres.Postgres, logger *logger.Logger, authHeaderKey string) gin.HandlerFunc {
	handler := func(context *gin.Context) {
		var (
			err         error
			clientID    uuid.UUID
			expiresAt   int64
			isDeleted   bool
			tokenString = context.GetHeader(authHeaderKey)
		)

		if tokenString == "" {
			context.JSON(http.StatusUnauthorized, "request does not contain an access token")
			context.Abort()

			return
		}

		if clientID, expiresAt, err = auth.ValidateJWT(tokenString); err != nil {
			context.JSON(http.StatusForbidden, "request contains invalid or expired authorization token")
			context.Abort()

			return
		}

		// Check for user deleted status.
		if isDeleted, err = db.UserIsDeleted(clientID); err != nil {
			logger.Error("unable to retrieve client account status", zap.Error(err))

			context.JSON(http.StatusInternalServerError, constants.RetryMessageString())
			context.Abort()

			return
		}

		if isDeleted {
			context.JSON(http.StatusForbidden, "request contains invalid or expired authorization token")
			context.Abort()

			return
		}

		// Store the extracted ClientID and expiration deadline values in the Gin context for handlers.
		context.Set(constants.ClientIDCtxKey(), clientID)
		context.Set(constants.ExpiresAtCtxKey(), expiresAt)

		context.Next()
	}

	return handler
}
