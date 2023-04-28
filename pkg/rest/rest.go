package rest

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
	_ "github.com/surahman/FTeX/docs" // Swaggo generated Swagger documentation
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/quotes"
	"github.com/surahman/FTeX/pkg/redis"
	restHandlers "github.com/surahman/FTeX/pkg/rest/handlers"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// Format and generate Swagger UI files using makefile.
//go:generate make -C ../../ swagger

// Server is the HTTP REST server.
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

// NewServer will create a new REST server instance in a non-running state.
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

	//	@title						FTeX, Inc. (Formerly Crypto-Bro's Bank, Inc.)
	//	@version					1.0.0
	//	@description				FTeX Fiat and Cryptocurrency Banking API.
	//	@description				Bank, buy, and sell Fiat and Cryptocurrencies. Prices for all currencies are retrieved from real-time quote providers.
	//
	//	@schemes					http
	//	@host						localhost:33723
	//	@BasePath					/api/rest/v1
	//
	//	@accept						json
	//	@produce					json
	//
	//	@contact.name				Saad Ur Rahman
	//	@contact.url				https://www.linkedin.com/in/saad-ur-rahman/
	//	@contact.email				saad.ur.rahman@gmail.com
	//
	//	@license.name				GPL-3.0
	//	@license.url				https://opensource.org/licenses/GPL-3.0
	//
	//	@securityDefinitions.apikey	ApiKeyAuth
	//	@in							header
	//	@name						Authorization

	s.router.GET(s.conf.Server.SwaggerPath, ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Endpoint configurations
	authMiddleware := restHandlers.AuthMiddleware(s.auth, s.conf.Authorization.HeaderKey)
	api := s.router.Group(s.conf.Server.BasePath)

	api.GET("/health", restHandlers.Healthcheck(s.logger, s.db, s.cache))

	userGroup := api.Group("/user")
	userGroup.POST("/register", restHandlers.RegisterUser(s.logger, s.auth, s.db))
	userGroup.POST("/login", restHandlers.LoginUser(s.logger, s.auth, s.db))
	userGroup.
		Use(authMiddleware).
		POST("/refresh", restHandlers.LoginRefresh(s.logger, s.auth, s.db, s.conf.Authorization.HeaderKey))
	userGroup.
		Use(authMiddleware).
		DELETE("/delete", restHandlers.DeleteUser(s.logger, s.auth, s.db, s.conf.Authorization.HeaderKey))

	fiatGroup := api.Group("/fiat")
	fiatGroup.
		Use(authMiddleware).
		POST("/open", restHandlers.OpenFiat(s.logger, s.auth, s.db, s.conf.Authorization.HeaderKey))
	fiatGroup.
		Use(authMiddleware).
		POST("/deposit", restHandlers.DepositFiat(s.logger, s.auth, s.db, s.conf.Authorization.HeaderKey))
	fiatGroup.
		Use(authMiddleware).
		POST("/exchange/offer",
			restHandlers.ExchangeOfferFiat(s.logger, s.auth, s.cache, s.quotes, s.conf.Authorization.HeaderKey))
}

// Run brings the HTTP service up.
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
