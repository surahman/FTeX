package rest

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/redis"
)

// Format and generate Swagger UI files using makefile.
//go:generate make -C ../../ swagger

// Server is the HTTP REST server.
type Server struct {
	auth   auth.Auth
	cache  redis.Redis
	db     postgres.Postgres
	conf   *config
	logger *logger.Logger
	router *gin.Engine
	wg     *sync.WaitGroup
}

// NewServer will create a new REST server instance in a non-running state.
func NewServer(fs *afero.Fs, auth auth.Auth, postgres postgres.Postgres, redis redis.Redis,
	logger *logger.Logger, wg *sync.WaitGroup) (server *Server, err error) {
	// Load configurations.
	conf := newConfig()
	if err = conf.Load(*fs); err != nil {
		return
	}

	return &Server{
			conf:   conf,
			auth:   auth,
			cache:  redis,
			db:     postgres,
			logger: logger,
			wg:     wg,
		},
		err
}
