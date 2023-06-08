package graphql

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
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/quotes"
)

func TestCryptoResolver_OpenCrypto(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		path                 string
		query                string
		expectErr            bool
		authValidateJWTErr   error
		authValidateJWTTimes int
		cryptoCreateAccErr   error
		cryptoCreateAccTimes int
	}{
		{
			name:                 "empty request",
			path:                 "/open-crypto/empty-request",
			query:                fmt.Sprintf(testCryptoQuery["openCrypto"], ""),
			expectErr:            true,
			authValidateJWTErr:   errors.New("invalid token"),
			authValidateJWTTimes: 1,
			cryptoCreateAccErr:   nil,
			cryptoCreateAccTimes: 0,
		}, {
			name:                 "db failure",
			path:                 "/open-crypto/db-failure",
			query:                fmt.Sprintf(testCryptoQuery["openCrypto"], "BTC"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			cryptoCreateAccErr:   postgres.ErrNotFound,
			cryptoCreateAccTimes: 1,
		}, {
			name:                 "valid",
			path:                 "/open-crypto/valid",
			query:                fmt.Sprintf(testCryptoQuery["openCrypto"], "BTC"),
			expectErr:            false,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			cryptoCreateAccErr:   nil,
			cryptoCreateAccTimes: 1,
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
			mockPostgres := mocks.NewMockPostgres(mockCtrl)
			mockRedis := mocks.NewMockRedis(mockCtrl)    // Not called.
			mockQuotes := quotes.NewMockQuotes(mockCtrl) // Not called.

			gomock.InOrder(
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(uuid.UUID{}, int64(-1), test.authValidateJWTErr).
					Times(test.authValidateJWTTimes),

				mockPostgres.EXPECT().CryptoCreateAccount(gomock.Any(), gomock.Any()).
					Return(test.cryptoCreateAccErr).
					Times(test.cryptoCreateAccTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.Use(GinContextToContextMiddleware())
			router.POST(test.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockPostgres, mockQuotes, zapLogger))

			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path,
				bytes.NewBufferString(test.query))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "some valid auth token goes here")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// Verify responses
			require.Equal(t, http.StatusOK, recorder.Code, "expected status codes do not match")

			response := map[string]any{}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response), "failed to unmarshal response body")

			// Error is expected check to ensure one is set.
			if test.expectErr {
				verifyErrorReturned(t, response)
			}
		})
	}
}

func TestCryptoResolver_CryptoOfferRequestResolver(t *testing.T) {
	t.Parallel()

	var (
		resolver     cryptoOfferRequestResolver
		input        models.HTTPCryptoOfferRequest
		sourceFloat  = 7890.1011
		sourceAmount = decimal.NewFromFloat(sourceFloat)
	)

	err := resolver.SourceAmount(context.TODO(), &input, sourceFloat)
	require.NoError(t, err, "source amount should always return a nil error.")

	require.Equal(t, sourceAmount, input.SourceAmount, "source amounts mismatched.")
}

