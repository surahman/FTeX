package redis

import (
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
	"gopkg.in/yaml.v3"
)

func TestNewRedisImpl(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		fileName  string
		input     string
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			name:      "File found",
			fileName:  constants.GetRedisFileName(),
			input:     redisConfigTestData["test_suite"],
			expectErr: require.NoError,
			expectNil: require.NotNil,
		}, {
			name:      "File not found",
			fileName:  "wrong_file_name.yaml",
			input:     redisConfigTestData["test_suite"],
			expectErr: require.Error,
			expectNil: require.Nil,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

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
		[]byte(redisConfigTestData["test_suite"]), 0644), "failed to write in memory file.")

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

	badConnection := redisImpl{redisDB: redis.NewClient(&redis.Options{})}
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
	require.NoError(t, yaml.Unmarshal([]byte(redisConfigTestData["test_suite"]), &conf), "failed to prepare test config.")

	testRedis := redisImpl{conf: &conf, logger: zapLogger}
	require.NoError(t, testRedis.Open(), "failed to create new cluster connection.")

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
	require.NoError(t, yaml.Unmarshal([]byte(redisConfigTestData["test_suite"]), &conf), "failed to prepare test config.")

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
	require.NoError(t, yaml.Unmarshal([]byte(redisConfigTestData["test_suite"]), &unhealthyConf),
		"failed to prepare unhealthy config")

	unhealthyConf.Connection.Addr = invalidServerAddr
	unhealthy := redisImpl{conf: &unhealthyConf, logger: zapLogger}
	require.Error(t, unhealthy.Open(), "opening a connection to bad endpoints should fail")
	err := unhealthy.Healthcheck()
	require.Error(t, err, "unhealthy healthcheck failed")
	require.Contains(t, err.Error(), "connection refused", "error is not about a bad connection")

	// Open healthy connection, ignore error, and run check.
	healthyConf := config{}
	require.NoError(t, yaml.Unmarshal([]byte(redisConfigTestData["test_suite"]), &healthyConf),
		"failed to prepare healthy config")

	healthy := redisImpl{conf: &healthyConf, logger: zapLogger}
	require.NoError(t, healthy.Open(), "opening a connection to good endpoints should not fail")
	err = healthy.Healthcheck()
	require.NoError(t, err, "healthy healthcheck failed")
}
