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

	// Close will shut down the connection pool and ensure that the connection to the cache cluster is terminated
	// correctly.
	Close() error

	// Healthcheck will ping all the nodes in the cluster to see if all the shards are reachable.
	Healthcheck() error

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

		// Successfully opened lazy connection with a ping.
		if err = r.redisDB.Ping(context.Background()).Err(); err == nil {
			return nil
		}
	}

	// Unable to ping Redis server and establish lazy connection.
	msg := "unable to establish connection to Redis cluster"
	r.logger.Error(msg, zap.Error(err))

	return fmt.Errorf(msg+" %w", err)
}

// Open will establish a connection to the Redis cache cluster.
func (r *redisImpl) Open() error {
	// Stop connection leaks.
	if err := r.verifySession(); err == nil {
		msg := "session to Redis server is already established"
		r.logger.Warn(msg)

		return fmt.Errorf(msg+" %w", err)
	}

	// Compile Redis connection configurations.
	redisConfig := &redis.Options{
		Addr:                  r.conf.Connection.Addr,
		Username:              r.conf.Authentication.Username,
		Password:              r.conf.Authentication.Password,
		MaxRetries:            r.conf.Connection.MaxRetries,
		PoolSize:              r.conf.Connection.PoolSize,
		MinIdleConns:          r.conf.Connection.MinIdleConns,
		ContextTimeoutEnabled: true,
	}

	if r.conf.Connection.MaxIdleConns > 0 {
		redisConfig.MaxIdleConns = r.conf.Connection.MaxIdleConns
	}

	r.redisDB = redis.NewClient(redisConfig)

	return r.createSessionRetry()
}

// Close will terminate a connection to the Redis cache cluster.
func (r *redisImpl) Close() error {
	var err error

	// Check for an open connection.
	if err = r.verifySession(); err != nil {
		msg := "no session to Redis server established to close"
		r.logger.Warn(msg)

		return fmt.Errorf(msg+" %w", err)
	}

	if err = r.redisDB.Close(); err != nil {
		msg := "failed to close Redis server connection"
		r.logger.Warn(msg)

		return fmt.Errorf(msg+" %w", err)
	}

	return nil
}

// Healthcheck will iterate through all the data shards and attempt to ping them to ensure they are all reachable.
func (r *redisImpl) Healthcheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:gomnd

	defer cancel()

	if err := r.redisDB.Ping(ctx).Err(); err != nil {
		msg := "redis health check ping failed"
		r.logger.Info(msg)

		return fmt.Errorf(msg+" %w", err)
	}

	return nil
}