func TestCryptoResolver_OfferCrypto(t *testing.T) {
	t.Parallel()

	var (
		validFloat    = float64(999)
		negativeFloat = float64(-999)
		tooManyFloat  = 999.999
		amountValid   = decimal.NewFromFloat(validFloat)
	)

	testCases := []struct {
		name               string
		path               string
		query              string
		expectErr          bool
		isPurchase         bool
		authValidateJWTErr error
		authValidateTimes  int
		quotesErr          error
		quotesAmount       decimal.Decimal
		quotesTimes        int
		authEncryptErr     error
		authEncryptTimes   int
		redisErr           error
		redisTimes         int
	}{
		{
			name:               "invalid source currency",
			path:               "/offer-crypto/invalid-fiat-currency",
			query:              fmt.Sprintf(testCryptoQuery["offerCrypto"], tooManyFloat, "INVALID", "BTC", true),
			isPurchase:         true,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesAmount:       amountValid,
			quotesTimes:        0,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:               "too many decimal places",
			path:               "/offer-crypto/too-many-decimal-places",
			query:              fmt.Sprintf(testCryptoQuery["offerCrypto"], tooManyFloat, "USD", "BTC", true),
			isPurchase:         true,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesAmount:       amountValid,
			quotesTimes:        0,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:               "negative",
			path:               "/offer-crypto/negative",
			query:              fmt.Sprintf(testCryptoQuery["offerCrypto"], negativeFloat, "USD", "BTC", true),
			isPurchase:         true,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesAmount:       amountValid,
			quotesTimes:        0,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:               "invalid jwt",
			path:               "/offer-crypto/invalid-jwt",
			query:              fmt.Sprintf(testCryptoQuery["offerCrypto"], validFloat, "USD", "BTC", true),
			isPurchase:         true,
			authValidateJWTErr: errors.New("invalid jwt"),
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesAmount:       amountValid,
			quotesTimes:        0,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:               "crypto conversion rate error",
			path:               "/offer-crypto/crypto-rate-error",
			query:              fmt.Sprintf(testCryptoQuery["offerCrypto"], validFloat, "USD", "BTC", true),
			isPurchase:         true,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          errors.New(""),
			quotesAmount:       amountValid,
			quotesTimes:        1,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:               "crypto conversion amount too small",
			path:               "/offer-crypto/crypto-amount-too-small",
			query:              fmt.Sprintf(testCryptoQuery["offerCrypto"], validFloat, "USD", "BTC", true),
			isPurchase:         true,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesAmount:       decimal.NewFromFloat(0),
			quotesTimes:        1,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:               "encryption error",
			path:               "/offer-crypto/encryption-error",
			query:              fmt.Sprintf(testCryptoQuery["offerCrypto"], validFloat, "USD", "BTC", true),
			isPurchase:         true,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesAmount:       amountValid,
			quotesTimes:        1,
			authEncryptErr:     errors.New("encryption error"),
			authEncryptTimes:   1,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:               "redis error",
			path:               "/offer-crypto/redis-error",
			query:              fmt.Sprintf(testCryptoQuery["offerCrypto"], validFloat, "USD", "BTC", true),
			isPurchase:         true,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesAmount:       amountValid,
			quotesTimes:        1,
			authEncryptErr:     nil,
			authEncryptTimes:   1,
			redisErr:           errors.New("redis error"),
			redisTimes:         1,
		}, {
			name:               "valid - purchase",
			path:               "/offer-crypto/valid-purchase",
			query:              fmt.Sprintf(testCryptoQuery["offerCrypto"], validFloat, "USD", "BTC", true),
			isPurchase:         true,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesAmount:       amountValid,
			quotesTimes:        1,
			authEncryptErr:     nil,
			authEncryptTimes:   1,
			redisErr:           nil,
			redisTimes:         1,
		}, {
			name:               "valid - sale",
			path:               "/offer-crypto/valid-sale",
			query:              fmt.Sprintf(testCryptoQuery["offerCrypto"], validFloat, "BTC", "USD", false),
			isPurchase:         false,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesAmount:       amountValid,
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
			mockPostgres := mocks.NewMockPostgres(mockCtrl) // Not called.
			mockRedis := mocks.NewMockRedis(mockCtrl)
			mockQuotes := quotes.NewMockQuotes(mockCtrl)

			gomock.InOrder(
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(uuid.UUID{}, int64(0), test.authValidateJWTErr).
					Times(test.authValidateTimes),

				mockQuotes.EXPECT().CryptoConversion(gomock.Any(), gomock.Any(), gomock.Any(), test.isPurchase, nil).
					Return(amountValid, test.quotesAmount, test.quotesErr).
					Times(test.quotesTimes),

				mockAuth.EXPECT().EncryptToString(gomock.Any()).
					Return("OFFER-ID", test.authEncryptErr).
					Times(test.authEncryptTimes),

				mockRedis.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(test.redisErr).
					Times(test.redisTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.Use(GinContextToContextMiddleware())
			router.POST(test.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockPostgres, mockQuotes, zapLogger))

			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path,
				bytes.NewBufferString(test.query))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "some valid auth token goes here")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// Verify responses
			require.Equal(t, http.StatusOK, recorder.Code, "expected status codes do not match")

			response := map[string]any{}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response), "failed to unmarshal response body")

			// Error is expected check to ensure one is set.
			if test.expectErr {
				verifyErrorReturned(t, response)
			}
		})
	}
}
