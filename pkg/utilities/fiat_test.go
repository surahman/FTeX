package utilities

import (
	"errors"
	"fmt"
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
	"github.com/surahman/FTeX/pkg/redis"
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
			expectErrMsg:  constants.GetInvalidCurrencyString(),
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
			expectErrMsg:     constants.GetValidationString(),
			expectErr:        require.Error,
			expectNilReceipt: require.Nil,
			expectNilPayload: require.NotNil,
		}, {
			name:             "invalid currency",
			request:          &models.HTTPDepositCurrencyRequest{Currency: "INVALID", Amount: validRequest.Amount},
			depositErr:       nil,
			depositTimes:     0,
			expectErrCode:    http.StatusBadRequest,
			expectErrMsg:     constants.GetInvalidCurrencyString(),
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

func TestUtilities_HTTPFiatTransfer(t *testing.T) { //nolint:maintidx
	t.Parallel()

	validDecimal, err := decimal.NewFromString("10101.11")
	require.NoError(t, err, "failed to parse valid decimal.")

	validClientID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate valid client id.")

	invalidClientID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate invalid client id.")

	validOfferID := []byte("VALID")
	validOffer := models.HTTPExchangeOfferResponse{
		PriceQuote: models.PriceQuote{
			ClientID:       validClientID,
			SourceAcc:      "USD",
			DestinationAcc: "CAD",
			Rate:           validDecimal,
			Amount:         validDecimal,
		},
	}

	invalidOfferClientID := models.HTTPExchangeOfferResponse{
		PriceQuote: models.PriceQuote{
			ClientID:       invalidClientID,
			SourceAcc:      "USD",
			DestinationAcc: "CAD",
			Rate:           validDecimal,
			Amount:         validDecimal,
		},
	}

	invalidOfferSource := models.HTTPExchangeOfferResponse{
		PriceQuote: models.PriceQuote{
			ClientID:       validClientID,
			SourceAcc:      "INVALID",
			DestinationAcc: "CAD",
			Rate:           validDecimal,
			Amount:         validDecimal,
		},
	}

	testCases := []struct {
		name              string
		expectedMsg       string
		expectedStatus    int
		request           models.HTTPTransferRequest
		authDecryptErr    error
		authDecryptTimes  int
		redisGetData      models.HTTPExchangeOfferResponse
		redisGetErr       error
		redisGetTimes     int
		redisDelErr       error
		redisDelTimes     int
		internalXferErr   error
		internalXferTimes int
		expectErr         require.ErrorAssertionFunc
		expectNilResponse require.ValueAssertionFunc
		expectNilPayload  require.ValueAssertionFunc
	}{
		{
			name:              "empty request",
			expectedMsg:       constants.GetValidationString(),
			expectedStatus:    http.StatusBadRequest,
			request:           models.HTTPTransferRequest{},
			authDecryptErr:    nil,
			authDecryptTimes:  0,
			redisGetData:      validOffer,
			redisGetErr:       nil,
			redisGetTimes:     0,
			redisDelErr:       nil,
			redisDelTimes:     0,
			internalXferErr:   nil,
			internalXferTimes: 0,
			expectErr:         require.Error,
			expectNilResponse: require.Nil,
			expectNilPayload:  require.NotNil,
		}, {
			name:              "decrypt offer ID",
			expectedMsg:       "retry",
			expectedStatus:    http.StatusInternalServerError,
			request:           models.HTTPTransferRequest{OfferID: "VALID"},
			authDecryptErr:    errors.New("decrypt offer id"),
			authDecryptTimes:  1,
			redisGetData:      validOffer,
			redisGetErr:       nil,
			redisGetTimes:     0,
			redisDelErr:       nil,
			redisDelTimes:     0,
			internalXferErr:   nil,
			internalXferTimes: 0,
			expectErr:         require.Error,
			expectNilResponse: require.Nil,
			expectNilPayload:  require.Nil,
		}, {
			name:              "cache unknown error",
			expectedMsg:       "retry",
			expectedStatus:    http.StatusInternalServerError,
			request:           models.HTTPTransferRequest{OfferID: "VALID"},
			authDecryptErr:    nil,
			authDecryptTimes:  1,
			redisGetData:      validOffer,
			redisGetErr:       errors.New("unknown error"),
			redisGetTimes:     1,
			redisDelErr:       nil,
			redisDelTimes:     0,
			internalXferErr:   nil,
			internalXferTimes: 0,
			expectErr:         require.Error,
			expectNilResponse: require.Nil,
			expectNilPayload:  require.Nil,
		}, {
			name:              "cache expired",
			expectedMsg:       "expired",
			expectedStatus:    http.StatusRequestTimeout,
			request:           models.HTTPTransferRequest{OfferID: "VALID"},
			authDecryptErr:    nil,
			authDecryptTimes:  1,
			redisGetData:      validOffer,
			redisGetErr:       redis.ErrCacheMiss,
			redisGetTimes:     1,
			redisDelErr:       nil,
			redisDelTimes:     0,
			internalXferErr:   nil,
			internalXferTimes: 0,
			expectErr:         require.Error,
			expectNilResponse: require.Nil,
			expectNilPayload:  require.Nil,
		}, {
			name:              "cache del failure",
			expectedMsg:       "retry",
			expectedStatus:    http.StatusInternalServerError,
			request:           models.HTTPTransferRequest{OfferID: "VALID"},
			authDecryptErr:    nil,
			authDecryptTimes:  1,
			redisGetData:      validOffer,
			redisGetErr:       nil,
			redisGetTimes:     1,
			redisDelErr:       redis.ErrCacheUnknown,
			redisDelTimes:     1,
			internalXferErr:   nil,
			internalXferTimes: 0,
			expectErr:         require.Error,
			expectNilResponse: require.Nil,
			expectNilPayload:  require.Nil,
		}, {
			name:              "cache del expired",
			expectedMsg:       "",
			expectedStatus:    0,
			request:           models.HTTPTransferRequest{OfferID: "VALID"},
			authDecryptErr:    nil,
			authDecryptTimes:  1,
			redisGetData:      validOffer,
			redisGetErr:       nil,
			redisGetTimes:     1,
			redisDelErr:       redis.ErrCacheMiss,
			redisDelTimes:     1,
			internalXferErr:   nil,
			internalXferTimes: 1,
			expectErr:         require.NoError,
			expectNilResponse: require.NotNil,
			expectNilPayload:  require.Nil,
		}, {
			name:              "client id mismatch",
			expectedMsg:       "retry",
			expectedStatus:    http.StatusInternalServerError,
			request:           models.HTTPTransferRequest{OfferID: "VALID"},
			authDecryptErr:    nil,
			authDecryptTimes:  1,
			redisGetData:      invalidOfferClientID,
			redisGetErr:       nil,
			redisGetTimes:     1,
			redisDelErr:       nil,
			redisDelTimes:     1,
			internalXferErr:   nil,
			internalXferTimes: 0,
			expectErr:         require.Error,
			expectNilResponse: require.Nil,
			expectNilPayload:  require.Nil,
		}, {
			name:             "crypto purchase",
			expectedMsg:      "invalid Fiat currency",
			expectedStatus:   http.StatusBadRequest,
			request:          models.HTTPTransferRequest{OfferID: "VALID"},
			authDecryptErr:   nil,
			authDecryptTimes: 1,
			redisGetData: models.HTTPExchangeOfferResponse{
				PriceQuote:       validOffer.PriceQuote,
				IsCryptoPurchase: true,
				IsCryptoSale:     false,
			},
			redisGetErr:       nil,
			redisGetTimes:     1,
			redisDelErr:       nil,
			redisDelTimes:     1,
			internalXferErr:   nil,
			internalXferTimes: 0,
			expectErr:         require.Error,
			expectNilResponse: require.Nil,
			expectNilPayload:  require.Nil,
		}, {
			name:             "crypto sale",
			expectedMsg:      "invalid Fiat currency",
			expectedStatus:   http.StatusBadRequest,
			request:          models.HTTPTransferRequest{OfferID: "VALID"},
			authDecryptErr:   nil,
			authDecryptTimes: 1,
			redisGetData: models.HTTPExchangeOfferResponse{
				PriceQuote:       validOffer.PriceQuote,
				IsCryptoPurchase: false,
				IsCryptoSale:     true,
			},
			redisGetErr:       nil,
			redisGetTimes:     1,
			redisDelErr:       nil,
			redisDelTimes:     1,
			internalXferErr:   nil,
			internalXferTimes: 0,
			expectErr:         require.Error,
			expectNilResponse: require.Nil,
			expectNilPayload:  require.Nil,
		}, {
			name:              "invalid source destination amount",
			expectedMsg:       "invalid",
			expectedStatus:    http.StatusBadRequest,
			request:           models.HTTPTransferRequest{OfferID: "VALID"},
			authDecryptErr:    nil,
			authDecryptTimes:  1,
			redisGetData:      invalidOfferSource,
			redisGetErr:       nil,
			redisGetTimes:     1,
			redisDelErr:       nil,
			redisDelTimes:     1,
			internalXferErr:   nil,
			internalXferTimes: 0,
			expectErr:         require.Error,
			expectNilResponse: require.Nil,
			expectNilPayload:  require.Nil,
		}, {
			name:              "transaction failure",
			expectedMsg:       "both currency accounts and enough funds",
			expectedStatus:    http.StatusBadRequest,
			request:           models.HTTPTransferRequest{OfferID: "VALID"},
			authDecryptErr:    nil,
			authDecryptTimes:  1,
			redisGetData:      validOffer,
			redisGetErr:       nil,
			redisGetTimes:     1,
			redisDelErr:       nil,
			redisDelTimes:     1,
			internalXferErr:   errors.New("transaction failure"),
			internalXferTimes: 1,
			expectErr:         require.Error,
			expectNilResponse: require.Nil,
			expectNilPayload:  require.Nil,
		}, {
			name:              "valid",
			expectedMsg:       "",
			expectedStatus:    0,
			request:           models.HTTPTransferRequest{OfferID: "VALID"},
			authDecryptErr:    nil,
			authDecryptTimes:  1,
			redisGetData:      validOffer,
			redisGetErr:       nil,
			redisGetTimes:     1,
			redisDelErr:       nil,
			redisDelTimes:     1,
			internalXferErr:   nil,
			internalXferTimes: 1,
			expectErr:         require.NoError,
			expectNilResponse: require.NotNil,
			expectNilPayload:  require.Nil,
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
			mockDB := mocks.NewMockPostgres(mockCtrl)

			gomock.InOrder(
				mockAuth.EXPECT().DecryptFromString(gomock.Any()).
					Return(validOfferID, test.authDecryptErr).
					Times(test.authDecryptTimes),

				mockCache.EXPECT().Get(gomock.Any(), gomock.Any()).
					Return(test.redisGetErr).
					SetArg(1, test.redisGetData).
					Times(test.redisGetTimes),

				mockCache.EXPECT().Del(gomock.Any()).
					Return(test.redisDelErr).
					Times(test.redisDelTimes),

				mockDB.EXPECT().FiatInternalTransfer(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil, test.internalXferErr).
					Times(test.internalXferTimes),
			)

			response, httpStatus, httpMessage, payload, err :=
				HTTPFiatTransfer(mockAuth, mockCache, mockDB, zapLogger, validClientID, &test.request)
			test.expectErr(t, err, "error expectation failed.")
			test.expectNilResponse(t, response, "nil response expectation failed.")
			test.expectNilPayload(t, payload, "nil payload expectation failed.")
			require.Equal(t, test.expectedStatus, httpStatus, "expected http status mismatched.")
			require.Contains(t, httpMessage, test.expectedMsg, "expected message mismatched.")
		})
	}
}

