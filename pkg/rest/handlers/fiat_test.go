package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/quotes"
	"github.com/surahman/FTeX/pkg/redis"
)

func TestHandlers_OpenFiat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		path               string
		expectedStatus     int
		request            *models.HTTPOpenCurrencyAccountRequest
		authValidateJWTErr error
		authValidateTimes  int
		fiatCreateAccErr   error
		fiatCreateAccTimes int
	}{
		{
			name:               "valid",
			path:               "/open/valid",
			expectedStatus:     http.StatusCreated,
			request:            &models.HTTPOpenCurrencyAccountRequest{Currency: "USD"},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			fiatCreateAccErr:   nil,
			fiatCreateAccTimes: 1,
		}, {
			name:               "validation",
			path:               "/open/validation",
			expectedStatus:     http.StatusBadRequest,
			request:            &models.HTTPOpenCurrencyAccountRequest{},
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			fiatCreateAccErr:   nil,
			fiatCreateAccTimes: 0,
		}, {
			name:               "invalid currency",
			path:               "/open/invalid-currency",
			expectedStatus:     http.StatusBadRequest,
			request:            &models.HTTPOpenCurrencyAccountRequest{Currency: "UVW"},
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			fiatCreateAccErr:   nil,
			fiatCreateAccTimes: 0,
		}, {
			name:               "invalid jwt",
			path:               "/open/invalid-jwt",
			expectedStatus:     http.StatusForbidden,
			request:            &models.HTTPOpenCurrencyAccountRequest{Currency: "USD"},
			authValidateJWTErr: errors.New("invalid jwt"),
			authValidateTimes:  1,
			fiatCreateAccErr:   nil,
			fiatCreateAccTimes: 0,
		}, {
			name:               "db failure",
			path:               "/open/db-failure",
			expectedStatus:     http.StatusConflict,
			request:            &models.HTTPOpenCurrencyAccountRequest{Currency: "USD"},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			fiatCreateAccErr:   postgres.ErrCreateFiat,
			fiatCreateAccTimes: 1,
		}, {
			name:               "db failure unknown",
			path:               "/open/db-failure-unknown",
			expectedStatus:     http.StatusInternalServerError,
			request:            &models.HTTPOpenCurrencyAccountRequest{Currency: "USD"},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			fiatCreateAccErr:   errors.New("unknown server error"),
			fiatCreateAccTimes: 1,
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
			mockPostgres := mocks.NewMockPostgres(mockCtrl)

			openReqJSON, err := json.Marshal(&test.request)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

			gomock.InOrder(
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(uuid.UUID{}, int64(0), test.authValidateJWTErr).
					Times(test.authValidateTimes),

				mockPostgres.EXPECT().FiatCreateAccount(gomock.Any(), gomock.Any()).
					Return(test.fiatCreateAccErr).
					Times(test.fiatCreateAccTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.POST(test.path, OpenFiat(zapLogger, mockAuth, mockPostgres, "Authorization"))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path, bytes.NewBuffer(openReqJSON))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, w.Code, "expected status codes do not match")
		})
	}
}

func TestHandlers_ValidateSourceDestinationAmount(t *testing.T) {
	t.Parallel()

	amountValid, err := decimal.NewFromString("10101.11")
	require.NoError(t, err, "failed to parse valid amount.")

	amountInvalidNegative, err := decimal.NewFromString("-10101.11")
	require.NoError(t, err, "failed to parse invalid negative amount")

	amountInvalidDecimal, err := decimal.NewFromString("10101.111")
	require.NoError(t, err, "failed to parse invalid decimal amount")

	testCases := []struct {
		name         string
		expectErrMsg string
		srcCurrency  string
		dstCurrency  string
		amount       decimal.Decimal
		expectErr    require.ErrorAssertionFunc
	}{
		{
			name:         "valid",
			expectErrMsg: "",
			srcCurrency:  "USD",
			dstCurrency:  "CAD",
			amount:       amountValid,
			expectErr:    require.NoError,
		}, {
			name:         "invalid source currency",
			expectErrMsg: "source currency",
			srcCurrency:  "INVALID",
			dstCurrency:  "CAD",
			amount:       amountValid,
			expectErr:    require.Error,
		}, {
			name:         "invalid destination currency",
			expectErrMsg: "destination currency",
			srcCurrency:  "USD",
			dstCurrency:  "INVALID",
			amount:       amountValid,
			expectErr:    require.Error,
		}, {
			name:         "invalid negative amount",
			expectErrMsg: "source amount",
			srcCurrency:  "USD",
			dstCurrency:  "CAD",
			amount:       amountInvalidNegative,
			expectErr:    require.Error,
		}, {
			name:         "invalid decimal amount",
			expectErrMsg: "source amount",
			srcCurrency:  "USD",
			dstCurrency:  "CAD",
			amount:       amountInvalidDecimal,
			expectErr:    require.Error,
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			src, dst, err := validateSourceDestinationAmount(test.srcCurrency, test.dstCurrency, test.amount)
			test.expectErr(t, err, "error expectation failed.")

			if err != nil {
				require.Contains(t, err.Error(), test.expectErrMsg, "error message is incorrect.")

				return
			}

			require.Equal(t, src, postgres.Currency(test.srcCurrency), "source currency mismatched.")
			require.Equal(t, dst, postgres.Currency(test.dstCurrency), "destination currency mismatched.")
		})
	}
}

