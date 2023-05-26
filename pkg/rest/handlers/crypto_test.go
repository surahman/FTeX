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

func TestHandlers_OpenCrypto(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		path                 string
		expectedStatus       int
		request              *models.HTTPOpenCurrencyAccountRequest
		authValidateJWTErr   error
		authValidateTimes    int
		cryptoCreateAccErr   error
		cryptoCreateAccTimes int
	}{
		{
			name:                 "valid",
			path:                 "/open/valid",
			expectedStatus:       http.StatusCreated,
			request:              &models.HTTPOpenCurrencyAccountRequest{Currency: "BTC"},
			authValidateJWTErr:   nil,
			authValidateTimes:    1,
			cryptoCreateAccErr:   nil,
			cryptoCreateAccTimes: 1,
		}, {
			name:                 "validation",
			path:                 "/open/validation",
			expectedStatus:       http.StatusBadRequest,
			request:              &models.HTTPOpenCurrencyAccountRequest{},
			authValidateJWTErr:   nil,
			authValidateTimes:    0,
			cryptoCreateAccErr:   nil,
			cryptoCreateAccTimes: 0,
		}, {
			name:                 "invalid jwt",
			path:                 "/open/invalid-jwt",
			expectedStatus:       http.StatusForbidden,
			request:              &models.HTTPOpenCurrencyAccountRequest{Currency: "BTC"},
			authValidateJWTErr:   errors.New("invalid jwt"),
			authValidateTimes:    1,
			cryptoCreateAccErr:   nil,
			cryptoCreateAccTimes: 0,
		}, {
			name:                 "db failure",
			path:                 "/open/db-failure",
			expectedStatus:       http.StatusConflict,
			request:              &models.HTTPOpenCurrencyAccountRequest{Currency: "ETH"},
			authValidateJWTErr:   nil,
			authValidateTimes:    1,
			cryptoCreateAccErr:   postgres.ErrCreateFiat,
			cryptoCreateAccTimes: 1,
		}, {
			name:                 "db failure unknown",
			path:                 "/open/db-failure-unknown",
			expectedStatus:       http.StatusInternalServerError,
			request:              &models.HTTPOpenCurrencyAccountRequest{Currency: "USDC"},
			authValidateJWTErr:   nil,
			authValidateTimes:    1,
			cryptoCreateAccErr:   errors.New("unknown server error"),
			cryptoCreateAccTimes: 1,
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

				mockPostgres.EXPECT().CryptoCreateAccount(gomock.Any(), gomock.Any()).
					Return(test.cryptoCreateAccErr).
					Times(test.cryptoCreateAccTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.POST(test.path, OpenCrypto(zapLogger, mockAuth, mockPostgres, "Authorization"))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path, bytes.NewBuffer(openReqJSON))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, w.Code, "expected status codes do not match")
		})
	}
}

func TestHandlers_PurchaseOfferCrypto(t *testing.T) {
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
			path:               "/purchase-offer-crypto/empty-request",
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
			path:           "/purchase-offer-crypto/invalid-fiat-currency",
			expectedStatus: http.StatusBadRequest,
			request: &models.HTTPExchangeOfferRequest{
				SourceCurrency:      "INVALID",
				DestinationCurrency: "BTC",
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
			path:           "/purchase-offer-crypto/too-many-decimal-places",
			expectedStatus: http.StatusBadRequest,
			request: &models.HTTPExchangeOfferRequest{
				SourceCurrency:      "USD",
				DestinationCurrency: "BTC",
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
			path:           "/purchase-offer-crypto/negative",
			expectedStatus: http.StatusBadRequest,
			request: &models.HTTPExchangeOfferRequest{
				SourceCurrency:      "USD",
				DestinationCurrency: "BTC",
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
			path:           "/purchase-offer-crypto/invalid-jwt",
			expectedStatus: http.StatusForbidden,
			request: &models.HTTPExchangeOfferRequest{
				SourceCurrency:      "USD",
				DestinationCurrency: "BTC",
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
			name:           "crypto conversion rate error",
			expectedMsg:    "retry",
			path:           "/purchase-offer-crypto/crypto-rate-error",
			expectedStatus: http.StatusInternalServerError,
			request: &models.HTTPExchangeOfferRequest{
				SourceCurrency:      "USD",
				DestinationCurrency: "BTC",
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
			path:           "/purchase-offer-crypto/encryption-error",
			expectedStatus: http.StatusInternalServerError,
			request: &models.HTTPExchangeOfferRequest{
				SourceCurrency:      "USD",
				DestinationCurrency: "BTC",
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
			path:           "/purchase-offer-crypto/redis-error",
			expectedStatus: http.StatusInternalServerError,
			request: &models.HTTPExchangeOfferRequest{
				SourceCurrency:      "USD",
				DestinationCurrency: "BTC",
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
			path:           "/purchase-offer-crypto/valid",
			expectedStatus: http.StatusOK,
			request: &models.HTTPExchangeOfferRequest{
				SourceCurrency:      "USD",
				DestinationCurrency: "BTC",
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

				mockQuotes.EXPECT().CryptoConversion(gomock.Any(), gomock.Any(), gomock.Any(), true, nil).
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
			router.POST(test.path, PurchaseOfferCrypto(zapLogger, mockAuth, mockCache, mockQuotes, "Authorization"))
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
			if errorMessage == "invalid request" { //nolint:goconst
				payload, ok := resp["payload"].(string)
				require.True(t, ok, "failed to extract payload from response.")
				require.Contains(t, payload, test.expectedMsg)
			} else {
				require.Contains(t, errorMessage, test.expectedMsg, "incorrect response message.")
			}
		})
	}
}
