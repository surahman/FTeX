package quotes

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/models"
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
			fileName:  constants.QuotesFileName(),
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
			require.NoError(t, fs.MkdirAll(constants.EtcDir(), 0644), "failed to create in memory directory.")
			require.NoError(t, afero.WriteFile(fs, constants.EtcDir()+test.fileName, []byte(test.input), 0644),
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
	require.Error(t, err, "no config should fail")
	require.Nil(t, testClient, "failure should return nil client.")

	testClient, err = configFiatClient(testConfigs)
	require.NoError(t, err, "failed to configure Quotes.")
	require.NotNil(t, testClient, "failed to configure Quotes.")
}

func TestQuotesImpl_ConfigCryptoClient(t *testing.T) {
	t.Parallel()

	testClient, err := configCryptoClient(nil)
	require.Error(t, err, "no config should fail.")
	require.Nil(t, testClient, "failure should return nil client.")

	testClient, err = configCryptoClient(testConfigs)
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

			result, err := quotes.fiatQuote(test.source, test.destination, amount)
			test.errExpectation(t, err, "error expectation failed.")
			test.successExpectation(t, result.Success, "success code incorrectly set.")

			if err != nil {
				require.NotEqual(t, 200, result.Error.Code, "received valid response code on error.")
				require.NotEmpty(t, result.Error.Type, "received no type on error.")
				require.NotEmpty(t, result.Error.Info, "received no info on error.")

				return
			}

			require.Equal(t, test.source, result.Query.From, "source currency mismatched in query parameters.")
			require.Equal(t, test.destination, result.Query.To, "destination currency mismatched in query parameters.")
			require.True(t, amount.Equal(result.Query.Amount), "transfer amount mismatched in query parameters.")
			require.True(t, result.Info.Rate.IsPositive() && !result.Info.Rate.IsZero(), "invalid rate.")
			require.Positive(t, result.Info.Timestamp, "invalid timestamp.")
			require.True(t, result.Result.IsPositive() && !result.Result.IsZero(), "invalid result.")
			require.NotEmpty(t, result.Date, "invalid date.")
		})
	}
}

func TestQuotesImpl_FiatConversion(t *testing.T) {
	t.Parallel()

	amount, err := decimal.NewFromString("1000")
	require.NoError(t, err, "failed to prepare test amount.")

	exchangeRate, convertedAmount, err := quotes.FiatConversion("USD", "CAD", amount, nil)
	require.NoError(t, err, "failed to retrieve price quote.")
	require.False(t, exchangeRate.IsZero(), "conversion rate not returned.")
	require.False(t, convertedAmount.IsZero(), "converted amount not returned.")
}

func TestQuotesImpl_FiatConversion_Mock(t *testing.T) {
	t.Parallel()

	amount, err := decimal.NewFromString("1000")
	require.NoError(t, err, "could not prepare amount to convert.")

	testCases := []struct {
		name         string
		rate         decimal.Decimal
		expectAmount decimal.Decimal
		expectErr    require.ErrorAssertionFunc
		expectTimes  int
		err          error
	}{
		{
			name:         "quote failure",
			rate:         decimal.NewFromFloat(0),
			expectAmount: decimal.NewFromFloat(0),
			expectErr:    require.Error,
			expectTimes:  1,
			err:          NewError("quote failure"),
		}, {
			name:         "3000.65",
			rate:         decimal.NewFromFloat(3.000654),
			expectAmount: decimal.NewFromFloat(3000.65),
			expectErr:    require.NoError,
			expectTimes:  1,
			err:          nil,
		}, {
			name:         "298.64",
			rate:         decimal.NewFromFloat(0.298645),
			expectAmount: decimal.NewFromFloat(298.64),
			expectErr:    require.NoError,
			expectTimes:  1,
			err:          nil,
		}, {
			name:         "298.65",
			rate:         decimal.NewFromFloat(0.298651),
			expectAmount: decimal.NewFromFloat(298.65),
			expectErr:    require.NoError,
			expectTimes:  1,
			err:          nil,
		}, {
			name:         "298.66",
			rate:         decimal.NewFromFloat(0.298655),
			expectAmount: decimal.NewFromFloat(298.66),
			expectErr:    require.NoError,
			expectTimes:  1,
			err:          nil,
		}, {
			name:         "0.66",
			rate:         decimal.NewFromFloat(0.000655),
			expectAmount: decimal.NewFromFloat(0.66),
			expectErr:    require.NoError,
			expectTimes:  1,
			err:          nil,
		}, {
			name:         "0.64",
			rate:         decimal.NewFromFloat(0.000645),
			expectAmount: decimal.NewFromFloat(0.64),
			expectErr:    require.NoError,
			expectTimes:  1,
			err:          nil,
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockQuotes := NewMockQuotes(mockCtrl)

			quote := models.FiatQuote{Info: models.FiatInfo{Rate: test.rate}}

			mockQuotes.EXPECT().fiatQuote(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(quote, test.err).
				Times(test.expectTimes)

			exchangeRate, convertedAmount, err := quotes.FiatConversion("XYZ", "UVW", amount, mockQuotes.fiatQuote)
			test.expectErr(t, err, "error expectation failed.")
			require.True(t, exchangeRate.Equal(test.rate), "exchange rate is incorrect.")
			require.True(t, convertedAmount.Equal(test.expectAmount), "converted amount is incorrect.")
		})
	}
}

