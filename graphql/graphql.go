package graphql

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/quotes"
	"github.com/surahman/FTeX/pkg/redis"
	"go.uber.org/zap"
)

// Generate the GraphQL Go code for resolvers.
//go:generate go run github.com/99designs/gqlgen generate

// Server is the HTTP GraphQL server.
type Server struct {
	auth   auth.Auth
	cache  redis.Redis
	db     postgres.Postgres
	quotes quotes.Quotes
	conf   *config
	logger *logger.Logger
	router *gin.Engine
	wg     *sync.WaitGroup
}

// NewServer will create a new GraphQL server instance in a non-running state.
func NewServer(fs *afero.Fs, auth auth.Auth, postgres postgres.Postgres, redis redis.Redis, quotes quotes.Quotes,
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
			quotes: quotes,
			logger: logger,
			wg:     wg,
		},
		err
}

// initialize will configure the HTTP server routes.
func (s *Server) initialize() {
	s.router = gin.Default()

	// Endpoint configurations
	api := s.router.Group(s.conf.Server.BasePath)
	_ = api //  REMOVE.

	//	api.Use(graphql_resolvers.GinContextToContextMiddleware())
	//
	//	api.POST(s.conf.Server.QueryPath, graphql_resolvers.QueryHandler(s.conf.Authorization.HeaderKey, s.auth, s.cache, s.db, s.grading, s.logger))
	//	api.GET(s.conf.Server.PlaygroundPath, graphql_resolvers.PlaygroundHandler(s.conf.Server.BasePath, s.conf.Server.QueryPath))
}

// Run brings the HTTP GraphQL service up.
func (s *Server) Run() {
	// Indicate to bootstrapping thread to wait for completion.
	defer s.wg.Done()

	// Configure routes.
	s.initialize()

	// Create server.
	srv := &http.Server{
		ReadTimeout:       s.conf.Server.ReadTimeout,
		WriteTimeout:      s.conf.Server.WriteTimeout,
		ReadHeaderTimeout: s.conf.Server.ReadHeaderTimeout,
		Addr:              fmt.Sprintf(":%d", s.conf.Server.PortNumber),
		Handler:           s.router,
	}

	// Error channel for failed server start.
	serverErr := make(chan error, 1)

	// Wait for interrupt signal to gracefully shut down the server.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP listener.
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Check for server start failure or shutdown signal.
	select {
	case err := <-serverErr:
		s.logger.Error(fmt.Sprintf("REST server failed to listen on port %d", s.conf.Server.PortNumber), zap.Error(err))

		return
	case <-quit:
		s.logger.Info("Shutting down REST server...", zap.Duration("waiting", s.conf.Server.ShutdownDelay))
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.conf.Server.ShutdownDelay)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		s.logger.Panic("Failed to shutdown REST server", zap.Error(err))
	}

	// 5 second wait to exit.
	<-ctx.Done()

	s.logger.Info("REST server exited")
}
