package utilities

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/quotes"
)

func TestUtilities_HTTPFiatOpen(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		currencyStr   string
		expectErrMsg  string
		expectErrCode int
		openAccErr    error
		openAccTimes  int
		expectErr     require.ErrorAssertionFunc
	}{
		{
			name:          "invalid currency",
			currencyStr:   "INVALID",
			openAccErr:    nil,
			openAccTimes:  0,
			expectErrCode: http.StatusBadRequest,
			expectErrMsg:  "invalid currency",
			expectErr:     require.Error,
		}, {
			name:          "unknown db failure",
			currencyStr:   "USD",
			openAccErr:    errors.New("unknown error"),
			openAccTimes:  1,
			expectErrCode: http.StatusInternalServerError,
			expectErrMsg:  retryMessage,
			expectErr:     require.Error,
		}, {
			name:          "known db failure",
			currencyStr:   "USD",
			openAccErr:    postgres.ErrNotFound,
			openAccTimes:  1,
			expectErrCode: http.StatusNotFound,
			expectErrMsg:  "records not found",
			expectErr:     require.Error,
		}, {
			name:          "USD",
			currencyStr:   "USD",
			openAccErr:    nil,
			openAccTimes:  1,
			expectErrCode: 0,
			expectErrMsg:  "",
			expectErr:     require.NoError,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockDB := mocks.NewMockPostgres(mockCtrl)

			mockDB.EXPECT().FiatCreateAccount(gomock.Any(), gomock.Any()).
				Return(test.openAccErr).
				Times(test.openAccTimes)

			actualErrCode, actualErrMsg, err := HTTPFiatOpen(mockDB, zapLogger, uuid.UUID{}, test.currencyStr)
			test.expectErr(t, err, "error expectation failed.")
			require.Equal(t, test.expectErrCode, actualErrCode, "error codes mismatched.")
			require.Contains(t, actualErrMsg, test.expectErrMsg, "expected error message mismatched.")
		})
	}
}

func TestUtilities_HTTPFiatDeposit(t *testing.T) {
	t.Parallel()

	validRequest := models.HTTPDepositCurrencyRequest{
		Amount:   decimal.NewFromFloat(49866.13),
		Currency: "USD",
	}

	testCases := []struct {
		name             string
		request          *models.HTTPDepositCurrencyRequest
		expectErrMsg     string
		expectErrCode    int
		depositErr       error
		depositTimes     int
		expectErr        require.ErrorAssertionFunc
		expectNilReceipt require.ValueAssertionFunc
		expectNilPayload require.ValueAssertionFunc
	}{
		{
			name:             "empty request",
			request:          &models.HTTPDepositCurrencyRequest{},
			depositErr:       nil,
			depositTimes:     0,
			expectErrCode:    http.StatusBadRequest,
			expectErrMsg:     "validation",
			expectErr:        require.Error,
			expectNilReceipt: require.Nil,
			expectNilPayload: require.NotNil,
		}, {
			name:             "invalid currency",
			request:          &models.HTTPDepositCurrencyRequest{Currency: "INVALID", Amount: validRequest.Amount},
			depositErr:       nil,
			depositTimes:     0,
			expectErrCode:    http.StatusBadRequest,
			expectErrMsg:     "invalid currency",
			expectErr:        require.Error,
			expectNilReceipt: require.Nil,
			expectNilPayload: require.NotNil,
		}, {
			name: "too many decimal places",
			request: &models.HTTPDepositCurrencyRequest{
				Currency: validRequest.Currency,
				Amount:   decimal.NewFromFloat(49866.123),
			},
			depositErr:       nil,
			depositTimes:     0,
			expectErrCode:    http.StatusBadRequest,
			expectErrMsg:     "invalid amount",
			expectErr:        require.Error,
			expectNilReceipt: require.Nil,
			expectNilPayload: require.NotNil,
		}, {
			name: "negative",
			request: &models.HTTPDepositCurrencyRequest{
				Currency: validRequest.Currency,
				Amount:   decimal.NewFromFloat(-49866.13),
			},
			depositErr:       nil,
			depositTimes:     0,
			expectErrCode:    http.StatusBadRequest,
			expectErrMsg:     "invalid amount",
			expectErr:        require.Error,
			expectNilReceipt: require.Nil,
			expectNilPayload: require.NotNil,
		}, {
			name:             "unknown db failure",
			request:          &validRequest,
			depositErr:       errors.New("unknown error"),
			depositTimes:     1,
			expectErrCode:    http.StatusInternalServerError,
			expectErrMsg:     retryMessage,
			expectErr:        require.Error,
			expectNilReceipt: require.Nil,
			expectNilPayload: require.Nil,
		}, {
			name:             "known db failure",
			request:          &validRequest,
			depositErr:       postgres.ErrNotFound,
			depositTimes:     1,
			expectErrCode:    http.StatusNotFound,
			expectErrMsg:     "records not found",
			expectErr:        require.Error,
			expectNilReceipt: require.Nil,
			expectNilPayload: require.Nil,
		}, {
			name:             "USD",
			request:          &validRequest,
			depositErr:       nil,
			depositTimes:     1,
			expectErrCode:    0,
			expectErrMsg:     "",
			expectErr:        require.NoError,
			expectNilReceipt: require.NotNil,
			expectNilPayload: require.Nil,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockDB := mocks.NewMockPostgres(mockCtrl)

			mockDB.EXPECT().FiatExternalTransfer(gomock.Any(), gomock.Any()).
				Return(&postgres.FiatAccountTransferResult{}, test.depositErr).
				Times(test.depositTimes)

			result, actualErrCode, actualErrMsg, payload, err := HTTPFiatDeposit(mockDB, zapLogger, uuid.UUID{}, test.request)
			test.expectErr(t, err, "error expectation failed.")
			test.expectNilReceipt(t, result, "nil result expectation failed.")
			test.expectNilPayload(t, payload, "nil payload expectation failed.")
			require.Equal(t, test.expectErrCode, actualErrCode, "error codes mismatched.")
			require.Contains(t, actualErrMsg, test.expectErrMsg, "expected error message mismatched.")
		})
	}
}