func TestHandlers_DepositFiat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		expectedMsg        string
		path               string
		expectedStatus     int
		request            *models.HTTPDepositCurrencyRequest
		authValidateJWTErr error
		authValidateTimes  int
		extTransferErr     error
		extTransferTimes   int
	}{
		{
			name:               "empty request",
			expectedMsg:        "validation",
			path:               "/fiat-deposit/empty-request",
			expectedStatus:     http.StatusBadRequest,
			request:            &models.HTTPDepositCurrencyRequest{},
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			extTransferErr:     nil,
			extTransferTimes:   0,
		}, {
			name:               "invalid currency",
			expectedMsg:        "currency",
			path:               "/fiat-deposit/invalid-currency",
			expectedStatus:     http.StatusBadRequest,
			request:            &models.HTTPDepositCurrencyRequest{Currency: "INVALID", Amount: decimal.NewFromFloat(1)},
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			extTransferErr:     nil,
			extTransferTimes:   0,
		}, {
			name:               "too many decimal places",
			expectedMsg:        "amount",
			path:               "/fiat-deposit/too-many-decimal-places",
			expectedStatus:     http.StatusBadRequest,
			request:            &models.HTTPDepositCurrencyRequest{Currency: "USD", Amount: decimal.NewFromFloat(1.234)},
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			extTransferErr:     nil,
			extTransferTimes:   0,
		}, {
			name:               "negative",
			expectedMsg:        "amount",
			path:               "/fiat-deposit/negative",
			expectedStatus:     http.StatusBadRequest,
			request:            &models.HTTPDepositCurrencyRequest{Currency: "USD", Amount: decimal.NewFromFloat(-1)},
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			extTransferErr:     nil,
			extTransferTimes:   0,
		}, {
			name:               "invalid jwt",
			expectedMsg:        "invalid jwt",
			path:               "/fiat-deposit/invalid-jwt",
			expectedStatus:     http.StatusForbidden,
			request:            &models.HTTPDepositCurrencyRequest{Currency: "USD", Amount: decimal.NewFromFloat(1337.89)},
			authValidateJWTErr: errors.New("invalid jwt"),
			authValidateTimes:  1,
			extTransferErr:     nil,
			extTransferTimes:   0,
		}, {
			name:               "unknown xfer error",
			expectedMsg:        "retry",
			path:               "/fiat-deposit/unknown-xfer-error",
			expectedStatus:     http.StatusInternalServerError,
			request:            &models.HTTPDepositCurrencyRequest{Currency: "USD", Amount: decimal.NewFromFloat(1337.89)},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			extTransferErr:     errors.New("unknown error"),
			extTransferTimes:   1,
		}, {
			name:               "xfer error",
			expectedMsg:        "could not complete",
			path:               "/fiat-deposit/xfer-error",
			expectedStatus:     http.StatusInternalServerError,
			request:            &models.HTTPDepositCurrencyRequest{Currency: "USD", Amount: decimal.NewFromFloat(1337.89)},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			extTransferErr:     postgres.ErrTransactFiat,
			extTransferTimes:   1,
		}, {
			name:               "valid",
			expectedMsg:        "successfully",
			path:               "/fiat-deposit/valid",
			expectedStatus:     http.StatusOK,
			request:            &models.HTTPDepositCurrencyRequest{Currency: "USD", Amount: decimal.NewFromFloat(1337.89)},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			extTransferErr:     nil,
			extTransferTimes:   1,
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
			mockPostgres := mocks.NewMockPostgres(mockCtrl)

			depositReqJSON, err := json.Marshal(&test.request)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

			gomock.InOrder(
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(uuid.UUID{}, int64(0), test.authValidateJWTErr).
					Times(test.authValidateTimes),

				mockPostgres.EXPECT().FiatExternalTransfer(gomock.Any(), gomock.Any()).
					Return(&postgres.FiatAccountTransferResult{}, test.extTransferErr).
					Times(test.extTransferTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.POST(test.path, DepositFiat(zapLogger, mockAuth, mockPostgres, "Authorization"))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path, bytes.NewBuffer(depositReqJSON))
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, recorder.Code, "expected status codes do not match")

			var resp map[string]interface{}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp), "failed to unpack success response.")

			errorMessage, ok := resp["message"].(string)
			require.True(t, ok, "failed to extract response message.")
			require.Contains(t, errorMessage, test.expectedMsg, "incorrect response message.")
		})
	}
}