func TestUtilities_HTTPFiatBalance(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		ticker              string
		expectedMsg         string
		expectedStatus      int
		fiatBalanceErr      error
		fiatBalanceTimes    int
		expectErr           require.ErrorAssertionFunc
		expectNilAccDetails require.ValueAssertionFunc
		expectNilPayload    require.ValueAssertionFunc
	}{
		{
			name:                "invalid currency",
			ticker:              "INVALID",
			expectedMsg:         constants.GetInvalidCurrencyString(),
			expectedStatus:      http.StatusBadRequest,
			fiatBalanceErr:      nil,
			fiatBalanceTimes:    0,
			expectErr:           require.Error,
			expectNilAccDetails: require.Nil,
			expectNilPayload:    require.NotNil,
		}, {
			name:                "unknown db error",
			ticker:              "AED",
			expectedMsg:         "retry",
			expectedStatus:      http.StatusInternalServerError,
			fiatBalanceErr:      errors.New("unknown error"),
			fiatBalanceTimes:    1,
			expectErr:           require.Error,
			expectNilAccDetails: require.Nil,
			expectNilPayload:    require.Nil,
		}, {
			name:                "known db error",
			ticker:              "CAD",
			expectedMsg:         "records not found",
			expectedStatus:      http.StatusNotFound,
			fiatBalanceErr:      postgres.ErrNotFound,
			fiatBalanceTimes:    1,
			expectErr:           require.Error,
			expectNilAccDetails: require.Nil,
			expectNilPayload:    require.Nil,
		}, {
			name:                "valid",
			ticker:              "USD",
			expectedMsg:         "",
			expectedStatus:      0,
			fiatBalanceErr:      nil,
			fiatBalanceTimes:    1,
			expectErr:           require.NoError,
			expectNilAccDetails: require.NotNil,
			expectNilPayload:    require.Nil,
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

			mockDB.EXPECT().FiatBalance(gomock.Any(), gomock.Any()).
				Return(postgres.FiatAccount{}, test.fiatBalanceErr).
				Times(test.fiatBalanceTimes)

			accDetails, httpStatus, httpMessage, payload, err :=
				HTTPFiatBalance(mockDB, zapLogger, uuid.UUID{}, test.ticker)
			test.expectErr(t, err, "error expectation failed.")
			test.expectNilAccDetails(t, accDetails, "nil account details expectation failed.")
			test.expectNilPayload(t, payload, "nil payload expectation failed.")
			require.Equal(t, test.expectedStatus, httpStatus, "http status mismatched.")
			require.Contains(t, httpMessage, test.expectedMsg, "http message mismatched.")
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

func TestUtilities_HTTPFiatBalancePaginated(t *testing.T) {
	t.Parallel()

	accDetails := []postgres.FiatAccount{{}, {}, {}, {}}

	testCases := []struct {
		name                 string
		pageCursor           string
		pageSize             string
		expectedMsg          string
		expectedStatus       int
		accDetails           []postgres.FiatAccount
		authDecryptStrErr    error
		authDecryptStrTimes  int
		fiatBalanceErr       error
		fiatBalanceTimes     int
		authEncryptStrErr    error
		authEncryptStrTimes  int
		expectErr            require.ErrorAssertionFunc
		expectNilFiatAccount require.ValueAssertionFunc
	}{
		{
			name:                 "decrypt cursor failure",
			pageCursor:           "PaGeCuRs0R==",
			pageSize:             "3",
			expectedMsg:          "invalid page cursor or page size",
			accDetails:           accDetails,
			expectedStatus:       http.StatusBadRequest,
			authDecryptStrErr:    errors.New("decrypt failure"),
			authDecryptStrTimes:  1,
			fiatBalanceErr:       nil,
			fiatBalanceTimes:     0,
			authEncryptStrErr:    nil,
			authEncryptStrTimes:  0,
			expectErr:            require.Error,
			expectNilFiatAccount: require.Nil,
		}, {
			name:                 "known db error",
			pageCursor:           "PaGeCuRs0R==",
			pageSize:             "3",
			expectedMsg:          "not found",
			accDetails:           accDetails,
			expectedStatus:       http.StatusNotFound,
			authDecryptStrErr:    nil,
			authDecryptStrTimes:  1,
			fiatBalanceErr:       postgres.ErrNotFound,
			fiatBalanceTimes:     1,
			authEncryptStrErr:    nil,
			authEncryptStrTimes:  0,
			expectErr:            require.Error,
			expectNilFiatAccount: require.Nil,
		}, {
			name:                 "unknown db error",
			pageCursor:           "PaGeCuRs0R==",
			pageSize:             "3",
			expectedMsg:          "retry",
			accDetails:           accDetails,
			expectedStatus:       http.StatusInternalServerError,
			authDecryptStrErr:    nil,
			authDecryptStrTimes:  1,
			fiatBalanceErr:       errors.New("unknown db error"),
			fiatBalanceTimes:     1,
			authEncryptStrErr:    nil,
			authEncryptStrTimes:  0,
			expectErr:            require.Error,
			expectNilFiatAccount: require.Nil,
		}, {
			name:                 "encrypt cursor failure",
			pageCursor:           "PaGeCuRs0R==",
			pageSize:             "3",
			expectedMsg:          "retry",
			accDetails:           accDetails,
			expectedStatus:       http.StatusInternalServerError,
			authDecryptStrErr:    nil,
			authDecryptStrTimes:  1,
			fiatBalanceErr:       nil,
			fiatBalanceTimes:     1,
			authEncryptStrErr:    errors.New("encrypt string error"),
			authEncryptStrTimes:  1,
			expectErr:            require.Error,
			expectNilFiatAccount: require.Nil,
		}, {
			name:                 "valid without query and 10 records",
			pageCursor:           "",
			pageSize:             "",
			expectedMsg:          "",
			accDetails:           []postgres.FiatAccount{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}},
			expectedStatus:       0,
			authDecryptStrErr:    nil,
			authDecryptStrTimes:  0,
			fiatBalanceErr:       nil,
			fiatBalanceTimes:     1,
			authEncryptStrErr:    nil,
			authEncryptStrTimes:  0,
			expectErr:            require.NoError,
			expectNilFiatAccount: require.NotNil,
		}, {
			name:                 "valid without query and 11 records",
			pageCursor:           "",
			pageSize:             "",
			expectedMsg:          "",
			accDetails:           []postgres.FiatAccount{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}},
			expectedStatus:       0,
			authDecryptStrErr:    nil,
			authDecryptStrTimes:  0,
			fiatBalanceErr:       nil,
			fiatBalanceTimes:     1,
			authEncryptStrErr:    nil,
			authEncryptStrTimes:  1,
			expectErr:            require.NoError,
			expectNilFiatAccount: require.NotNil,
		}, {
			name:                 "valid without query",
			pageCursor:           "",
			pageSize:             "",
			expectedMsg:          "",
			accDetails:           accDetails,
			expectedStatus:       0,
			authDecryptStrErr:    nil,
			authDecryptStrTimes:  0,
			fiatBalanceErr:       nil,
			fiatBalanceTimes:     1,
			authEncryptStrErr:    nil,
			authEncryptStrTimes:  0,
			expectErr:            require.NoError,
			expectNilFiatAccount: require.NotNil,
		}, {
			name:                 "valid with query",
			pageCursor:           "PaGeCuRs0R==",
			pageSize:             "3",
			expectedMsg:          "",
			accDetails:           accDetails,
			expectedStatus:       0,
			authDecryptStrErr:    nil,
			authDecryptStrTimes:  1,
			fiatBalanceErr:       nil,
			fiatBalanceTimes:     1,
			authEncryptStrErr:    nil,
			authEncryptStrTimes:  1,
			expectErr:            require.NoError,
			expectNilFiatAccount: require.NotNil,
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
			mockDB := mocks.NewMockPostgres(mockCtrl)

			gomock.InOrder(
				mockAuth.EXPECT().DecryptFromString(gomock.Any()).
					Return([]byte{}, test.authDecryptStrErr).
					Times(test.authDecryptStrTimes),

				mockDB.EXPECT().FiatBalancePaginated(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(test.accDetails, test.fiatBalanceErr).
					Times(test.fiatBalanceTimes),

				mockAuth.EXPECT().EncryptToString(gomock.Any()).
					Return("encrypted-page-cursor", test.authEncryptStrErr).
					Times(test.authEncryptStrTimes),
			)

			fiatAccount, httpStatus, httpMessage, err :=
				HTTPFiatBalancePaginated(mockAuth, mockDB, zapLogger, uuid.UUID{}, test.pageCursor, test.pageSize, true)
			test.expectErr(t, err, "error expectation failed.")
			test.expectNilFiatAccount(t, fiatAccount, "nil fiat account expectation failed.")
			require.Equal(t, test.expectedStatus, httpStatus, "http status mismatched.")
			require.Contains(t, httpMessage, test.expectedMsg, "http message mismatched.")
		})
	}
}

