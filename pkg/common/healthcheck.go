package common

import (
	"errors"
	"net/http"

	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/redis"
	"go.uber.org/zap"
)

// HTTPHealthcheck checks if the service is healthy by pinging the data tier comprised of Postgres and Redis.
func HTTPHealthcheck(db postgres.Postgres, cache redis.Redis, logger *logger.Logger) (int, string, error) {
	// Database health.
	if err := db.Healthcheck(); err != nil {
		msg := "healthcheck failed, Postgres database could not be pinged"
		logger.Warn(msg, zap.Error(err))

		return http.StatusServiceUnavailable, msg, errors.New(msg)
	}

	// Cache health.
	if err := cache.Healthcheck(); err != nil {
		msg := "healthcheck failed, Redis cache could not be pinged"
		logger.Warn(msg, zap.Error(err))

		return http.StatusServiceUnavailable, msg, errors.New(msg)
	}

	return http.StatusOK, "healthy", nil
}