func TestHandlers_ExchangeOfferFiat(t *testing.T) { //nolint:maintidx
	t.Parallel()

	amountValid, err := decimal.NewFromString("999")
	require.NoError(t, err, "failed to convert valid amount")

	amountInvalidDecimal, err := decimal.NewFromString("999.999")
	require.NoError(t, err, "failed to convert invalid decimal amount")

	amountInvalidNegative, err := decimal.NewFromString("-999")
	require.NoError(t, err, "failed to convert invalid negative amount")

	testCases := []struct {
		name               string
		expectedMsg        string
		path               string
		expectedStatus     int
		request            *models.HTTPFiatExchangeOfferRequest
		authValidateJWTErr error
		authValidateTimes  int
		quotesErr          error
		quotesTimes        int
		authEncryptErr     error
		authEncryptTimes   int
		redisErr           error
		redisTimes         int
	}{
		{
			name:               "empty request",
			expectedMsg:        "validation",
			path:               "/exchange-offer-fiat/empty-request",
			expectedStatus:     http.StatusBadRequest,
			request:            &models.HTTPFiatExchangeOfferRequest{},
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			quotesErr:          nil,
			quotesTimes:        0,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:           "invalid source currency",
			expectedMsg:    "source currency",
			path:           "/exchange-offer-fiat/invalid-src-currency",
			expectedStatus: http.StatusBadRequest,
			request: &models.HTTPFiatExchangeOfferRequest{
				SourceCurrency:      "INVALID",
				DestinationCurrency: "USD",
				SourceAmount:        amountValid,
			},
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			quotesErr:          nil,
			quotesTimes:        0,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:           "invalid destination currency",
			expectedMsg:    "destination currency",
			path:           "/exchange-offer-fiat/invalid-dst-currency",
			expectedStatus: http.StatusBadRequest,
			request: &models.HTTPFiatExchangeOfferRequest{
				SourceCurrency:      "USD",
				DestinationCurrency: "INVALID",
				SourceAmount:        amountValid,
			},
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			quotesErr:          nil,
			quotesTimes:        0,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:           "too many decimal places",
			expectedMsg:    "amount",
			path:           "/exchange-offer-fiat/too-many-decimal-places",
			expectedStatus: http.StatusBadRequest,
			request: &models.HTTPFiatExchangeOfferRequest{
				SourceCurrency:      "USD",
				DestinationCurrency: "USD",
				SourceAmount:        amountInvalidDecimal,
			},
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			quotesErr:          nil,
			quotesTimes:        0,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:           "negative",
			expectedMsg:    "amount",
			path:           "/exchange-offer-fiat/negative",
			expectedStatus: http.StatusBadRequest,
			request: &models.HTTPFiatExchangeOfferRequest{
				SourceCurrency:      "USD",
				DestinationCurrency: "USD",
				SourceAmount:        amountInvalidNegative,
			},
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			quotesErr:          nil,
			quotesTimes:        0,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:           "invalid jwt",
			expectedMsg:    "invalid jwt",
			path:           "/exchange-offer-fiat/invalid-jwt",
			expectedStatus: http.StatusForbidden,
			request: &models.HTTPFiatExchangeOfferRequest{
				SourceCurrency:      "USD",
				DestinationCurrency: "USD",
				SourceAmount:        amountValid,
			},
			authValidateJWTErr: errors.New("invalid jwt"),
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesTimes:        0,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:           "fiat conversion error",
			expectedMsg:    "retry",
			path:           "/exchange-offer-fiat/fiat-conversion-error",
			expectedStatus: http.StatusInternalServerError,
			request: &models.HTTPFiatExchangeOfferRequest{
				SourceCurrency:      "USD",
				DestinationCurrency: "USD",
				SourceAmount:        amountValid,
			},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          errors.New(""),
			quotesTimes:        1,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:           "encryption error",
			expectedMsg:    "retry",
			path:           "/exchange-offer-fiat/encryption-error",
			expectedStatus: http.StatusInternalServerError,
			request: &models.HTTPFiatExchangeOfferRequest{
				SourceCurrency:      "USD",
				DestinationCurrency: "USD",
				SourceAmount:        amountValid,
			},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesTimes:        1,
			authEncryptErr:     errors.New(""),
			authEncryptTimes:   1,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:           "redis error",
			expectedMsg:    "retry",
			path:           "/exchange-offer-fiat/redis-error",
			expectedStatus: http.StatusInternalServerError,
			request: &models.HTTPFiatExchangeOfferRequest{
				SourceCurrency:      "USD",
				DestinationCurrency: "USD",
				SourceAmount:        amountValid,
			},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesTimes:        1,
			authEncryptErr:     nil,
			authEncryptTimes:   1,
			redisErr:           errors.New(""),
			redisTimes:         1,
		}, {
			name:           "valid",
			expectedMsg:    "",
			path:           "/exchange-offer-fiat/valid",
			expectedStatus: http.StatusOK,
			request: &models.HTTPFiatExchangeOfferRequest{
				SourceCurrency:      "USD",
				DestinationCurrency: "USD",
				SourceAmount:        amountValid,
			},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesTimes:        1,
			authEncryptErr:     nil,
			authEncryptTimes:   1,
			redisErr:           nil,
			redisTimes:         1,
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

			offerReqJSON, err := json.Marshal(&test.request)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

			gomock.InOrder(
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(uuid.UUID{}, int64(0), test.authValidateJWTErr).
					Times(test.authValidateTimes),

				mockQuotes.EXPECT().FiatConversion(gomock.Any(), gomock.Any(), gomock.Any(), nil).
					Return(amountValid, amountValid, test.quotesErr).
					Times(test.quotesTimes),

				mockAuth.EXPECT().EncryptToString(gomock.Any()).
					Return("OFFER-ID", test.authEncryptErr).
					Times(test.authEncryptTimes),

				mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(test.redisErr).
					Times(test.redisTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.POST(test.path, ExchangeOfferFiat(zapLogger, mockAuth, mockCache, mockQuotes, "Authorization"))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path, bytes.NewBuffer(offerReqJSON))
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, recorder.Code, "expected status codes do not match")

			var resp map[string]interface{}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp), "failed to unpack response.")

			errorMessage, ok := resp["message"].(string)
			require.True(t, ok, "failed to extract response message.")

			// Check for invalid currency codes and amount.
			if errorMessage == "invalid request" {
				payload, ok := resp["payload"].(string)
				require.True(t, ok, "failed to extract payload from response.")
				require.Contains(t, payload, test.expectedMsg)
			} else {
				require.Contains(t, errorMessage, test.expectedMsg, "incorrect response message.")
			}
		})
	}
}