func TestUtilities_HTTPFiatOffer(t *testing.T) {
	t.Parallel()

	var (
		sourceAmount = decimal.NewFromFloat(23123.12)
		quotesAmount = decimal.NewFromFloat(23100.44)
		quotesRate   = decimal.NewFromFloat(1.35)
		validRequest = models.HTTPExchangeOfferRequest{
			SourceCurrency:      "USD",
			DestinationCurrency: "CAD",
			SourceAmount:        sourceAmount,
		}
	)

	testCases := []struct {
		name             string
		source           string
		destination      string
		expectErrMsg     string
		httpMessage      string
		httpStatus       int
		request          *models.HTTPExchangeOfferRequest
		quotesAmount     decimal.Decimal
		quotesTimes      int
		quotesErr        error
		authEncryptTimes int
		authEncryptErr   error
		redisTimes       int
		redisErr         error
		expectErr        require.ErrorAssertionFunc
		expectNilPayload require.ValueAssertionFunc
	}{
		{
			name:         "validate offer - purchase",
			source:       "INVALID",
			destination:  "CAD",
			expectErrMsg: "INVALID",
			httpMessage:  constants.GetInvalidRequest(),
			httpStatus:   http.StatusBadRequest,
			request: &models.HTTPExchangeOfferRequest{
				SourceCurrency:      "INVALID",
				DestinationCurrency: "CAD",
				SourceAmount:        sourceAmount,
			},
			quotesAmount:     quotesAmount,
			quotesTimes:      0,
			quotesErr:        nil,
			authEncryptTimes: 0,
			authEncryptErr:   nil,
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
			expectNilPayload: require.NotNil,
		}, {
			name:             "crypto conversion - purchase",
			expectErrMsg:     "quote failure",
			httpMessage:      retryMessage,
			httpStatus:       http.StatusInternalServerError,
			request:          &validRequest,
			quotesAmount:     quotesAmount,
			quotesTimes:      1,
			quotesErr:        errors.New("quote failure"),
			authEncryptTimes: 0,
			authEncryptErr:   nil,
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
			expectNilPayload: require.Nil,
		}, {
			name:             "zero amount",
			expectErrMsg:     "too small",
			httpMessage:      "too small",
			httpStatus:       http.StatusBadRequest,
			request:          &validRequest,
			quotesAmount:     decimal.NewFromFloat(0),
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 0,
			authEncryptErr:   nil,
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
			expectNilPayload: require.Nil,
		}, {
			name:             "encryption failure",
			expectErrMsg:     "encryption failure",
			httpMessage:      retryMessage,
			httpStatus:       http.StatusInternalServerError,
			request:          &validRequest,
			quotesAmount:     quotesAmount,
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 1,
			authEncryptErr:   errors.New("encryption failure"),
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
			expectNilPayload: require.Nil,
		}, {
			name:             "cache failure",
			expectErrMsg:     "cache failure",
			httpMessage:      retryMessage,
			httpStatus:       http.StatusInternalServerError,
			request:          &validRequest,
			quotesAmount:     quotesAmount,
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisTimes:       1,
			redisErr:         errors.New("cache failure"),
			expectErr:        require.Error,
			expectNilPayload: require.Nil,
		}, {
			name:             "valid",
			expectErrMsg:     "",
			httpMessage:      "",
			httpStatus:       0,
			request:          &validRequest,
			quotesAmount:     quotesAmount,
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisTimes:       1,
			redisErr:         nil,
			expectErr:        require.NoError,
			expectNilPayload: require.Nil,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockAuth := mocks.NewMockAuth(mockCtrl)
			mockCache := mocks.NewMockRedis(mockCtrl)
			mockQuotes := quotes.NewMockQuotes(mockCtrl)

			gomock.InOrder(
				mockQuotes.EXPECT().FiatConversion(
					test.request.SourceCurrency, test.request.DestinationCurrency, test.request.SourceAmount, nil).
					Return(quotesRate, test.quotesAmount, test.quotesErr).
					Times(test.quotesTimes),

				mockAuth.EXPECT().EncryptToString(gomock.Any()).
					Return("OFFER-ID", test.authEncryptErr).
					Times(test.authEncryptTimes),

				mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(test.redisErr).
					Times(test.redisTimes),
			)

			offer, status, msg, payload, err := HTTPFiatOffer(mockAuth, mockCache, zapLogger, mockQuotes,
				uuid.UUID{}, test.request)
			test.expectErr(t, err, "error expectation failed.")
			test.expectNilPayload(t, payload, "nil payload expectation failed.")

			if err != nil {
				require.Contains(t, err.Error(), test.expectErrMsg, "error message is incorrect.")
				require.Contains(t, msg, test.httpMessage, "http error message mismatched.")
				require.Equal(t, test.httpStatus, status, "http status mismatched.")

				return
			}

			require.Equal(t, test.request.SourceCurrency, offer.SourceAcc, "source account mismatch.")
			require.Equal(t, test.request.DestinationCurrency, offer.DestinationAcc, "destination account mismatch.")
			require.Equal(t, test.request.SourceAmount, offer.DebitAmount, "debit amount mismatch.")
			require.Equal(t, quotesRate, offer.Rate, "offer rate mismatch.")
			require.Equal(t, test.quotesAmount, offer.Amount, "offer amount mismatch.")
		})
	}
}

