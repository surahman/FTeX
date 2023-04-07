package redis

import (
	"errors"
	"strconv"
	"testing"

	"github.com/rs/xid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/validator"
)

func TestRedisConfigs_Load(t *testing.T) {
	envAuthKey := constants.GetRedisPrefix() + "_AUTHENTICATION."
	envConnKey := constants.GetRedisPrefix() + "_CONNECTION."
	envDataKey := constants.GetRedisPrefix() + "_DATA."

	testCases := []struct {
		name         string
		input        string
		envValue     string
		expectErrCnt int
		expectErr    require.ErrorAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			name:         "empty - etc dir",
			input:        redisConfigTestData["empty"],
			envValue:     xid.New().String(),
			expectErrCnt: 6,
			expectErr:    require.Error,
		}, {
			name:         "valid - etc dir",
			input:        redisConfigTestData["valid"],
			envValue:     xid.New().String(),
			expectErrCnt: 0,
			expectErr:    require.NoError,
		}, {
			name:         "github-ci-runner - etc dir",
			input:        redisConfigTestData["github-ci-runner"],
			envValue:     xid.New().String(),
			expectErrCnt: 0,
			expectErr:    require.NoError,
		}, {
			name:         "no username - etc dir",
			input:        redisConfigTestData["username_empty"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
		}, {
			name:         "no password - etc dir",
			input:        redisConfigTestData["password_empty"],
			envValue:     xid.New().String(),
			expectErrCnt: 0,
			expectErr:    require.NoError,
		}, {
			name:         "no addrs - etc dir",
			input:        redisConfigTestData["no_addr"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
		}, {
			name:         "invalid max retries - etc dir",
			input:        redisConfigTestData["invalid_max_retries"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
		}, {
			name:         "invalid pool size - etc dir",
			input:        redisConfigTestData["invalid_pool_size"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
		}, {
			name:         "invalid min idle conns - etc dir",
			input:        redisConfigTestData["invalid_min_idle_conns"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
		}, {
			name:         "no max idle conns - etc dir",
			input:        redisConfigTestData["no_max_idle_conns"],
			envValue:     xid.New().String(),
			expectErrCnt: 0,
			expectErr:    require.NoError,
		}, {
			name:         "invalid min TTL - etc dir",
			input:        redisConfigTestData["invalid_min_ttl"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Configure mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetRedisFileName(),
				[]byte(testCase.input), 0644), "Failed to write in memory file")

			// Load from mock filesystem.
			actual := &config{}
			err := actual.Load(fs)
			testCase.expectErr(t, err)

			validationError := &validator.ValidationError{}
			if errors.As(err, &validationError) {
				require.Equalf(t, testCase.expectErrCnt, len(validationError.Errors),
					"expected errors count is incorrect: %v", err)

				return
			}

			// Test configuring of environment variable.
			username := xid.New().String()
			password := xid.New().String()
			t.Setenv(envAuthKey+"USERNAME", username)
			t.Setenv(envAuthKey+"PASSWORD", password)

			addr := xid.New().String()
			maxConnAttempts := 12
			maxRetries := 55
			poolSize := 164
			minIdleConns := 9
			maxIdleConns := 101
			t.Setenv(envConnKey+"ADDR", addr)
			t.Setenv(envConnKey+"MAXCONNATTEMPTS", strconv.Itoa(maxConnAttempts))
			t.Setenv(envConnKey+"MAXRETRIES", strconv.Itoa(maxRetries))
			t.Setenv(envConnKey+"POOLSIZE", strconv.Itoa(poolSize))
			t.Setenv(envConnKey+"MINIDLECONNS", strconv.Itoa(minIdleConns))
			t.Setenv(envConnKey+"MAXIDLECONNS", strconv.Itoa(maxIdleConns))

			ttl := 999
			t.Setenv(envDataKey+"TTL", strconv.Itoa(ttl))

			require.NoErrorf(t, actual.Load(fs), "failed to load constants file: %v", err)

			require.Equal(t, username, actual.Authentication.Username, "failed to load username.")
			require.Equal(t, password, actual.Authentication.Password, "failed to load password.")

			require.Equal(t, addr, actual.Connection.Addr, "failed to load address.")
			require.Equal(t, maxRetries, actual.Connection.MaxRetries, "failed to load max retries.")
			require.Equal(t, maxConnAttempts, actual.Connection.MaxConnAttempts,
				"failed to load max connection attempts.")
			require.Equal(t, poolSize, actual.Connection.PoolSize, "failed to load pool size.")
			require.Equal(t, minIdleConns, actual.Connection.MinIdleConns, "failed to load min idle conns.")
			require.Equal(t, maxIdleConns, actual.Connection.MaxIdleConns, "failed to load max idle conns.")

			require.Equal(t, int64(ttl), actual.Data.TTL, "failed to load ttl.")
		})
	}
}
