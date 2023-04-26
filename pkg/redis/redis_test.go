package redis

import (
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
	"gopkg.in/yaml.v3"
)

func TestNewRedisImpl(t *testing.T) {
	testCases := []struct {
		name      string
		fileName  string
		input     string
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
	}{
		{
			name:      "File found",
			fileName:  constants.GetRedisFileName(),
			input:     redisConfigTestData[configFileKey],
			expectErr: require.NoError,
			expectNil: require.NotNil,
		}, {
			name:      "File not found",
			fileName:  "wrong_file_name.yaml",
			input:     redisConfigTestData[configFileKey],
			expectErr: require.Error,
			expectNil: require.Nil,
		},
	}
	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			// Configure mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "failed to create in memory directory.")
			require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+test.fileName, []byte(test.input), 0644),
				"failed to write in memory file.")

			c, err := newRedisImpl(&fs, zapLogger)
			test.expectErr(t, err)
			test.expectNil(t, c)
		})
	}
}

func TestNewRedis(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "failed to create in memory directory.")
	require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetRedisFileName(),
		[]byte(redisConfigTestData[configFileKey]), 0644), "failed to write in memory file.")

	testCases := []struct {
		name      string
		fs        *afero.Fs
		log       *logger.Logger
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
	}{
		{
			name:      "Invalid file system and logger",
			fs:        nil,
			log:       nil,
			expectErr: require.Error,
			expectNil: require.Nil,
		}, {
			name:      "Invalid file system",
			fs:        nil,
			log:       zapLogger,
			expectErr: require.Error,
			expectNil: require.Nil,
		}, {
			name:      "Invalid logger",
			fs:        &fs,
			log:       nil,
			expectErr: require.Error,
			expectNil: require.Nil,
		}, {
			name:      "Valid",
			fs:        &fs,
			log:       zapLogger,
			expectErr: require.NoError,
			expectNil: require.NotNil,
		},
	}
	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			redisDB, err := NewRedis(test.fs, test.log)
			test.expectErr(t, err, "error expectation failed.")
			test.expectNil(t, redisDB, "nil expectation for Redis client failed.")
		})
	}
}

func TestVerifySession(t *testing.T) {
	nilConnection := redisImpl{redisDB: nil}
	require.Error(t, nilConnection.verifySession(), "nil connection should return an error.")

	badConnection := redisImpl{redisDB: redis.NewClient(&redis.Options{Addr: invalidServerAddr})}
	require.Error(t, badConnection.verifySession(), "verifying a not open connection should return an error.")
}

func TestRedisImpl_Open(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Ping failure.
	badCfg := &config{}
	badCfg.Connection.Addr = invalidServerAddr
	badCfg.Connection.MaxConnAttempts = 1
	noNodes := redisImpl{conf: badCfg, logger: zapLogger}
	err := noNodes.Open()
	require.Error(t, err, "connection should fail to ping the Redis server.")
	require.Contains(t, err.Error(), "connection refused", "error should contain information connection refused.")

	// Connection success.
	conf := config{}
	require.NoError(t, yaml.Unmarshal([]byte(redisConfigTestData[configFileKey]), &conf), "failed to prepare test config.")

	testRedis := redisImpl{conf: &conf, logger: zapLogger}
	require.NoError(t, testRedis.Open(), "failed to create new Redis server connection.")

	// Leaked connection check.
	require.Error(t, testRedis.Open(), "leaking a connection should raise an error.")
}

func TestRedisImpl_Close(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Ping failure.
	badCfg := &config{}
	badCfg.Connection.Addr = invalidServerAddr
	badCfg.Connection.MaxConnAttempts = 1
	noNodes := redisImpl{conf: badCfg, logger: zapLogger}
	err := noNodes.Close()
	require.Error(t, err, "connection should fail to ping the Redis server.")
	require.Contains(t, err.Error(), "no session", "error should contain information on no session.")

	// Connection success.
	conf := config{}
	require.NoError(t, yaml.Unmarshal([]byte(redisConfigTestData[configFileKey]), &conf), "failed to prepare test config.")

	testRedis := redisImpl{conf: &conf, logger: zapLogger}
	require.NoError(t, testRedis.Open(), "failed to open Redis server connection for test.")
	require.NoError(t, testRedis.Close(), "failed to close Redis server connection.")

	// Leaked connection check.
	require.Error(t, testRedis.Close(), "closing a closed Redis client connection should raise an error.")
}

func TestRedisImpl_Healthcheck(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Open unhealthy connection, ignore error, and run check.
	unhealthyConf := config{}
	require.NoError(t, yaml.Unmarshal([]byte(redisConfigTestData[configFileKey]), &unhealthyConf),
		"failed to prepare unhealthy config")

	unhealthyConf.Connection.Addr = invalidServerAddr
	unhealthy := redisImpl{conf: &unhealthyConf, logger: zapLogger}
	require.Error(t, unhealthy.Open(), "opening a connection to bad endpoints should fail")
	err := unhealthy.Healthcheck()
	require.Error(t, err, "unhealthy healthcheck failed")
	require.Contains(t, err.Error(), "connection refused", "error is not about a bad connection")

	// Open healthy connection, ignore error, and run check.
	healthyConf := config{}
	require.NoError(t, yaml.Unmarshal([]byte(redisConfigTestData[configFileKey]), &healthyConf),
		"failed to prepare healthy config")

	healthy := redisImpl{conf: &healthyConf, logger: zapLogger}
	require.NoError(t, healthy.Open(), "opening a connection to good endpoints should not fail")
	err = healthy.Healthcheck()
	require.NoError(t, err, "healthy healthcheck failed")
}

func TestRedisImpl_Set_Get_Del(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	testCases := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "First test key",
			key:   xid.New().String(),
			value: xid.New().String(),
		}, {
			name:  "Second test key",
			key:   xid.New().String(),
			value: xid.New().String(),
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			// Write to Redis.
			require.NoError(t, connection.Set(test.key, test.value, time.Duration(2)), "failed to write to Redis")

			// Get data and validate it.
			retrieved := ""
			err := connection.Get(test.key, &retrieved)
			require.NoError(t, err, "failed to retrieve data from Redis")
			require.Equal(t, retrieved, test.value, "retrieved value does not match expected")

			// Remove data from Redis server.
			require.NoError(t, connection.Del(test.key), "failed to remove key from Redis server")

			// Check to see if data has been removed.
			var deleted *string
			err = connection.Get(test.key, deleted)
			require.Nil(t, deleted, "returned data from a deleted record should be nil")
			require.Error(t, err, "deleted record should not be found on redis Redis server")

			// Double-delete data.
			require.Error(t, connection.Del(test.key), "removing a nonexistent key from Redis server should fail")
		})
	}

	// Sleep to allow cache expiration.
	time.Sleep(2 * time.Second)

	for _, testCase := range testCases {
		test := testCase

		t.Run(fmt.Sprintf("Expiration check: %s", test.name), func(t *testing.T) {
			var deleted *string
			err := connection.Get(test.key, deleted)
			require.Nil(t, deleted, "returned data from a deleted record should be nil")
			require.Error(t, err, "deleted record should not be found on redis Redis server")
		})
	}
}
