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

			conversionReqJSON, err := json.Marshal(&test.request)
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
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path, bytes.NewBuffer(conversionReqJSON))
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
