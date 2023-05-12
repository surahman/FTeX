package graphql

import (
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/quotes"
	"github.com/surahman/FTeX/pkg/redis"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	authHeaderKey string
	auth          auth.Auth
	cache         redis.Redis
	db            postgres.Postgres
	quotes        quotes.Quotes
	logger        *logger.Logger
}
