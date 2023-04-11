package quotes

import (
	"errors"
	"testing"
	"time"

	"github.com/rs/xid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/validator"
)

func TestQuotesConfigs_Load(t *testing.T) {
	envFiatKey := constants.GetQuotesPrefix() + "_FIATCURRENCY."
	envCryptoKey := constants.GetQuotesPrefix() + "_CRYPTOCURRENCY."
	envConnKey := constants.GetQuotesPrefix() + "_CONNECTION."

	testCases := []struct {
		name         string
		input        string
		envValue     string
		expectErrCnt int
		expectErr    require.ErrorAssertionFunc
	}{
		{
			name:         "empty - etc dir",
			input:        quotesConfigTestData["empty"],
			expectErrCnt: 6,
			expectErr:    require.Error,
		}, {
			name:         "valid - etc dir",
			input:        quotesConfigTestData["valid"],
			expectErrCnt: 0,
			expectErr:    require.NoError,
		}, {
			name:         "no api key fiat",
			input:        quotesConfigTestData["no fiat api key"],
			expectErrCnt: 1,
			expectErr:    require.Error,
		}, {
			name:         "no api endpoint fiat",
			input:        quotesConfigTestData["no fiat api endpoint"],
			expectErrCnt: 1,
			expectErr:    require.Error,
		}, {
			name:         "no fiat",
			input:        quotesConfigTestData["no fiat"],
			expectErrCnt: 2,
			expectErr:    require.Error,
		}, {
			name:         "no api key crypto",
			input:        quotesConfigTestData["no crypto api key"],
			expectErrCnt: 1,
			expectErr:    require.Error,
		}, {
			name:         "no api endpoint crypto",
			input:        quotesConfigTestData["no crypto api endpoint"],
			expectErrCnt: 1,
			expectErr:    require.Error,
		}, {
			name:         "no crypto",
			input:        quotesConfigTestData["no crypto"],
			expectErrCnt: 2,
			expectErr:    require.Error,
		}, {
			name:         "no connection user-agent",
			input:        quotesConfigTestData["no connection user-agent"],
			expectErrCnt: 1,
			expectErr:    require.Error,
		}, {
			name:         "no connection timeout",
			input:        quotesConfigTestData["no connection timeout"],
			expectErrCnt: 1,
			expectErr:    require.Error,
		}, {
			name:         "no connection",
			input:        quotesConfigTestData["no connection"],
			expectErrCnt: 2,
			expectErr:    require.Error,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Configure mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetQuotesFileName(),
				[]byte(testCase.input), 0644), "Failed to write in memory file")

			// Load from mock filesystem.
			actual := &config{}
			err := actual.Load(fs)
			testCase.expectErr(t, err, "Failed to load test config from mock filesystem.")

			validationError := &validator.ValidationError{}
			if errors.As(err, &validationError) {
				require.Equalf(t, testCase.expectErrCnt, len(validationError.Errors),
					"expected errors count is incorrect: %v", err)

				return
			}

			// Test configuring of environment variable.
			apiKeyFiat := xid.New().String()
			apiEndpointFiat := xid.New().String()
			t.Setenv(envFiatKey+"APIKEY", apiKeyFiat)
			t.Setenv(envFiatKey+"ENDPOINT", apiEndpointFiat)

			apiKeyCrypto := xid.New().String()
			apiEndpointCrypto := xid.New().String()
			t.Setenv(envCryptoKey+"APIKEY", apiKeyCrypto)
			t.Setenv(envCryptoKey+"ENDPOINT", apiEndpointCrypto)

			timeout := 999 * time.Second
			userAgent := xid.New().String()
			t.Setenv(envConnKey+"TIMEOUT", timeout.String())
			t.Setenv(envConnKey+"USERAGENT", userAgent)

			require.NoErrorf(t, actual.Load(fs), "failed to load configurations file: %v", err)

			require.Equal(t, apiKeyFiat, actual.FiatCurrency.APIKey, "failed to load fiat API Key.")
			require.Equal(t, apiEndpointFiat, actual.FiatCurrency.Endpoint, "failed to load fiat API endpoint.")

			require.Equal(t, apiKeyCrypto, actual.CryptoCurrency.APIKey, "failed to load crypto API Key.")
			require.Equal(t, apiEndpointCrypto, actual.CryptoCurrency.Endpoint, "failed to load crypto API endpoint.")

			require.Equal(t, timeout, actual.Connection.Timeout, "failed to load timeout.")
			require.Equal(t, userAgent, actual.Connection.UserAgent, "failed to load user-agent.")
		})
	}
}
