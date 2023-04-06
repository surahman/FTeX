package redis

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/logger"
	"go.uber.org/zap"
)

// Mock Redis interface stub generation.
//go:generate mockgen -destination=../mocks/mock_redis.go -package=mocks github.com/surahman/mcq-platform/pkg/redis Redis

// Redis is the interface through which the cluster can be accessed. Created to support mock testing.
type Redis interface {
	// Open will create a connection pool and establish a connection to the cache cluster.
	Open() error

	//// Close will shut down the connection pool and ensure that the connection to the cache cluster is terminated correctly.
	//Close() error
	//
	//// Healthcheck will ping all the nodes in the cluster to see if all the shards are reachable.
	//Healthcheck() error
	//
	//// Set will place a key with a given value in the cluster with a TTL, if specified in the configurations.
	//Set(string, any) error
	//
	//// Get will retrieve a value associated with a provided key.
	//Get(string, any) error
	//
	//// Del will remove all keys provided as a set of keys.
	//Del(...string) error
}

// Check to ensure the Redis interface has been implemented.
var _ Redis = &redisImpl{}

// redisImpl implements the Redis interface and contains the logic to interface with the cluster.
type redisImpl struct {
	conf    *config
	redisDB *redis.Client
	logger  *logger.Logger
}

// NewRedis will create a new Redis configuration by loading it.
func NewRedis(fs *afero.Fs, logger *logger.Logger) (Redis, error) {
	if fs == nil || logger == nil {
		return nil, errors.New("nil file system or logger supplied")
	}

	return newRedisImpl(fs, logger)
}

// newRedisImpl will create a new redisImpl configuration and load it from disk.
func newRedisImpl(fs *afero.Fs, logger *logger.Logger) (c *redisImpl, err error) {
	c = &redisImpl{conf: newConfig(), logger: logger}
	if err = c.conf.Load(*fs); err != nil {
		c.logger.Error("failed to load Redis configurations from disk", zap.Error(err))

		return nil, err
	}

	return
}

// verifySession will check to see if a session is established.
func (r *redisImpl) verifySession() error {
	if r.redisDB == nil || r.redisDB.Ping(context.Background()).Err() != nil {
		return errors.New("no session established")
	}

	return nil
}

// createSessionRetry will attempt to open the connection using binary exponential back-off and stop on the first
// success or fail after the last one.
func (r *redisImpl) createSessionRetry() error {
	var err error

	for attempt := 1; attempt <= r.conf.Connection.MaxConnAttempts; attempt++ {
		waitTime := time.Duration(math.Pow(2, float64(attempt))) * time.Second
		r.logger.Info(fmt.Sprintf("Attempting connection to Redis server in %s...", waitTime),
			zap.String("attempt", strconv.Itoa(attempt)))

		time.Sleep(waitTime)

		if err = r.redisDB.Ping(context.Background()).Err(); err == nil {
			return nil
		}
	}
	r.logger.Error("unable to establish connection to Redis cluster", zap.Error(err))

	return nil
}

// Open will establish a connection to the Redis cache cluster.
func (r *redisImpl) Open() error {
	// Stop connection leaks.
	if err := r.verifySession(); err == nil {
		msg := "session to Redis server is already established"
		r.logger.Warn(msg)

		return errors.New(msg)
	}

	// Compile Redis connection configurations.
	redisConfig := &redis.Options{
		Addr:         r.conf.Connection.Addr,
		Username:     r.conf.Authentication.Username,
		Password:     r.conf.Authentication.Password,
		MaxRetries:   r.conf.Connection.MaxRetries,
		PoolSize:     r.conf.Connection.PoolSize,
		MinIdleConns: r.conf.Connection.MinIdleConns,
	}

	if r.conf.Connection.MaxIdleConns > 0 {
		redisConfig.MaxIdleConns = r.conf.Connection.MaxIdleConns
	}

	r.redisDB = redis.NewClient(redisConfig)

	return r.createSessionRetry()
}
