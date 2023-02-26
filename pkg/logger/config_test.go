package logger

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/rs/xid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/validator"
	"gopkg.in/yaml.v3"
)

func TestZapConfig_Load(t *testing.T) {
	envCfgKey := fmt.Sprintf("%s_BUILTIN_CONFIG", constants.GetLoggerPrefix())
	envEncKey := fmt.Sprintf("%s_BUILTIN_ENCODER_CONFIG", constants.GetLoggerPrefix())

	testCases := []struct {
		name      string
		input     string
		cfgKey    string
		encKey    string
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
		expectLen int
	}{
		// ----- test cases start ----- //
		{
			name:      "invalid - empty",
			input:     loggerConfigTestData["empty"],
			cfgKey:    "Production",
			encKey:    "Production",
			expectErr: require.Error,
			expectNil: require.Nil,
			expectLen: 2,
		}, {
			name:      "invalid - builtin",
			input:     loggerConfigTestData["invalid_builtin"],
			cfgKey:    xid.New().String(),
			encKey:    xid.New().String(),
			expectErr: require.Error,
			expectNil: require.Nil,
			expectLen: 2,
		}, {
			name:      "valid - development",
			input:     loggerConfigTestData["valid_devel"],
			cfgKey:    "Production",
			encKey:    "Production",
			expectErr: require.NoError,
			expectNil: require.Nil,
			expectLen: 0,
		}, {
			name:      "valid - production",
			input:     loggerConfigTestData["valid_prod"],
			cfgKey:    "Development",
			encKey:    "Development",
			expectErr: require.NoError,
			expectNil: require.Nil,
			expectLen: 0,
		}, {
			name:      "valid - full constants",
			input:     loggerConfigTestData["valid_config"],
			cfgKey:    "Production",
			encKey:    "Production",
			expectErr: require.NoError,
			expectNil: require.NotNil,
			expectLen: 0,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Configure mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetLoggerFileName(), []byte(testCase.input), 0644), "Failed to write in memory file")

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
			require.True(t, reflect.DeepEqual(expected, actual))

			// Test configuring of environment variable.
			t.Setenv(envCfgKey, testCase.cfgKey)
			t.Setenv(envEncKey, testCase.encKey)
			require.NoErrorf(t, actual.Load(fs), "Failed to load constants file: %v", err)

			require.Equalf(t, testCase.cfgKey, actual.BuiltinConfig, "Failed to load environment variable into constants")
			require.Equalf(t, testCase.encKey, actual.BuiltinEncoderConfig, "Failed to load environment variable into encoder")

			testCase.expectNil(t, actual.GeneralConfig, "Check for nil general constants failed")
		})
	}
}