func TestHandlers_GetCachedOffer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		expectErrMsg  string
		expectStatus  int
		expectErr     require.ErrorAssertionFunc
		redisGetErr   error
		redisGetData  models.HTTPFiatExchangeOfferResponse
		redisGetTimes int
		redisDelErr   error
		redisDelTimes int
	}{
		{
			name:          "get unknown error",
			expectErrMsg:  "retry",
			expectStatus:  http.StatusInternalServerError,
			expectErr:     require.Error,
			redisGetErr:   errors.New("unknown error"),
			redisGetData:  models.HTTPFiatExchangeOfferResponse{},
			redisGetTimes: 1,
			redisDelErr:   nil,
			redisDelTimes: 0,
		}, {
			name:          "get unknown package error",
			expectErrMsg:  "retry",
			expectStatus:  http.StatusInternalServerError,
			expectErr:     require.Error,
			redisGetErr:   redis.ErrCacheUnknown,
			redisGetData:  models.HTTPFiatExchangeOfferResponse{},
			redisGetTimes: 1,
			redisDelErr:   nil,
			redisDelTimes: 0,
		}, {
			name:          "get package error",
			expectErrMsg:  "expired",
			expectStatus:  http.StatusRequestTimeout,
			expectErr:     require.Error,
			redisGetErr:   redis.ErrCacheMiss,
			redisGetData:  models.HTTPFiatExchangeOfferResponse{},
			redisGetTimes: 1,
			redisDelErr:   nil,
			redisDelTimes: 0,
		}, {
			name:          "del unknown error",
			expectErrMsg:  "retry",
			expectStatus:  http.StatusInternalServerError,
			expectErr:     require.Error,
			redisGetErr:   nil,
			redisGetData:  models.HTTPFiatExchangeOfferResponse{},
			redisGetTimes: 1,
			redisDelErr:   errors.New("unknown error"),
			redisDelTimes: 1,
		}, {
			name:          "del unknown package error",
			expectErrMsg:  "retry",
			expectStatus:  http.StatusInternalServerError,
			expectErr:     require.Error,
			redisGetErr:   nil,
			redisGetData:  models.HTTPFiatExchangeOfferResponse{},
			redisGetTimes: 1,
			redisDelErr:   redis.ErrCacheUnknown,
			redisDelTimes: 1,
		}, {
			name:          "del cache miss",
			expectErrMsg:  "",
			expectStatus:  http.StatusOK,
			expectErr:     require.NoError,
			redisGetErr:   nil,
			redisGetData:  models.HTTPFiatExchangeOfferResponse{},
			redisGetTimes: 1,
			redisDelErr:   redis.ErrCacheMiss,
			redisDelTimes: 1,
		}, {
			name:          "valid",
			expectErrMsg:  "",
			expectStatus:  http.StatusOK,
			expectErr:     require.NoError,
			redisGetErr:   nil,
			redisGetData:  models.HTTPFiatExchangeOfferResponse{},
			redisGetTimes: 1,
			redisDelErr:   nil,
			redisDelTimes: 1,
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockCache := mocks.NewMockRedis(mockCtrl)

			gomock.InOrder(
				mockCache.EXPECT().Get(gomock.Any(), gomock.Any()).
					Return(test.redisGetErr).
					SetArg(1, test.redisGetData).
					Times(test.redisGetTimes),

				mockCache.EXPECT().Del(gomock.Any()).
					Return(test.redisDelErr).
					Times(test.redisDelTimes),
			)

			_, status, msg, err := getCachedOffer(mockCache, zapLogger, "SOME-OFFER-ID")
			test.expectErr(t, err, "error expectation failed.")
			require.Equal(t, test.expectStatus, status, "expected and actual status codes did not match.")
			require.Contains(t, msg, test.expectErrMsg, "expected error message did not match.")
		})
	}
}

