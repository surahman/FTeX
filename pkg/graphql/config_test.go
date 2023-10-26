package graphql

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

func TestGraphQLConfigs_Load(t *testing.T) {
	keyspaceServer := constants.HTTPGraphQLPrefix() + "_SERVER."
	keyspaceAuth := constants.HTTPGraphQLPrefix() + "_AUTHORIZATION."

	testCases := []struct {
		name         string
		input        string
		expectErr    require.ErrorAssertionFunc
		expectErrCnt int
	}{
		{
			name:         "empty - etc dir",
			input:        graphQLConfigTestData["empty"],
			expectErr:    require.Error,
			expectErrCnt: 9,
		}, {
			name:         "valid - etc dir",
			input:        graphQLConfigTestData["valid"],
			expectErr:    require.NoError,
			expectErrCnt: 0,
		}, {
			name:         "out of range port - etc dir",
			input:        graphQLConfigTestData["out of range port"],
			expectErr:    require.Error,
			expectErrCnt: 1,
		}, {
			name:         "out of range time delay - etc dir",
			input:        graphQLConfigTestData["out of range time delay"],
			expectErr:    require.Error,
			expectErrCnt: 4,
		}, {
			name:         "no base path - etc dir",
			input:        graphQLConfigTestData["no base path"],
			expectErr:    require.Error,
			expectErrCnt: 1,
		}, {
			name:         "no playground path - etc dir",
			input:        graphQLConfigTestData["no playground path"],
			expectErr:    require.Error,
			expectErrCnt: 1,
		}, {
			name:         "no query path - etc dir",
			input:        graphQLConfigTestData["no query path"],
			expectErr:    require.Error,
			expectErrCnt: 1,
		}, {
			name:         "no read timeout - etc dir",
			input:        graphQLConfigTestData["no read timeout"],
			expectErr:    require.Error,
			expectErrCnt: 1,
		}, {
			name:         "no write timeout - etc dir",
			input:        graphQLConfigTestData["no write timeout"],
			expectErr:    require.Error,
			expectErrCnt: 1,
		}, {
			name:         "no read header timeout - etc dir",
			input:        graphQLConfigTestData["no read header timeout"],
			expectErr:    require.Error,
			expectErrCnt: 1,
		}, {
			name:         "no auth header - etc dir",
			input:        graphQLConfigTestData["no auth header"],
			expectErr:    require.Error,
			expectErrCnt: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Configure mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(constants.EtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, constants.EtcDir()+constants.HTTPGraphQLFileName(),
				[]byte(testCase.input), 0644), "Failed to write in memory file")

			// Load from mock filesystem.
			actual := &config{}
			err := actual.Load(fs)
			testCase.expectErr(t, err)

			validationError := &validator.ValidationError{}
			if errors.As(err, &validationError) {
				require.Lenf(t, validationError.Errors, testCase.expectErrCnt,
					"expected errors count is incorrect: %v", err)

				return
			}

			// Load expected struct.
			expected := &config{}
			require.NoError(t, yaml.Unmarshal([]byte(testCase.input), expected), "failed to unmarshal expected constants")
			require.True(t, reflect.DeepEqual(expected, actual))

			// Test configuring of environment variable.
			basePath := xid.New().String()
			playgroundPath := xid.New().String()
			queryPath := xid.New().String()
			headerKey := xid.New().String()
			portNumber := 1600
			shutdownDelay := time.Duration(36)
			readTimeout := time.Duration(4)
			writeTimeout := time.Duration(5)
			readHeaderTimeout := time.Duration(7)
			t.Setenv(keyspaceServer+"BASEPATH", basePath)
			t.Setenv(keyspaceServer+"PLAYGROUNDPATH", playgroundPath)
			t.Setenv(keyspaceServer+"QUERYPATH", queryPath)
			t.Setenv(keyspaceServer+"PORTNUMBER", strconv.Itoa(portNumber))
			t.Setenv(keyspaceServer+"SHUTDOWNDELAY", shutdownDelay.String())
			t.Setenv(keyspaceServer+"READTIMEOUT", readTimeout.String())
			t.Setenv(keyspaceServer+"WRITETIMEOUT", writeTimeout.String())
			t.Setenv(keyspaceServer+"READHEADERTIMEOUT", readHeaderTimeout.String())
			t.Setenv(keyspaceAuth+"HEADERKEY", headerKey)

			err = actual.Load(fs)
			require.NoErrorf(t, err, "Failed to load constants file: %v", err)

			require.Equal(t, basePath, actual.Server.BasePath, "Failed to load base path environment variable into configs")
			require.Equal(t, playgroundPath, actual.Server.PlaygroundPath,
				"Failed to load playground path environment variable into configs")
			require.Equal(t, queryPath, actual.Server.QueryPath,
				"Failed to load query path environment variable into configs")
			require.Equal(t, portNumber, actual.Server.PortNumber, "Failed to load port environment variable into configs")
			require.Equal(t, shutdownDelay, actual.Server.ShutdownDelay,
				"Failed to load shutdown delay environment variable into configs")
			require.Equal(t, readTimeout, actual.Server.ReadTimeout,
				"failed to load read timeout environment variable into configs")
			require.Equal(t, writeTimeout, actual.Server.WriteTimeout,
				"failed to load write timeout environment variable into configs")
			require.Equal(t, readHeaderTimeout, actual.Server.ReadHeaderTimeout,
				"failed to load read header timeout environment variable into configs")
			require.Equal(t, headerKey, actual.Authorization.HeaderKey,
				"Failed to load authorization header key environment variable into configs")
		})
	}
}