func TestQuotesImpl_CryptoQuote(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		source         string
		destination    string
		errExpectation require.ErrorAssertionFunc
	}{
		{
			name:           "BTC to USD",
			source:         "BTC",
			destination:    "USD",
			errExpectation: require.NoError,
		}, {
			name:           "USD to BTC",
			source:         "USD",
			destination:    "BTC",
			errExpectation: require.NoError,
		}, {
			name:           "invalid",
			source:         "INVALID",
			destination:    "CAD",
			errExpectation: require.Error,
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result, err := quotes.cryptoQuote(test.source, test.destination)
			test.errExpectation(t, err, "error expectation failed.")

			if err != nil {
				return
			}

			require.Equal(t, test.source, result.BaseCurrency, "source mismatched base currency.")
			require.Equal(t, test.destination, result.QuoteCurrency, "destination mismatched quote currency.")
			require.NotEmpty(t, result.Time, "no time stamp returned.")
			require.True(t, result.Rate.IsPositive() && !result.Rate.IsZero(), "invalid rate.")
		})
	}
}

func TestQuotesImpl_CryptoConversion(t *testing.T) {
	t.Parallel()

	amount := decimal.NewFromFloat(1000)

	exchangeRate, convertedAmount, err := quotes.CryptoConversion("USD", "BTC", amount, true, nil)
	require.NoError(t, err, "failed to retrieve price quote.")
	require.False(t, exchangeRate.IsZero(), "conversion rate not returned.")
	require.False(t, convertedAmount.IsZero(), "converted amount not returned.")

	exchangeRate, convertedAmount, err = quotes.CryptoConversion("BTC", "USD", amount, false, nil)
	require.NoError(t, err, "failed to retrieve price quote.")
	require.False(t, exchangeRate.IsZero(), "conversion rate not returned.")
	require.False(t, convertedAmount.IsZero(), "converted amount not returned.")
}