func TestHandler_ExchangeTransferFiat(t *testing.T) { //nolint:maintidx
	t.Parallel()

	validDecimal, err := decimal.NewFromString("10101.11")
	require.NoError(t, err, "failed to parse valid decimal.")

	validClientID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate valid client id.")

	invalidClientID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate invalid client id.")

	validOfferID := []byte("VALID")
	validOffer := models.HTTPFiatExchangeOfferResponse{
		PriceQuote: models.PriceQuote{
			ClientID:       validClientID,
			SourceAcc:      "USD",
			DestinationAcc: "CAD",
			Rate:           validDecimal,
			Amount:         validDecimal,
		},
	}

	invalidOfferClientID := models.HTTPFiatExchangeOfferResponse{
		PriceQuote: models.PriceQuote{
			ClientID:       invalidClientID,
			SourceAcc:      "USD",
			DestinationAcc: "CAD",
			Rate:           validDecimal,
			Amount:         validDecimal,
		},
	}

	invalidOfferSource := models.HTTPFiatExchangeOfferResponse{
		PriceQuote: models.PriceQuote{
			ClientID:       validClientID,
			SourceAcc:      "INVALID",
			DestinationAcc: "CAD",
			Rate:           validDecimal,
			Amount:         validDecimal,
		},
	}

	testCases := []struct {
		name               string
		path               string
		expectedMsg        string
		expectedStatus     int
		request            models.HTTPFiatTransferRequest
		authValidateJWTErr error
		authValidateTimes  int
		authDecryptErr     error
		authDecryptTimes   int
		redisGetData       models.HTTPFiatExchangeOfferResponse
		redisGetErr        error
		redisGetTimes      int
		redisDelErr        error
		redisDelTimes      int
		internalXferErr    error
		internalXferTimes  int
	}{
		{
			name:               "empty request",
			path:               "/exchange-xfer-fiat/empty-request",
			expectedMsg:        "validation",
			expectedStatus:     http.StatusBadRequest,
			request:            models.HTTPFiatTransferRequest{},
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			authDecryptErr:     nil,
			authDecryptTimes:   0,
			redisGetData:       validOffer,
			redisGetErr:        nil,
			redisGetTimes:      0,
			redisDelErr:        nil,
			redisDelTimes:      0,
			internalXferErr:    nil,
			internalXferTimes:  0,
		}, {
			name:               "invalid JWT",
			path:               "/exchange-xfer-fiat/invalid-jwt",
			expectedMsg:        "bad auth",
			expectedStatus:     http.StatusForbidden,
			request:            models.HTTPFiatTransferRequest{OfferID: "VALID"},
			authValidateJWTErr: errors.New("bad auth"),
			authValidateTimes:  1,
			authDecryptErr:     nil,
			authDecryptTimes:   0,
			redisGetData:       validOffer,
			redisGetErr:        nil,
			redisGetTimes:      0,
			redisDelErr:        nil,
			redisDelTimes:      0,
			internalXferErr:    nil,
			internalXferTimes:  0,
		}, {
			name:               "decrypt offer ID",
			path:               "/exchange-xfer-fiat/decrypt-offer-id",
			expectedMsg:        "retry",
			expectedStatus:     http.StatusInternalServerError,
			request:            models.HTTPFiatTransferRequest{OfferID: "VALID"},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			authDecryptErr:     errors.New("decrypt offer id"),
			authDecryptTimes:   1,
			redisGetData:       validOffer,
			redisGetErr:        nil,
			redisGetTimes:      0,
			redisDelErr:        nil,
			redisDelTimes:      0,
			internalXferErr:    nil,
			internalXferTimes:  0,
		}, {
			name:               "cache unknown error",
			path:               "/exchange-xfer-fiat/cache-unknown-error",
			expectedMsg:        "retry",
			expectedStatus:     http.StatusInternalServerError,
			request:            models.HTTPFiatTransferRequest{OfferID: "VALID"},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			authDecryptErr:     nil,
			authDecryptTimes:   1,
			redisGetData:       validOffer,
			redisGetErr:        errors.New("unknown error"),
			redisGetTimes:      1,
			redisDelErr:        nil,
			redisDelTimes:      0,
			internalXferErr:    nil,
			internalXferTimes:  0,
		}, {
			name:               "cache expired",
			path:               "/exchange-xfer-fiat/cache-expired",
			expectedMsg:        "expired",
			expectedStatus:     http.StatusRequestTimeout,
			request:            models.HTTPFiatTransferRequest{OfferID: "VALID"},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			authDecryptErr:     nil,
			authDecryptTimes:   1,
			redisGetData:       validOffer,
			redisGetErr:        redis.ErrCacheMiss,
			redisGetTimes:      1,
			redisDelErr:        nil,
			redisDelTimes:      0,
			internalXferErr:    nil,
			internalXferTimes:  0,
		}, {
			name:               "cache del failure",
			path:               "/exchange-xfer-fiat/cache-del-failure",
			expectedMsg:        "retry",
			expectedStatus:     http.StatusInternalServerError,
			request:            models.HTTPFiatTransferRequest{OfferID: "VALID"},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			authDecryptErr:     nil,
			authDecryptTimes:   1,
			redisGetData:       validOffer,
			redisGetErr:        nil,
			redisGetTimes:      1,
			redisDelErr:        redis.ErrCacheUnknown,
			redisDelTimes:      1,
			internalXferErr:    nil,
			internalXferTimes:  0,
		}, {
			name:               "cache del expired",
			path:               "/exchange-xfer-fiat/cache-del-expired",
			expectedMsg:        "successful",
			expectedStatus:     http.StatusOK,
			request:            models.HTTPFiatTransferRequest{OfferID: "VALID"},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			authDecryptErr:     nil,
			authDecryptTimes:   1,
			redisGetData:       validOffer,
			redisGetErr:        nil,
			redisGetTimes:      1,
			redisDelErr:        redis.ErrCacheMiss,
			redisDelTimes:      1,
			internalXferErr:    nil,
			internalXferTimes:  1,
		}, {
			name:               "client id mismatch",
			path:               "/exchange-xfer-fiat/client-id-mismatch",
			expectedMsg:        "retry",
			expectedStatus:     http.StatusInternalServerError,
			request:            models.HTTPFiatTransferRequest{OfferID: "VALID"},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			authDecryptErr:     nil,
			authDecryptTimes:   1,
			redisGetData:       invalidOfferClientID,
			redisGetErr:        nil,
			redisGetTimes:      1,
			redisDelErr:        nil,
			redisDelTimes:      1,
			internalXferErr:    nil,
			internalXferTimes:  0,
		}, {
			name:               "invalid source destination amount",
			path:               "/exchange-xfer-fiat/invalid-source-destination-amount",
			expectedMsg:        "invalid",
			expectedStatus:     http.StatusBadRequest,
			request:            models.HTTPFiatTransferRequest{OfferID: "VALID"},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			authDecryptErr:     nil,
			authDecryptTimes:   1,
			redisGetData:       invalidOfferSource,
			redisGetErr:        nil,
			redisGetTimes:      1,
			redisDelErr:        nil,
			redisDelTimes:      1,
			internalXferErr:    nil,
			internalXferTimes:  0,
		}, {
			name:               "transaction failure",
			path:               "/exchange-xfer-fiat/transaction-failure",
			expectedMsg:        "both currency accounts and enough funds",
			expectedStatus:     http.StatusBadRequest,
			request:            models.HTTPFiatTransferRequest{OfferID: "VALID"},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			authDecryptErr:     nil,
			authDecryptTimes:   1,
			redisGetData:       validOffer,
			redisGetErr:        nil,
			redisGetTimes:      1,
			redisDelErr:        nil,
			redisDelTimes:      1,
			internalXferErr:    errors.New("transaction failure"),
			internalXferTimes:  1,
		}, {
			name:               "valid",
			path:               "/exchange-xfer-fiat/valid",
			expectedMsg:        "successful",
			expectedStatus:     http.StatusOK,
			request:            models.HTTPFiatTransferRequest{OfferID: "VALID"},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			authDecryptErr:     nil,
			authDecryptTimes:   1,
			redisGetData:       validOffer,
			redisGetErr:        nil,
			redisGetTimes:      1,
			redisDelErr:        nil,
			redisDelTimes:      1,
			internalXferErr:    nil,
			internalXferTimes:  1,
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

			xferReqJSON, err := json.Marshal(&test.request)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

			gomock.InOrder(
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(validClientID, int64(0), test.authValidateJWTErr).
					Times(test.authValidateTimes),

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

			// Endpoint setup for test.
			router := gin.Default()
			router.POST(test.path, ExchangeTransferFiat(zapLogger, mockAuth, mockCache, mockDB, "Authorization"))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path, bytes.NewBuffer(xferReqJSON))
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, recorder.Code, "expected status codes do not match")

			var resp map[string]interface{}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp), "failed to unpack response.")

			errorMessage, ok := resp["message"].(string)
			require.True(t, ok, "failed to extract response message.")

			// Check for invalid currency codes and amount.
			if errorMessage == "invalid request" {
				payload, ok := resp["payload"].(string)
				require.True(t, ok, "failed to extract payload from response.")
				require.Contains(t, payload, test.expectedMsg)
			} else {
				require.Contains(t, errorMessage, test.expectedMsg, "incorrect response message.")
			}
		})
	}
}

