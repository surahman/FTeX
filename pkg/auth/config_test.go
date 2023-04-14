package auth

import (
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/rs/xid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/validator"
	"gopkg.in/yaml.v3"
)

func TestAuthConfigs_Load(t *testing.T) {
	keyspaceJwt := constants.GetAuthPrefix() + "_JWT."
	keyspaceGen := constants.GetAuthPrefix() + "_GENERAL."

	testCases := []struct {
		name         string
		input        string
		expectErr    require.ErrorAssertionFunc
		expectErrCnt int
	}{
		{
			name:         "empty - etc dir",
			input:        authConfigTestData["empty"],
			expectErr:    require.Error,
			expectErrCnt: 6,
		}, {
			name:         "valid - etc dir",
			input:        authConfigTestData["valid"],
			expectErr:    require.NoError,
			expectErrCnt: 0,
		}, {
			name:         "no issuer - etc dir",
			input:        authConfigTestData["no_issuer"],
			expectErr:    require.Error,
			expectErrCnt: 1,
		}, {
			name:         "bcrypt cost below 4 - etc dir",
			input:        authConfigTestData["bcrypt_cost_below_4"],
			expectErr:    require.Error,
			expectErrCnt: 1,
		}, {
			name:         "bcrypt cost above 31 - etc dir",
			input:        authConfigTestData["bcrypt_cost_above_31"],
			expectErr:    require.Error,
			expectErrCnt: 1,
		}, {
			name:         "jwt expiration below 10s - etc dir",
			input:        authConfigTestData["jwt_expiration_below_60s"],
			expectErr:    require.Error,
			expectErrCnt: 1,
		}, {
			name:         "jwt key below 8 - etc dir",
			input:        authConfigTestData["jwt_key_below_8"],
			expectErr:    require.Error,
			expectErrCnt: 1,
		}, {
			name:         "jwt key above 256 - etc dir",
			input:        authConfigTestData["jwt_key_above_256"],
			expectErr:    require.Error,
			expectErrCnt: 1,
		}, {
			name:         "low refresh threshold - etc dir",
			input:        authConfigTestData["low_refresh_threshold"],
			expectErr:    require.Error,
			expectErrCnt: 1,
		}, {
			name:         "refresh_threshold_gt_expiration - etc dir",
			input:        authConfigTestData["refresh_threshold_gt_expiration"],
			expectErr:    require.Error,
			expectErrCnt: 2,
		}, {
			name:         "crypto_key_too_short- etc dir",
			input:        authConfigTestData["crypto_key_too_short"],
			expectErr:    require.Error,
			expectErrCnt: 1,
		}, {
			name:         "crypto_key_too_long- etc dir",
			input:        authConfigTestData["crypto_key_too_long"],
			expectErr:    require.Error,
			expectErrCnt: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Configure mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetAuthFileName(),
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

			// Load expected struct.
			expected := &config{}
			require.NoError(t, yaml.Unmarshal([]byte(testCase.input), expected), "failed to unmarshal expected constants")
			require.True(t, reflect.DeepEqual(expected, actual))

			// Test configuring of environment variable.
			testKey := xid.New().String()
			testExpDur := int64(999)
			testRefThreshold := int64(555)
			testBcryptCost := 16
			testIssuer := "test issuer"
			testCryptoSecret := "**crypto secret set in env var**"
			t.Setenv(keyspaceJwt+"KEY", testKey)
			t.Setenv(keyspaceJwt+"ISSUER", testIssuer)
			t.Setenv(keyspaceJwt+"EXPIRATION_DURATION", strconv.FormatInt(testExpDur, 10))
			t.Setenv(keyspaceJwt+"REFRESH_THRESHOLD", strconv.FormatInt(testRefThreshold, 10))
			t.Setenv(keyspaceGen+"BCRYPT_COST", strconv.Itoa(testBcryptCost))
			t.Setenv(keyspaceGen+"CRYPTO_SECRET", testCryptoSecret)

			err = actual.Load(fs)
			require.NoErrorf(t, err, "Failed to load constants file: %v", err)

			require.Equal(t, testKey, actual.JWTConfig.Key, "Failed to load key environment variable into configs")
			require.Equal(t, testIssuer, actual.JWTConfig.Issuer, "Failed to load issuer environment variable into configs")
			require.Equal(t, testExpDur, actual.JWTConfig.ExpirationDuration,
				"Failed to load duration environment variable into configs")
			require.Equal(t, testRefThreshold, actual.JWTConfig.RefreshThreshold,
				"Failed to load refresh threshold environment variable into configs")
			require.Equal(t, testBcryptCost, actual.General.BcryptCost,
				"Failed to load bcrypt cost environment variable into configs")
			require.Equal(t, testCryptoSecret, actual.General.CryptoSecret,
				"Failed to load crypto secret environment variable into configs")
		})
	}
}
