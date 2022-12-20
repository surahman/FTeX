package postgres

import (
	"errors"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/rs/xid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/validator"
	"gopkg.in/yaml.v3"
)

func TestNewConfig(t *testing.T) {
	cfg := newConfig()
	require.Equal(t, reflect.TypeOf(&config{}), reflect.TypeOf(cfg), "new config type mismatch")
}

func TestConfigLoader(t *testing.T) {
	envAuthKey := constants.GetPostgresPrefix() + "_AUTHENTICATION."
	envConnKey := constants.GetPostgresPrefix() + "_CONNECTION."
	envPoolKey := constants.GetPostgresPrefix() + "_POOL."

	testCases := []struct {
		name      string
		input     string
		expectErr require.ErrorAssertionFunc
		expectLen int
	}{
		// ----- test cases start ----- //
		{
			name:      "empty - etc dir",
			input:     postgresConfigTestData["empty"],
			expectErr: require.Error,
			expectLen: 9,
		}, {
			name:      "valid - etc dir",
			input:     postgresConfigTestData["valid"],
			expectErr: require.NoError,
			expectLen: 0,
		}, {
			name:      "valid - true bool",
			input:     postgresConfigTestData["valid_true_bool"],
			expectErr: require.NoError,
			expectLen: 0,
		}, {
			name:      "valid - no bool",
			input:     postgresConfigTestData["valid_prod_no_bool"],
			expectErr: require.NoError,
			expectLen: 0,
		}, {
			name:      "bad health check",
			input:     postgresConfigTestData["bad_health_check"],
			expectErr: require.Error,
			expectLen: 1,
		}, {
			name:      "invalid connections",
			input:     postgresConfigTestData["invalid_conns"],
			expectErr: require.Error,
			expectLen: 2,
		}, {
			name:      "invalid max connection attempts",
			input:     postgresConfigTestData["invalid_max_conn_attempts"],
			expectErr: require.Error,
			expectLen: 1,
		}, {
			name:      "invalid timeout",
			input:     postgresConfigTestData["invalid_timeout"],
			expectErr: require.Error,
			expectLen: 1,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Configure mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetPostgresFileName(), []byte(testCase.input), 0644), "Failed to write in memory file")

			// Load from mock filesystem.
			actual := &config{}
			err := actual.Load(fs)
			testCase.expectErr(t, err)

			validationError := &validator.ValidationError{}
			if errors.As(err, &validationError) {
				require.Equalf(t, testCase.expectLen, len(validationError.Errors), "Expected errors count is incorrect: %v", err)
				return
			}

			// Load expected struct.
			expected := &config{}
			require.NoError(t, yaml.Unmarshal([]byte(testCase.input), expected), "failed to unmarshal expected constants")
			require.Truef(t, reflect.DeepEqual(expected, actual), "configurations loaded from disk do not match, expected %v, actual %v", expected, actual)

			// Test configuring of environment variable.
			username := xid.New().String()
			password := xid.New().String()
			t.Setenv(envAuthKey+"USERNAME", username)
			t.Setenv(envAuthKey+"PASSWORD", password)

			database := xid.New().String()
			host := xid.New().String()
			port := 5555
			timeout := 47 * time.Second
			max_conn_attempts := 9
			t.Setenv(envConnKey+"DATABASE", database)
			t.Setenv(envConnKey+"HOST", host)
			t.Setenv(envConnKey+"MAX_CONNECTION_ATTEMPTS", strconv.Itoa(max_conn_attempts))
			t.Setenv(envConnKey+"PORT", strconv.Itoa(port))
			t.Setenv(envConnKey+"TIMEOUT", timeout.String())
			t.Setenv(envConnKey+"SSL_ENABLED", strconv.FormatBool(true))

			health_check_period := 13 * time.Second
			max_conns := 60
			min_conns := 40
			t.Setenv(envPoolKey+"HEALTH_CHECK_PERIOD", health_check_period.String())
			t.Setenv(envPoolKey+"MAX_CONNS", strconv.Itoa(max_conns))
			t.Setenv(envPoolKey+"MIN_CONNS", strconv.Itoa(min_conns))
			t.Setenv(envPoolKey+"LAZY_CONNECT", strconv.FormatBool(true))

			require.NoErrorf(t, actual.Load(fs), "Failed to load constants file: %v", err)

			require.Equal(t, username, actual.Authentication.Username, "failed to load username")
			require.Equal(t, password, actual.Authentication.Password, "failed to load password")

			require.Equal(t, database, actual.Connection.Database, "failed to load database")
			require.Equal(t, host, actual.Connection.Host, "failed to load host")
			require.Equal(t, max_conn_attempts, actual.Connection.MaxConnAttempts, "failed to max connection attempts")
			require.Equal(t, uint16(port), actual.Connection.Port, "failed to load port")
			require.True(t, actual.Connection.SslEnabled, "failed to load ssl enabled")
			require.Equal(t, timeout, actual.Connection.Timeout, "failed to load timeout")

			require.Equal(t, health_check_period, actual.Pool.HealthCheckPeriod, "failed to load duration")
			require.Equal(t, int32(max_conns), actual.Pool.MaxConns, "failed to load max conns")
			require.Equal(t, int32(min_conns), actual.Pool.MinConns, "failed to load min conns")
			require.True(t, actual.Pool.LazyConnect, "failed to load lazy connect")
		})
	}
}
