package main

import (
	"log"
	"sync"

	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/redis"
	"github.com/surahman/FTeX/pkg/rest"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

func main() {
	var (
		err           error
		serverREST    *rest.Server
		logging       *logger.Logger
		authorization auth.Auth
		database      postgres.Postgres
		cache         redis.Redis
		waitGroup     sync.WaitGroup
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
		logging.Panic("failed to configure Cassandra module", zap.Error(err))
	}

	if err = database.Open(); err != nil {
		logging.Panic("failed open a connection to the Cassandra cluster", zap.Error(err))
	}
	defer func(database postgres.Postgres) {
		if err = database.Close(); err != nil {
			logging.Panic("failed close the connection to the Cassandra cluster", zap.Error(err))
		}
	}(database)

	// Cache setup
	if cache, err = redis.NewRedis(&fs, logging); err != nil {
		logging.Panic("failed to configure Redis module", zap.Error(err))
	}

	if err = cache.Open(); err != nil {
		logging.Panic("failed open a connection to the Redis cluster", zap.Error(err))
	}
	defer func(cache redis.Redis) {
		if err = cache.Close(); err != nil {
			logging.Panic("failed close the connection to the Redis cluster", zap.Error(err))
		}
	}(cache)

	// Authorization setup.
	if authorization, err = auth.NewAuth(&fs, logging); err != nil {
		logging.Panic("failed to configure authorization module", zap.Error(err))
	}

	// Setup REST server and start it.
	waitGroup.Add(1)

	if serverREST, err = rest.NewServer(&fs, authorization, database, cache, logging, &waitGroup); err != nil {
		logging.Panic("failed to create the REST server", zap.Error(err))
	}

	go serverREST.Run()

	waitGroup.Wait()
}