func TestHandler_BalanceCurrencyFiat(t *testing.T) {
	t.Parallel()

	const basePath = "/fiat/balance/currency/"

	testCases := []struct {
		name               string
		currency           string
		expectedMsg        string
		expectedStatus     int
		authValidateJWTErr error
		authValidateTimes  int
		fiatBalanceErr     error
		fiatBalanceTimes   int
	}{
		{
			name:               "invalid currency",
			currency:           "INVALID",
			expectedMsg:        "invalid currency",
			expectedStatus:     http.StatusBadRequest,
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			fiatBalanceErr:     nil,
			fiatBalanceTimes:   0,
		}, {
			name:               "invalid JWT",
			currency:           "EUR",
			expectedMsg:        "invalid JWT",
			expectedStatus:     http.StatusForbidden,
			authValidateJWTErr: errors.New("invalid JWT"),
			authValidateTimes:  1,
			fiatBalanceErr:     nil,
			fiatBalanceTimes:   0,
		}, {
			name:               "unknown db error",
			currency:           "AED",
			expectedMsg:        "retry",
			expectedStatus:     http.StatusInternalServerError,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			fiatBalanceErr:     errors.New("unknown error"),
			fiatBalanceTimes:   1,
		}, {
			name:               "known db error",
			currency:           "CAD",
			expectedMsg:        "account not found",
			expectedStatus:     http.StatusNotFound,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			fiatBalanceErr:     postgres.ErrNotFound,
			fiatBalanceTimes:   1,
		}, {
			name:               "valid",
			currency:           "USD",
			expectedMsg:        "account balance",
			expectedStatus:     http.StatusOK,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			fiatBalanceErr:     nil,
			fiatBalanceTimes:   1,
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
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(uuid.UUID{}, int64(0), test.authValidateJWTErr).
					Times(test.authValidateTimes),

				mockDB.EXPECT().FiatBalanceCurrency(gomock.Any(), gomock.Any()).
					Return(postgres.FiatAccount{}, test.fiatBalanceErr).
					Times(test.fiatBalanceTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.GET(basePath+":currencyCode", BalanceCurrencyFiat(zapLogger, mockAuth, mockDB, "Authorization"))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodGet, basePath+test.currency, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, recorder.Code, "expected status codes do not match")

			var resp map[string]interface{}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp), "failed to unpack response.")

			actualMessage, ok := resp["message"].(string)
			require.True(t, ok, "failed to extract response message.")
			require.Contains(t, actualMessage, test.expectedMsg, "response message mismatch.")
		})
	}
}