func TestUtilities_HTTPFiatBalancePaginatedRequest(t *testing.T) {
	t.Parallel()

	encAED, err := testAuth.EncryptToString([]byte("AED"))
	require.NoError(t, err, "failed to encrypt AED currency.")

	encUSD, err := testAuth.EncryptToString([]byte("USD"))
	require.NoError(t, err, "failed to encrypt USD currency.")

	encEUR, err := testAuth.EncryptToString([]byte("EUR"))
	require.NoError(t, err, "failed to encrypt EUR currency.")

	testCases := []struct {
		name           string
		currencyStr    string
		limitStr       string
		expectCurrency postgres.Currency
		expectLimit    int32
	}{
		{
			name:           "empty currency",
			currencyStr:    "",
			limitStr:       "5",
			expectCurrency: postgres.CurrencyAED,
			expectLimit:    5,
		}, {
			name:           "AED",
			currencyStr:    encAED,
			limitStr:       "5",
			expectCurrency: postgres.CurrencyAED,
			expectLimit:    5,
		}, {
			name:           "USD",
			currencyStr:    encUSD,
			limitStr:       "5",
			expectCurrency: postgres.CurrencyUSD,
			expectLimit:    5,
		}, {
			name:           "EUR",
			currencyStr:    encEUR,
			limitStr:       "5",
			expectCurrency: postgres.CurrencyEUR,
			expectLimit:    5,
		}, {
			name:           "base bound limit",
			currencyStr:    encEUR,
			limitStr:       "0",
			expectCurrency: postgres.CurrencyEUR,
			expectLimit:    10,
		}, {
			name:           "above base bound limit",
			currencyStr:    encEUR,
			limitStr:       "999",
			expectCurrency: postgres.CurrencyEUR,
			expectLimit:    999,
		}, {
			name:           "empty request",
			currencyStr:    "",
			limitStr:       "",
			expectCurrency: postgres.CurrencyAED,
			expectLimit:    10,
		}, {
			name:           "empty currency",
			currencyStr:    "",
			limitStr:       "999",
			expectCurrency: postgres.CurrencyAED,
			expectLimit:    999,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actualCurr, actualLimit, err := HTTPFiatBalancePaginatedRequest(testAuth, test.currencyStr, test.limitStr)
			require.NoError(t, err, "error returned from query unpacking")
			require.Equal(t, test.expectCurrency, actualCurr, "currencies mismatched.")
			require.Equal(t, test.expectLimit, actualLimit, "request limit size mismatched.")
		})
	}
}
