package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/constants"
)

// AuthMiddleware is the middleware that checks whether a JWT is valid and can access an endpoint.
func AuthMiddleware(auth auth.Auth, authHeaderKey string) gin.HandlerFunc {
	handler := func(context *gin.Context) {
		var (
			err         error
			clientID    uuid.UUID
			expiresAt   int64
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

		// Store the extracted ClientID and expiration deadline values in the Gin context for handlers.
		context.Set(constants.ClientIDCtxKey(), clientID)
		context.Set(constants.ExpiresAtCtxKey(), expiresAt)

		context.Next()
	}

	return handler
}
