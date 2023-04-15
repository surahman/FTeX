package rest

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/redis"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

// initialize will configure the HTTP server routes.
func (s *Server) initialize() {
	s.router = gin.Default()

	//	@title						FTeX, Incorporated. (Formerly Crypto-Bro's Bank, Inc.)
	//	@version					1.0.0
	//	@description				FTeX Fiat and Cryptocurrency Banking API.
	//	@description				Bank, buy, and sell Fiat and Cryptocurrencies. Prices for all currencies are
	//	@description				retrieved from real-time quote providers.
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
}