func TestHandler_TxDetailsCurrencyFiat(t *testing.T) {
	t.Parallel()

	const basePath = "/fiat/transaction/details/"

	txID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate new UUID.")

	journalEntries := []postgres.FiatJournal{{}}

	testCases := []struct {
		name               string
		transactionID      string
		expectedMsg        string
		expectedStatus     int
		journalEntries     []postgres.FiatJournal
		authValidateJWTErr error
		authValidateTimes  int
		fiatTxDetailsErr   error
		fiatTxDetailsTimes int
	}{
		{
			name:               "invalid transaction ID",
			transactionID:      "INVALID",
			expectedMsg:        "invalid transaction ID",
			expectedStatus:     http.StatusBadRequest,
			journalEntries:     journalEntries,
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			fiatTxDetailsErr:   nil,
			fiatTxDetailsTimes: 0,
		}, {
			name:               "invalid JWT",
			transactionID:      txID.String(),
			expectedMsg:        "invalid JWT",
			expectedStatus:     http.StatusForbidden,
			journalEntries:     journalEntries,
			authValidateJWTErr: errors.New("invalid JWT"),
			authValidateTimes:  1,
			fiatTxDetailsErr:   nil,
			fiatTxDetailsTimes: 0,
		}, {
			name:               "unknown db error",
			transactionID:      txID.String(),
			expectedMsg:        "retry",
			expectedStatus:     http.StatusInternalServerError,
			journalEntries:     journalEntries,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			fiatTxDetailsErr:   errors.New("unknown error"),
			fiatTxDetailsTimes: 1,
		}, {
			name:               "known db error",
			transactionID:      txID.String(),
			expectedMsg:        "account not found",
			expectedStatus:     http.StatusNotFound,
			journalEntries:     journalEntries,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			fiatTxDetailsErr:   postgres.ErrNotFound,
			fiatTxDetailsTimes: 1,
		}, {
			name:               "transaction id not found",
			transactionID:      txID.String(),
			expectedMsg:        "transaction id not found",
			journalEntries:     nil,
			expectedStatus:     http.StatusNotFound,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			fiatTxDetailsErr:   nil,
			fiatTxDetailsTimes: 1,
		}, {
			name:               "valid",
			transactionID:      txID.String(),
			expectedMsg:        "transaction details",
			journalEntries:     journalEntries,
			expectedStatus:     http.StatusOK,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			fiatTxDetailsErr:   nil,
			fiatTxDetailsTimes: 1,
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
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(uuid.UUID{}, int64(0), test.authValidateJWTErr).
					Times(test.authValidateTimes),

				mockDB.EXPECT().FiatTxDetailsCurrency(gomock.Any(), gomock.Any()).
					Return(test.journalEntries, test.fiatTxDetailsErr).
					Times(test.fiatTxDetailsTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.GET(basePath+":transactionID", TxDetailsCurrencyFiat(zapLogger, mockAuth, mockDB, "Authorization"))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodGet, basePath+test.transactionID, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, recorder.Code, "expected status codes do not match")

			var resp map[string]interface{}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp), "failed to unpack response.")

			actualMessage, ok := resp["message"].(string)
			require.True(t, ok, "failed to extract response message.")
			require.Contains(t, actualMessage, test.expectedMsg, "response message mismatch.")
		})
	}
}

