package quotes

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
)

func TestQuotesImpl_New(t *testing.T) {
	testCases := []struct {
		name      string
		fileName  string
		input     string
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
	}{
		{
			name:      "File found",
			fileName:  constants.GetQuotesFileName(),
			input:     quotesConfigTestData["valid"],
			expectErr: require.NoError,
			expectNil: require.NotNil,
		}, {
			name:      "File not found",
			fileName:  "wrong_file_name.yaml",
			input:     quotesConfigTestData["valid"],
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

			c, err := newQuotesImpl(&fs, zapLogger)
			test.expectErr(t, err)
			test.expectNil(t, c)
		})
	}
}

func TestQuotesImpl_ConfigFiatClient(t *testing.T) {
	t.Parallel()

	testClient, err := configFiatClient(nil)
	require.Error(t, err, "neither config nor logger.")
	require.Nil(t, testClient, "failure should return nil client.")

	testClient, err = configFiatClient(testConfigs)
	require.NoError(t, err, "failed to configure Quotes.")
	require.NotNil(t, testClient, "failed to configure Quotes.")
}

func TestQuotesImpl_FiatQuote(t *testing.T) {
	t.Parallel()

	amount, err := decimal.NewFromString("1000")
	require.NoError(t, err, "failed to prepare test amount.")

	testCases := []struct {
		name               string
		source             string
		destination        string
		successExpectation require.BoolAssertionFunc
		errExpectation     require.ErrorAssertionFunc
	}{
		{
			name:               "USD to CAD",
			source:             "USD",
			destination:        "CAD",
			successExpectation: require.True,
			errExpectation:     require.NoError,
		}, {
			name:               "invalid",
			source:             "INVALID",
			destination:        "CAD",
			successExpectation: require.False,
			errExpectation:     require.Error,
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result, err := quotes.FiatQuote(test.source, test.destination, amount)
			test.errExpectation(t, err, "error expectation failed.")
			test.successExpectation(t, result.Success, "success code incorrectly set.")

			if err != nil {
				require.False(t, result.Error.Code == 200, "received valid response code on error.")
				require.True(t, len(result.Error.Type) > 0, "received no type on error.")
				require.True(t, len(result.Error.Info) > 0, "received no info on error.")

				return
			}

			require.Equal(t, test.source, result.Query.From, "source currency mismatched in query parameters.")
			require.Equal(t, test.destination, result.Query.To, "destination currency mismatched in query parameters.")
			require.True(t, amount.Equal(result.Query.Amount), "transfer amount mismatched in query parameters.")
			require.True(t, result.Info.Rate.IsPositive() && !result.Info.Rate.IsZero(), "invalid rate.")
			require.True(t, result.Info.Timestamp > 0, "invalid timestamp.")
			require.True(t, result.Result.IsPositive() && !result.Result.IsZero(), "invalid result.")
			require.True(t, len(result.Date) > 0, "invalid date.")
		})
	}
}
