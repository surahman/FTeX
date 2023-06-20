package main

import (
	"log"
	"sync"

	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/graphql"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/quotes"
	"github.com/surahman/FTeX/pkg/redis"
	"github.com/surahman/FTeX/pkg/rest"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

// callbacks is an array of callback functions that will be iterated over and called.
type callbacks []func() error

// add will append callback functions to the callback function array.
func (c *callbacks) add(callBackFunc func() error) {
	*c = append(*c, callBackFunc)
}

// callback will iterate over the array of callback functions, executing each one, and log any failures.
func (c *callbacks) callback(logger *logger.Logger) {
	for _, callbackFunc := range *c {
		if err := callbackFunc(); err != nil {
			logger.Error("failed to successfully execute callback", zap.Error(err))
		}
	}
}

func main() {
	var (
		authorization   auth.Auth
		cache           redis.Redis
		cleanup         callbacks
		database        postgres.Postgres
		err             error
		logging         *logger.Logger
		conversionRates quotes.Quotes
		serverGraphQL   *graphql.Server
		serverREST      *rest.Server
		waitGroup       sync.WaitGroup
	)

	// File system setup.
	fs := afero.NewOsFs()

	// Logger setup.
	logging = logger.NewLogger()
	if err = logging.Init(&fs); err != nil {
		log.Fatalf("failed to initialize logger module: %v", err)
	}

	// Postgres setup.
	if database, err = postgres.NewPostgres(&fs, logging); err != nil {
		cleanup.callback(logging)
		logging.Panic("failed to configure Postgres module", zap.Error(err))
	}

	if err = database.Open(); err != nil {
		cleanup.callback(logging)
		logging.Panic("failed open a connection to the Postgres database", zap.Error(err))
	}

	cleanup.add(database.Close)

	// Cache setup
	if cache, err = redis.NewRedis(&fs, logging); err != nil {
		cleanup.callback(logging)
		logging.Panic("failed to configure Redis module", zap.Error(err))
	}

	if err = cache.Open(); err != nil {
		cleanup.callback(logging)
		logging.Panic("failed open a connection to the Redis cluster", zap.Error(err))
	}

	cleanup.add(cache.Close)

	// Quotes setup.
	if conversionRates, err = quotes.NewQuote(&fs, logging); err != nil {
		cleanup.callback(logging)
		logging.Panic("failed to configure Quotes module", zap.Error(err))
	}

	// Authorization setup.
	if authorization, err = auth.NewAuth(&fs, logging); err != nil {
		cleanup.callback(logging)
		logging.Panic("failed to configure authorization module", zap.Error(err))
	}

	// Setup is completed. Configure the cleanup callbacks to be executed on shutdown/exit.
	defer cleanup.callback(logging)

	// Setup REST server and start it.
	waitGroup.Add(1)

	if serverREST, err = rest.
		NewServer(&fs, authorization, database, cache, conversionRates, logging, &waitGroup); err != nil {
		logging.Panic("failed to create the REST server", zap.Error(err))
	}

	go serverREST.Run()

	// Setup GraphQL server and start it.
	waitGroup.Add(1)

	if serverGraphQL, err = graphql.
		NewServer(&fs, authorization, database, cache, conversionRates, logging, &waitGroup); err != nil {
		logging.Panic("failed to create the GraphQL server", zap.Error(err))
	}

	go serverGraphQL.Run()

	waitGroup.Wait()
}
