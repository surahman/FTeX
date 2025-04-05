package graphql

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/surahman/FTeX/pkg/auth"
	graphql_generated "github.com/surahman/FTeX/pkg/graphql/generated"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/quotes"
	"github.com/surahman/FTeX/pkg/redis"
)

// QueryHandler is the endpoint through which GraphQL can be accessed.
func QueryHandler(authHeaderKey string, auth auth.Auth, cache redis.Redis, db postgres.Postgres,
	quotes quotes.Quotes, logger *logger.Logger) gin.HandlerFunc {
	gqlHandler := handler.New(graphql_generated.NewExecutableSchema(
		graphql_generated.Config{
			Resolvers: &Resolver{
				authHeaderKey: authHeaderKey,
				auth:          auth,
				cache:         cache,
				db:            db,
				quotes:        quotes,
				logger:        logger,
			},
		},
	))
	gqlHandler.AddTransport(transport.POST{})

	return func(c *gin.Context) {
		gqlHandler.ServeHTTP(c.Writer, c.Request)
	}
}

// PlaygroundHandler is the endpoint through which the GraphQL playground can be accessed.
func PlaygroundHandler(baseURL, queryURL string) gin.HandlerFunc {
	h := playground.Handler("GraphQL", fmt.Sprintf("/%s%s", baseURL, queryURL))

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// GinContextToContextMiddleware is middleware that will place the Gin context into a context for the GraphQL resolvers.
func GinContextToContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), GinContextKey{}, c)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