func TestHandler_BalanceCurrencyFiatPaginated(t *testing.T) {
	t.Parallel()

	const basePath = "/fiat/transaction/details/paginated/"

	accDetails := []postgres.FiatAccount{{}, {}, {}, {}}

	testCases := []struct {
		name                string
		path                string
		querySegment        string
		expectedMsg         string
		expectedStatus      int
		accDetails          []postgres.FiatAccount
		authValidateJWTErr  error
		authValidateTimes   int
		authDecryptStrErr   error
		authDecryptStrTimes int
		fiatBalanceErr      error
		fiatBalanceTimes    int
		authEncryptStrErr   error
		authEncryptStrTimes int
	}{
		{
			name:                "invalid JWT",
			path:                "invalid-jwt",
			querySegment:        "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedMsg:         "invalid JWT",
			accDetails:          accDetails,
			expectedStatus:      http.StatusForbidden,
			authValidateJWTErr:  errors.New("invalid JWT"),
			authValidateTimes:   1,
			authDecryptStrErr:   nil,
			authDecryptStrTimes: 0,
			fiatBalanceErr:      nil,
			fiatBalanceTimes:    0,
			authEncryptStrErr:   nil,
			authEncryptStrTimes: 0,
		}, {
			name:                "decrypt cursor failure",
			path:                "decrypt-cursor-failure",
			querySegment:        "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedMsg:         "invalid page cursor or page size",
			accDetails:          accDetails,
			expectedStatus:      http.StatusBadRequest,
			authValidateJWTErr:  nil,
			authValidateTimes:   1,
			authDecryptStrErr:   errors.New("decrypt failure"),
			authDecryptStrTimes: 1,
			fiatBalanceErr:      nil,
			fiatBalanceTimes:    0,
			authEncryptStrErr:   nil,
			authEncryptStrTimes: 0,
		}, {
			name:                "known db error",
			path:                "known-db-error",
			querySegment:        "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedMsg:         "not found",
			accDetails:          accDetails,
			expectedStatus:      http.StatusNotFound,
			authValidateJWTErr:  nil,
			authValidateTimes:   1,
			authDecryptStrErr:   nil,
			authDecryptStrTimes: 1,
			fiatBalanceErr:      postgres.ErrNotFound,
			fiatBalanceTimes:    1,
			authEncryptStrErr:   nil,
			authEncryptStrTimes: 0,
		}, {
			name:                "unknown db error",
			path:                "unknown-db-error",
			querySegment:        "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedMsg:         "retry",
			accDetails:          accDetails,
			expectedStatus:      http.StatusInternalServerError,
			authValidateJWTErr:  nil,
			authValidateTimes:   1,
			authDecryptStrErr:   nil,
			authDecryptStrTimes: 1,
			fiatBalanceErr:      errors.New("unknown db error"),
			fiatBalanceTimes:    1,
			authEncryptStrErr:   nil,
			authEncryptStrTimes: 0,
		}, {
			name:                "encrypt cursor failure",
			path:                "encrypt-cursor-failure",
			querySegment:        "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedMsg:         "retry",
			accDetails:          accDetails,
			expectedStatus:      http.StatusInternalServerError,
			authValidateJWTErr:  nil,
			authValidateTimes:   1,
			authDecryptStrErr:   nil,
			authDecryptStrTimes: 1,
			fiatBalanceErr:      nil,
			fiatBalanceTimes:    1,
			authEncryptStrErr:   errors.New("encrypt string error"),
			authEncryptStrTimes: 1,
		}, {
			name:                "valid without query and 10 records",
			path:                "valid-no-query-10-records",
			querySegment:        "",
			expectedMsg:         "account balances",
			accDetails:          []postgres.FiatAccount{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}},
			expectedStatus:      http.StatusOK,
			authValidateJWTErr:  nil,
			authValidateTimes:   1,
			authDecryptStrErr:   nil,
			authDecryptStrTimes: 0,
			fiatBalanceErr:      nil,
			fiatBalanceTimes:    1,
			authEncryptStrErr:   nil,
			authEncryptStrTimes: 0,
		}, {
			name:                "valid without query and 11 records",
			path:                "valid-no-query-11-records",
			querySegment:        "",
			expectedMsg:         "account balances",
			accDetails:          []postgres.FiatAccount{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}},
			expectedStatus:      http.StatusOK,
			authValidateJWTErr:  nil,
			authValidateTimes:   1,
			authDecryptStrErr:   nil,
			authDecryptStrTimes: 0,
			fiatBalanceErr:      nil,
			fiatBalanceTimes:    1,
			authEncryptStrErr:   nil,
			authEncryptStrTimes: 1,
		}, {
			name:                "valid without query",
			path:                "valid-no-query",
			querySegment:        "",
			expectedMsg:         "account balances",
			accDetails:          accDetails,
			expectedStatus:      http.StatusOK,
			authValidateJWTErr:  nil,
			authValidateTimes:   1,
			authDecryptStrErr:   nil,
			authDecryptStrTimes: 0,
			fiatBalanceErr:      nil,
			fiatBalanceTimes:    1,
			authEncryptStrErr:   nil,
			authEncryptStrTimes: 0,
		}, {
			name:                "valid with query",
			path:                "valid-with-query",
			querySegment:        "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedMsg:         "account balances",
			accDetails:          accDetails,
			expectedStatus:      http.StatusOK,
			authValidateJWTErr:  nil,
			authValidateTimes:   1,
			authDecryptStrErr:   nil,
			authDecryptStrTimes: 1,
			fiatBalanceErr:      nil,
			fiatBalanceTimes:    1,
			authEncryptStrErr:   nil,
			authEncryptStrTimes: 1,
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
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(uuid.UUID{}, int64(0), test.authValidateJWTErr).
					Times(test.authValidateTimes),

				mockAuth.EXPECT().DecryptFromString(gomock.Any()).
					Return([]byte{}, test.authDecryptStrErr).
					Times(test.authDecryptStrTimes),

				mockDB.EXPECT().FiatBalanceCurrencyPaginated(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(test.accDetails, test.fiatBalanceErr).
					Times(test.fiatBalanceTimes),

				mockAuth.EXPECT().EncryptToString(gomock.Any()).
					Return("encrypted-page-cursor", test.authEncryptStrErr).
					Times(test.authEncryptStrTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.GET(basePath+test.path, BalanceCurrencyFiatPaginated(zapLogger, mockAuth, mockDB, "Authorization"))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodGet, basePath+test.path+test.querySegment, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, recorder.Code, "expected status codes do not match")

			var resp map[string]interface{}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp), "failed to unpack response.")

			actualMessage, ok := resp["message"].(string)
			require.True(t, ok, "failed to extract response message.")
			require.Contains(t, actualMessage, test.expectedMsg, "response message mismatch.")
		})
	}
}