func TestQuotesImpl_CryptoConversion_Mock(t *testing.T) {
	t.Parallel()

	amount, err := decimal.NewFromString("1000")
	require.NoError(t, err, "could not prepare amount to convert.")

	testCases := []struct {
		name               string
		rate               decimal.Decimal
		expectAmount       decimal.Decimal
		expectErr          require.ErrorAssertionFunc
		expectTimes        int
		isPurchasingCrypto bool
		err                error
	}{
		{
			name:               "quote failure",
			rate:               decimal.NewFromFloat(0),
			expectAmount:       decimal.NewFromFloat(0),
			isPurchasingCrypto: true,
			expectErr:          require.Error,
			expectTimes:        1,
			err:                NewError("quote failure"),
		}, {
			name:               "3000.00000065 - Crypto Purchase",
			rate:               decimal.NewFromFloat(3.000000000654),
			expectAmount:       decimal.NewFromFloat(3000.00000065),
			isPurchasingCrypto: true,
			expectErr:          require.NoError,
			expectTimes:        1,
			err:                nil,
		}, {
			name:               "298.00000064 - Crypto Purchase",
			rate:               decimal.NewFromFloat(0.298000000645),
			expectAmount:       decimal.NewFromFloat(298.00000064),
			isPurchasingCrypto: true,
			expectErr:          require.NoError,
			expectTimes:        1,
			err:                nil,
		}, {
			name:               "298.00000065 - Crypto Purchase",
			rate:               decimal.NewFromFloat(0.298000000651),
			expectAmount:       decimal.NewFromFloat(298.00000065),
			isPurchasingCrypto: true,
			expectErr:          require.NoError,
			expectTimes:        1,
			err:                nil,
		}, {
			name:               "298.00000066 - Crypto Purchase",
			rate:               decimal.NewFromFloat(0.298000000655),
			expectAmount:       decimal.NewFromFloat(298.00000066),
			isPurchasingCrypto: true,
			expectErr:          require.NoError,
			expectTimes:        1,
			err:                nil,
		}, {
			name:               "0.00000066 - Crypto Purchase",
			rate:               decimal.NewFromFloat(0.000000000655),
			expectAmount:       decimal.NewFromFloat(0.00000066),
			isPurchasingCrypto: true,
			expectErr:          require.NoError,
			expectTimes:        1,
			err:                nil,
		}, {
			name:               "0.00000064 - Crypto Purchase",
			rate:               decimal.NewFromFloat(0.000000000645),
			expectAmount:       decimal.NewFromFloat(0.00000064),
			isPurchasingCrypto: true,
			expectErr:          require.NoError,
			expectTimes:        1,
			err:                nil,
		}, {
			name:               "3000.65 - Crypto sale",
			rate:               decimal.NewFromFloat(3.000654),
			expectAmount:       decimal.NewFromFloat(3000.65),
			isPurchasingCrypto: false,
			expectErr:          require.NoError,
			expectTimes:        1,
			err:                nil,
		}, {
			name:               "298.64 - Crypto sale",
			rate:               decimal.NewFromFloat(0.298645),
			expectAmount:       decimal.NewFromFloat(298.64),
			isPurchasingCrypto: false,
			expectErr:          require.NoError,
			expectTimes:        1,
			err:                nil,
		}, {
			name:               "298.65 - Crypto sale",
			rate:               decimal.NewFromFloat(0.298651),
			expectAmount:       decimal.NewFromFloat(298.65),
			isPurchasingCrypto: false,
			expectErr:          require.NoError,
			expectTimes:        1,
			err:                nil,
		}, {
			name:               "298.66 - Crypto sale",
			rate:               decimal.NewFromFloat(0.298655),
			expectAmount:       decimal.NewFromFloat(298.66),
			isPurchasingCrypto: false,
			expectErr:          require.NoError,
			expectTimes:        1,
			err:                nil,
		}, {
			name:               "0.66 - Crypto sale",
			rate:               decimal.NewFromFloat(0.000655),
			expectAmount:       decimal.NewFromFloat(0.66),
			isPurchasingCrypto: false,
			expectErr:          require.NoError,
			expectTimes:        1,
			err:                nil,
		}, {
			name:               "0.64 - Crypto sale",
			rate:               decimal.NewFromFloat(0.000645),
			expectAmount:       decimal.NewFromFloat(0.64),
			isPurchasingCrypto: false,
			expectErr:          require.NoError,
			expectTimes:        1,
			err:                nil,
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockQuotes := NewMockQuotes(mockCtrl)

			quote := models.CryptoQuote{Rate: test.rate}

			mockQuotes.EXPECT().cryptoQuote(gomock.Any(), gomock.Any()).
				Return(quote, test.err).
				Times(test.expectTimes)

			exchangeRate, convertedAmount, err := quotes.CryptoConversion(
				"XYZ", "UVW", amount, test.isPurchasingCrypto, mockQuotes.cryptoQuote)
			test.expectErr(t, err, "error expectation failed.")
			require.True(t, exchangeRate.Equal(test.rate), "exchange rate is incorrect.")
			require.True(t, convertedAmount.Equal(test.expectAmount), "converted amount is incorrect.")
		})
	}
}