func TestHandler_TxDetailsFiatPaginated(t *testing.T) {
	t.Parallel()

	decryptedCursor := fmt.Sprintf("%s,%s,%d",
		fmt.Sprintf(constants.GetMonthFormatString(), 2023, 6, "-04:00"),
		fmt.Sprintf(constants.GetMonthFormatString(), 2023, 7, "-04:00"),
		10)

	journalEntries := []postgres.FiatJournal{{}, {}, {}, {}}

	paramsCursor := &HTTPPaginatedTxParams{PageCursorStr: "page-cursor"}

	testCases := []struct {
		name                   string
		ticker                 string
		expectedMsg            string
		expectedStatus         int
		params                 *HTTPPaginatedTxParams
		journalEntries         []postgres.FiatJournal
		authDecryptCursorErr   error
		authDecryptCursorTimes int
		authEncryptCursorErr   error
		authEncryptCursorTimes int
		fiatTxPaginatedErr     error
		fiatTxPaginatedTimes   int
		expectErr              require.ErrorAssertionFunc
		expectNilJournal       require.ValueAssertionFunc
		expectNilPayload       require.ValueAssertionFunc
	}{
		{
			name:                   "bad currency",
			ticker:                 "INVALID",
			params:                 paramsCursor,
			expectedMsg:            constants.GetInvalidCurrencyString(),
			journalEntries:         journalEntries,
			expectedStatus:         http.StatusBadRequest,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 0,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 0,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   0,
			expectErr:              require.Error,
			expectNilJournal:       require.Nil,
			expectNilPayload:       require.NotNil,
		}, {
			name:                   "no cursor or params",
			ticker:                 "USD",
			params:                 &HTTPPaginatedTxParams{},
			expectedMsg:            "missing required parameters",
			journalEntries:         journalEntries,
			expectedStatus:         http.StatusBadRequest,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 0,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 0,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   0,
			expectErr:              require.Error,
			expectNilJournal:       require.Nil,
			expectNilPayload:       require.Nil,
		}, {
			name:                   "db failure",
			ticker:                 "USD",
			params:                 paramsCursor,
			expectedMsg:            "records not found",
			journalEntries:         journalEntries,
			expectedStatus:         http.StatusNotFound,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 1,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     postgres.ErrNotFound,
			fiatTxPaginatedTimes:   1,
			expectErr:              require.Error,
			expectNilJournal:       require.Nil,
			expectNilPayload:       require.Nil,
		}, {
			name:                   "unknown db failure",
			ticker:                 "USD",
			params:                 paramsCursor,
			expectedMsg:            "retry",
			journalEntries:         journalEntries,
			expectedStatus:         http.StatusInternalServerError,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 1,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     errors.New("db failure"),
			fiatTxPaginatedTimes:   1,
			expectErr:              require.Error,
			expectNilJournal:       require.Nil,
			expectNilPayload:       require.Nil,
		}, {
			name:                   "no transactions",
			ticker:                 "USD",
			params:                 paramsCursor,
			expectedMsg:            "no transactions",
			journalEntries:         []postgres.FiatJournal{},
			expectedStatus:         http.StatusRequestedRangeNotSatisfiable,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 1,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   1,
			expectErr:              require.Error,
			expectNilJournal:       require.Nil,
			expectNilPayload:       require.Nil,
		}, {
			name:                   "valid with cursor",
			ticker:                 "USD",
			params:                 paramsCursor,
			expectedMsg:            "",
			journalEntries:         journalEntries,
			expectedStatus:         0,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 1,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   1,
			expectErr:              require.NoError,
			expectNilJournal:       require.NotNil,
			expectNilPayload:       require.Nil,
		}, {
			name:   "valid with query",
			ticker: "USD",
			params: &HTTPPaginatedTxParams{
				PageSizeStr: "3",
				TimezoneStr: "-04:00",
				MonthStr:    "6",
				YearStr:     "2023",
			},
			expectedMsg:            "",
			journalEntries:         journalEntries,
			expectedStatus:         0,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 0,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   1,
			expectErr:              require.NoError,
			expectNilJournal:       require.NotNil,
			expectNilPayload:       require.Nil,
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
			mockDB := mocks.NewMockPostgres(mockCtrl)

			gomock.InOrder(
				mockAuth.EXPECT().DecryptFromString(gomock.Any()).
					Return([]byte(decryptedCursor), test.authDecryptCursorErr).
					Times(test.authDecryptCursorTimes),

				mockAuth.EXPECT().EncryptToString(gomock.Any()).
					Return("encrypted-cursor", test.authEncryptCursorErr).
					Times(test.authEncryptCursorTimes),

				mockDB.EXPECT().FiatTransactionsPaginated(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any()).
					Return(test.journalEntries, test.fiatTxPaginatedErr).
					Times(test.fiatTxPaginatedTimes),
			)

			actualEntries, httpStatus, httpMessage, payload, err :=
				HTTPFiatTransactionsPaginated(mockAuth, mockDB, zapLogger, uuid.UUID{}, test.ticker, test.params, true)
			test.expectErr(t, err, "error expectation failed.")
			test.expectNilJournal(t, actualEntries, "nil journal expectation failed.")
			test.expectNilPayload(t, payload, "nil payload expectation failed.")
			require.Equal(t, test.expectedStatus, httpStatus, "http status codes mismatched.")
			require.Contains(t, httpMessage, test.expectedMsg, "http message mismatched.")
		})
	}
}
