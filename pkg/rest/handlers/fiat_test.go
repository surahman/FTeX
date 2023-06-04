package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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

	//nolint:dupl
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
		request            *models.HTTPExchangeOfferRequest
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
			request:            &models.HTTPExchangeOfferRequest{},
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
			expectedMsg:    "Fiat currency",
			path:           "/exchange-offer-fiat/invalid-src-currency",
			expectedStatus: http.StatusBadRequest,
			request: &models.HTTPExchangeOfferRequest{
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
			expectedMsg:    "Fiat currency",
			path:           "/exchange-offer-fiat/invalid-dst-currency",
			expectedStatus: http.StatusBadRequest,
			request: &models.HTTPExchangeOfferRequest{
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
			request: &models.HTTPExchangeOfferRequest{
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
			request: &models.HTTPExchangeOfferRequest{
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
			request: &models.HTTPExchangeOfferRequest{
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
			request: &models.HTTPExchangeOfferRequest{
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
			request: &models.HTTPExchangeOfferRequest{
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
			request: &models.HTTPExchangeOfferRequest{
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
			request: &models.HTTPExchangeOfferRequest{
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
			if errorMessage == constants.GetInvalidRequest() {
				payload, ok := resp["payload"].(string)
				require.True(t, ok, "failed to extract payload from response.")
				require.Contains(t, payload, test.expectedMsg)
			} else {
				require.Contains(t, errorMessage, test.expectedMsg, "incorrect response message.")
			}
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
		name               string
		path               string
		expectedMsg        string
		expectedStatus     int
		request            models.HTTPTransferRequest
		authValidateJWTErr error
		authValidateTimes  int
		authDecryptErr     error
		authDecryptTimes   int
		redisGetData       models.HTTPExchangeOfferResponse
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
			request:            models.HTTPTransferRequest{},
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
			request:            models.HTTPTransferRequest{OfferID: "VALID"},
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
			request:            models.HTTPTransferRequest{OfferID: "VALID"},
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
			request:            models.HTTPTransferRequest{OfferID: "VALID"},
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
			request:            models.HTTPTransferRequest{OfferID: "VALID"},
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
			request:            models.HTTPTransferRequest{OfferID: "VALID"},
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
			request:            models.HTTPTransferRequest{OfferID: "VALID"},
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
			request:            models.HTTPTransferRequest{OfferID: "VALID"},
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
			name:               "crypto purchase",
			path:               "/exchange-xfer-fiat/crypto-purchase",
			expectedMsg:        "invalid Fiat currency",
			expectedStatus:     http.StatusBadRequest,
			request:            models.HTTPTransferRequest{OfferID: "VALID"},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			authDecryptErr:     nil,
			authDecryptTimes:   1,
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
		}, {
			name:               "crypto sale",
			path:               "/exchange-xfer-fiat/crypto-sale",
			expectedMsg:        "invalid Fiat currency",
			expectedStatus:     http.StatusBadRequest,
			request:            models.HTTPTransferRequest{OfferID: "VALID"},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			authDecryptErr:     nil,
			authDecryptTimes:   1,
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
		}, {
			name:               "invalid source destination amount",
			path:               "/exchange-xfer-fiat/invalid-source-destination-amount",
			expectedMsg:        "invalid",
			expectedStatus:     http.StatusBadRequest,
			request:            models.HTTPTransferRequest{OfferID: "VALID"},
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
			request:            models.HTTPTransferRequest{OfferID: "VALID"},
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
			request:            models.HTTPTransferRequest{OfferID: "VALID"},
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
			if errorMessage == constants.GetInvalidRequest() {
				payload, ok := resp["payload"].(string)
				require.True(t, ok, "failed to extract payload from response.")
				require.Contains(t, payload, test.expectedMsg)
			} else {
				require.Contains(t, errorMessage, test.expectedMsg, "incorrect response message.")
			}
		})
	}
}

func TestHandler_BalanceCurrencyFiat(t *testing.T) { //nolint:dupl
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
			expectedMsg:        "records not found",
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
			router.GET(basePath+":ticker", BalanceFiat(zapLogger, mockAuth, mockDB, "Authorization"))
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

	cryptoJournal := []postgres.CryptoJournal{{}}
	fiatJournal := []postgres.FiatJournal{{}}

	testCases := []struct {
		name               string
		transactionID      string
		expectedMsg        string
		expectedStatus     int
		fiatJournal        []postgres.FiatJournal
		cryptoJournal      []postgres.CryptoJournal
		authValidateJWTErr error
		authValidateTimes  int
		fiatTxErr          error
		fiatTxTimes        int
		cryptoTxErr        error
		cryptoTxTimes      int
	}{
		{
			name:               "invalid transaction ID",
			transactionID:      "INVALID",
			expectedMsg:        "invalid transaction ID",
			expectedStatus:     http.StatusBadRequest,
			fiatJournal:        fiatJournal,
			cryptoJournal:      cryptoJournal,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			fiatTxErr:          nil,
			fiatTxTimes:        0,
			cryptoTxErr:        nil,
			cryptoTxTimes:      0,
		}, {
			name:               "invalid JWT",
			transactionID:      txID.String(),
			expectedMsg:        "invalid JWT",
			expectedStatus:     http.StatusForbidden,
			fiatJournal:        fiatJournal,
			cryptoJournal:      cryptoJournal,
			authValidateJWTErr: errors.New("invalid JWT"),
			authValidateTimes:  1,
			fiatTxErr:          nil,
			fiatTxTimes:        0,
			cryptoTxErr:        nil,
			cryptoTxTimes:      0,
		}, {
			name:               "unknown db error",
			transactionID:      txID.String(),
			expectedMsg:        "retry",
			expectedStatus:     http.StatusInternalServerError,
			fiatJournal:        fiatJournal,
			cryptoJournal:      cryptoJournal,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			fiatTxErr:          errors.New("unknown error"),
			fiatTxTimes:        1,
			cryptoTxErr:        nil,
			cryptoTxTimes:      0,
		}, {
			name:               "known db error",
			transactionID:      txID.String(),
			expectedMsg:        "records not found",
			expectedStatus:     http.StatusNotFound,
			fiatJournal:        fiatJournal,
			cryptoJournal:      cryptoJournal,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			fiatTxErr:          postgres.ErrNotFound,
			fiatTxTimes:        1,
			cryptoTxErr:        nil,
			cryptoTxTimes:      0,
		}, {
			name:               "transaction id not found",
			transactionID:      txID.String(),
			expectedMsg:        "transaction id not found",
			fiatJournal:        nil,
			cryptoJournal:      nil,
			expectedStatus:     http.StatusNotFound,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			fiatTxErr:          nil,
			fiatTxTimes:        1,
			cryptoTxErr:        nil,
			cryptoTxTimes:      1,
		}, {
			name:               "valid",
			transactionID:      txID.String(),
			expectedMsg:        "transaction details",
			fiatJournal:        fiatJournal,
			cryptoJournal:      cryptoJournal,
			expectedStatus:     http.StatusOK,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			fiatTxErr:          nil,
			fiatTxTimes:        1,
			cryptoTxErr:        nil,
			cryptoTxTimes:      1,
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
					Return(test.fiatJournal, test.fiatTxErr).
					Times(test.fiatTxTimes),

				mockDB.EXPECT().CryptoTxDetailsCurrency(gomock.Any(), gomock.Any()).
					Return(test.cryptoJournal, test.cryptoTxErr).
					Times(test.cryptoTxTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.GET(basePath+":transactionID", TxDetailsFiat(zapLogger, mockAuth, mockDB, "Authorization"))
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

func TestHandler_BalanceCurrencyFiatPaginated(t *testing.T) { //nolint:dupl
	t.Parallel()

	const basePath = "/fiat/balances/details/paginated/"

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
			router.GET(basePath+test.path, BalanceFiatPaginated(zapLogger, mockAuth, mockDB, "Authorization"))
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

func TestHandler_TxDetailsCurrencyFiatPaginated(t *testing.T) {
	t.Parallel()

	const basePath = "/fiat/transaction/details-journal/paginated/"

	decryptedCursor := fmt.Sprintf("%s,%s,%d",
		fmt.Sprintf(constants.GetMonthFormatString(), 2023, 6, "-04:00"),
		fmt.Sprintf(constants.GetMonthFormatString(), 2023, 7, "-04:00"),
		10)

	journalEntries := []postgres.FiatJournal{{}, {}, {}, {}}

	testCases := []struct {
		name                   string
		path                   string
		currency               string
		querySegment           string
		expectedMsg            string
		expectedStatus         int
		journalEntries         []postgres.FiatJournal
		authValidateJWTErr     error
		authValidateTimes      int
		authDecryptCursorErr   error
		authDecryptCursorTimes int
		authEncryptCursorErr   error
		authEncryptCursorTimes int
		fiatTxPaginatedErr     error
		fiatTxPaginatedTimes   int
	}{
		{
			name:                   "auth failure",
			path:                   "auth-failure/",
			currency:               "USD",
			querySegment:           "?pageCursor=page-cursor",
			expectedMsg:            "auth failure",
			journalEntries:         journalEntries,
			expectedStatus:         http.StatusForbidden,
			authValidateJWTErr:     errors.New("auth failure"),
			authValidateTimes:      1,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 0,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 0,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   0,
		}, {
			name:                   "bad currency",
			path:                   "bad-currency/",
			currency:               "INVALID",
			querySegment:           "?pageCursor=page-cursor",
			expectedMsg:            "invalid currency",
			journalEntries:         journalEntries,
			expectedStatus:         http.StatusBadRequest,
			authValidateJWTErr:     nil,
			authValidateTimes:      1,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 0,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 0,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   0,
		}, {
			name:                   "no cursor or params",
			path:                   "no-cursor-or-params/",
			currency:               "USD",
			querySegment:           "?",
			expectedMsg:            "missing required parameters",
			journalEntries:         journalEntries,
			expectedStatus:         http.StatusBadRequest,
			authValidateJWTErr:     nil,
			authValidateTimes:      1,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 0,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 0,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   0,
		}, {
			name:                   "db failure",
			path:                   "db-failure/",
			currency:               "USD",
			querySegment:           "?pageCursor=page-cursor",
			expectedMsg:            "records not found",
			journalEntries:         journalEntries,
			expectedStatus:         http.StatusNotFound,
			authValidateJWTErr:     nil,
			authValidateTimes:      1,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 1,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     postgres.ErrNotFound,
			fiatTxPaginatedTimes:   1,
		}, {
			name:                   "unknown db failure",
			path:                   "unknown-db-failure/",
			currency:               "USD",
			querySegment:           "?pageCursor=page-cursor",
			expectedMsg:            "retry",
			journalEntries:         journalEntries,
			expectedStatus:         http.StatusInternalServerError,
			authValidateJWTErr:     nil,
			authValidateTimes:      1,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 1,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     errors.New("db failure"),
			fiatTxPaginatedTimes:   1,
		}, {
			name:                   "no transactions",
			path:                   "no-transactions/",
			currency:               "USD",
			querySegment:           "?pageCursor=page-cursor",
			expectedMsg:            "no transactions",
			journalEntries:         []postgres.FiatJournal{},
			expectedStatus:         http.StatusRequestedRangeNotSatisfiable,
			authValidateJWTErr:     nil,
			authValidateTimes:      1,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 1,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   1,
		}, {
			name:                   "valid with cursor",
			path:                   "valid-with-cursor/",
			currency:               "USD",
			querySegment:           "?pageCursor=page-cursor",
			expectedMsg:            "account transactions",
			journalEntries:         journalEntries,
			expectedStatus:         http.StatusOK,
			authValidateJWTErr:     nil,
			authValidateTimes:      1,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 1,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   1,
		}, {
			name:                   "valid with query",
			path:                   "valid-with-query/",
			currency:               "USD",
			querySegment:           "?month=6&year=2023&timezone=%2B04:00&pageSize=3",
			expectedMsg:            "account transactions",
			journalEntries:         journalEntries,
			expectedStatus:         http.StatusOK,
			authValidateJWTErr:     nil,
			authValidateTimes:      1,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 0,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   1,
		},
	}

	for _, testCase := range testCases { //nolint:dupl
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
					Return([]byte(decryptedCursor), test.authDecryptCursorErr).
					Times(test.authDecryptCursorTimes),

				mockAuth.EXPECT().EncryptToString(gomock.Any()).
					Return("encrypted-cursor", test.authEncryptCursorErr).
					Times(test.authEncryptCursorTimes),

				mockDB.EXPECT().FiatTransactionsCurrencyPaginated(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any()).
					Return(test.journalEntries, test.fiatTxPaginatedErr).
					Times(test.fiatTxPaginatedTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.GET(basePath+test.path+":currencyCode",
				TxDetailsFiatPaginated(zapLogger, mockAuth, mockDB, "Authorization"))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodGet,
				basePath+test.path+test.currency+test.querySegment, nil)
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
