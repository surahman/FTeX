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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
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

	t.Run("SourceAmount", func(t *testing.T) {
		t.Parallel()

		err := resolver.SourceAmount(context.TODO(), &input, sourceFloat)
		require.NoError(t, err, "source amount should always return a nil error.")
		require.Equal(t, sourceAmount, input.SourceAmount, "source amounts mismatched.")
	})
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

func TestCryptoResolver_CryptoJournalResolver(t *testing.T) {
	t.Parallel()

	resolver := cryptoJournalResolver{}

	clientID, err := uuid.NewV4()
	require.NoError(t, err, "client id generation failed")

	txID, err := uuid.NewV4()
	require.NoError(t, err, "tx id generation failed")

	obj := &postgres.CryptoJournal{
		Ticker:       "BTC",
		Amount:       decimal.NewFromFloat(78910.11),
		TransactedAt: pgtype.Timestamptz{},
		ClientID:     clientID,
		TxID:         txID,
	}

	t.Run("Amount", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.Amount(context.TODO(), obj)
		require.NoError(t, err, "failed to resolve amount.")
		require.InDelta(t, obj.Amount.InexactFloat64(), result, 0.01, "amount mismatched.")
	})

	t.Run("TransactedAt", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.TransactedAt(context.TODO(), obj)
		require.NoError(t, err, "failed to resolve transacted at.")
		require.Equal(t, obj.TransactedAt.Time.String(), result, "transacted at mismatched.")
	})

	t.Run("ClientID", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.ClientID(context.TODO(), obj)
		require.NoError(t, err, "failed to resolve client id")
		require.Equal(t, obj.ClientID.String(), result, "client id mismatched.")
	})

	t.Run("TxID", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.TxID(context.TODO(), obj)
		require.NoError(t, err, "failed to resolve tx id")
		require.Equal(t, obj.TxID.String(), result, "client tx mismatched.")
	})
}

func TestCryptoResolver_ExchangeCrypto(t *testing.T) {
	t.Parallel()

	validClientID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate a valid uuid.")

	cryptoAmount := decimal.NewFromFloat(1234.56)
	fiatAmount := decimal.NewFromFloat(78910.11)

	validSale := models.HTTPExchangeOfferResponse{
		PriceQuote: models.PriceQuote{
			ClientID:       validClientID,
			SourceAcc:      "BTC",
			DestinationAcc: "USD",
			Rate:           decimal.Decimal{},
			Amount:         fiatAmount,
		},
		DebitAmount:      cryptoAmount,
		OfferID:          "OFFER-ID",
		Expires:          0,
		IsCryptoPurchase: false,
		IsCryptoSale:     true,
	}

	validPurchase := models.HTTPExchangeOfferResponse{
		PriceQuote: models.PriceQuote{
			ClientID:       validClientID,
			SourceAcc:      "USD",
			DestinationAcc: "BTC",
			Rate:           decimal.Decimal{},
			Amount:         cryptoAmount,
		},
		DebitAmount:      fiatAmount,
		OfferID:          "OFFER-ID",
		Expires:          0,
		IsCryptoPurchase: true,
		IsCryptoSale:     false,
	}

	testCases := []struct {
		name               string
		path               string
		query              string
		expectErr          bool
		authValidateJWTErr error
		authValidateTimes  int
		authEncryptTimes   int
		authEncryptErr     error
		redisGetData       models.HTTPExchangeOfferResponse
		redisGetTimes      int
		redisDelTimes      int
		purchaseTimes      int
		sellTimes          int
	}{
		{
			name:               "invalid jwt",
			path:               "/exchange-crypto/invalid-jwt",
			query:              fmt.Sprintf(testCryptoQuery["exchangeCrypto"], "OFFER-ID"),
			expectErr:          true,
			authValidateTimes:  1,
			authValidateJWTErr: errors.New("invalid jwt"),
			authEncryptTimes:   0,
			authEncryptErr:     nil,
			redisGetData:       validPurchase,
			redisGetTimes:      0,
			redisDelTimes:      0,
			purchaseTimes:      0,
			sellTimes:          0,
		}, {
			name:               "transaction failure",
			path:               "/exchange-crypto/transaction-failure",
			query:              fmt.Sprintf(testCryptoQuery["exchangeCrypto"], "OFFER-ID"),
			expectErr:          true,
			authValidateTimes:  1,
			authValidateJWTErr: nil,
			authEncryptTimes:   1,
			authEncryptErr:     errors.New("transaction failure"),
			redisGetData:       validPurchase,
			redisGetTimes:      0,
			redisDelTimes:      0,
			purchaseTimes:      0,
			sellTimes:          0,
		}, {
			name:               "valid - purchase",
			path:               "/exchange-crypto/valid-purchase",
			query:              fmt.Sprintf(testCryptoQuery["exchangeCrypto"], "OFFER-ID"),
			expectErr:          false,
			authValidateTimes:  1,
			authValidateJWTErr: nil,
			authEncryptTimes:   1,
			authEncryptErr:     nil,
			redisGetData:       validPurchase,
			redisGetTimes:      1,
			redisDelTimes:      1,
			purchaseTimes:      1,
			sellTimes:          0,
		}, {
			name:               "valid - sale",
			path:               "/exchange-crypto/valid-sale",
			query:              fmt.Sprintf(testCryptoQuery["exchangeCrypto"], "OFFER-ID"),
			expectErr:          false,
			authValidateTimes:  1,
			authValidateJWTErr: nil,
			authEncryptTimes:   1,
			authEncryptErr:     nil,
			redisGetData:       validSale,
			redisGetTimes:      1,
			redisDelTimes:      1,
			purchaseTimes:      0,
			sellTimes:          1,
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
			mockRedis := mocks.NewMockRedis(mockCtrl)
			mockQuotes := quotes.NewMockQuotes(mockCtrl) // Not called.

			gomock.InOrder(
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(validClientID, int64(0), test.authValidateJWTErr).
					Times(test.authValidateTimes),

				mockAuth.EXPECT().DecryptFromString(gomock.Any()).
					Return([]byte("OFFER-ID"), test.authEncryptErr).
					Times(test.authEncryptTimes),

				mockRedis.EXPECT().Get(gomock.Any(), gomock.Any()).
					Return(nil).
					SetArg(1, test.redisGetData).
					Times(test.redisGetTimes),

				mockRedis.EXPECT().Del(gomock.Any()).
					Return(nil).
					Times(test.redisDelTimes),

				mockPostgres.EXPECT().CryptoPurchase(
					gomock.Any(), postgres.CurrencyUSD, fiatAmount, "BTC", cryptoAmount).
					Return(&postgres.FiatJournal{}, &postgres.CryptoJournal{}, nil).
					Times(test.purchaseTimes),

				mockPostgres.EXPECT().CryptoSell(
					gomock.Any(), postgres.CurrencyUSD, fiatAmount, "BTC", cryptoAmount).
					Return(&postgres.FiatJournal{}, &postgres.CryptoJournal{}, nil).
					Times(test.sellTimes),
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

func TestCryptoResolver_CryptoAccountResolver(t *testing.T) {
	t.Parallel()

	resolver := cryptoAccountResolver{}

	clientID, err := uuid.NewV4()
	require.NoError(t, err, "client id generation failed")

	lastTxTS := time.Now().Add(-15 * time.Second)
	lastTxTSPG := pgtype.Timestamptz{}
	require.NoError(t, lastTxTSPG.Scan(lastTxTS), "failed to generate lastTxTs.")

	createdAt := time.Now().Add(-15 * time.Minute)
	createdAtPG := pgtype.Timestamptz{}
	require.NoError(t, createdAtPG.Scan(createdAt), "failed to generate createdAt.")

	obj := &postgres.CryptoAccount{
		Ticker:    "BTC",
		Balance:   decimal.NewFromFloat(46.39),
		LastTx:    decimal.NewFromFloat(-789.33),
		LastTxTs:  lastTxTSPG,
		CreatedAt: createdAtPG,
		ClientID:  clientID,
	}

	t.Run("Balance", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.Balance(context.TODO(), obj)
		require.NoError(t, err, "failed to resolve balance.")
		require.InDelta(t, obj.Balance.InexactFloat64(), result, 0.01, "balance mismatched.")
	})

	t.Run("LastTx", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.LastTx(context.TODO(), obj)
		require.NoError(t, err, "failed to resolve last tx.")
		require.InDelta(t, obj.LastTx.InexactFloat64(), result, 0.01, "last tx mismatched.")
	})

	t.Run("LastTxTs", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.LastTxTs(context.TODO(), obj)
		require.NoError(t, err, "failed to resolve LastTxTs.")
		require.Equal(t, obj.LastTxTs.Time.String(), result, "LastTxTs mismatched.")
	})

	t.Run("CreatedAt", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.CreatedAt(context.TODO(), obj)
		require.NoError(t, err, "failed to resolve CreatedAt.")
		require.Equal(t, obj.CreatedAt.Time.String(), result, "CreatedAt mismatched.")
	})

	t.Run("ClientID", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.ClientID(context.TODO(), obj)
		require.NoError(t, err, "failed to resolve client id")
		require.Equal(t, obj.ClientID.String(), result, "client id mismatched.")
	})
}

func TestCryptoResolver_BalanceCrypto(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		path               string
		query              string
		expectErr          bool
		authValidateJWTErr error
		authValidateTimes  int
		balanceErr         error
		balanceTimes       int
	}{
		{
			name:               "invalid jwt",
			path:               "/balance-crypto/invalid-jwt",
			query:              fmt.Sprintf(testCryptoQuery["balanceCrypto"], "ETH"),
			expectErr:          true,
			authValidateTimes:  1,
			authValidateJWTErr: errors.New("invalid jwt"),
			balanceTimes:       0,
			balanceErr:         nil,
		}, {
			name:               "invalid",
			path:               "/balance-crypto/invalid",
			query:              fmt.Sprintf(testCryptoQuery["balanceCrypto"], "INVALID"),
			expectErr:          true,
			authValidateTimes:  1,
			authValidateJWTErr: nil,
			balanceTimes:       0,
			balanceErr:         nil,
		}, {
			name:               "db failure",
			path:               "/balance-crypto/db-failure",
			query:              fmt.Sprintf(testCryptoQuery["balanceCrypto"], "ETH"),
			expectErr:          false,
			authValidateTimes:  1,
			authValidateJWTErr: nil,
			balanceTimes:       1,
			balanceErr:         postgres.ErrNotFound,
		}, {
			name:               "valid",
			path:               "/balance-crypto/valid",
			query:              fmt.Sprintf(testCryptoQuery["balanceCrypto"], "ETH"),
			expectErr:          false,
			authValidateTimes:  1,
			authValidateJWTErr: nil,
			balanceTimes:       1,
			balanceErr:         nil,
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
					Return(uuid.UUID{}, int64(0), test.authValidateJWTErr).
					Times(test.authValidateTimes),

				mockPostgres.EXPECT().CryptoBalance(gomock.Any(), gomock.Any()).
					Return(postgres.CryptoAccount{}, test.balanceErr).
					Times(test.balanceTimes),
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

func TestCryptoResolver_BalanceAllCrypto(t *testing.T) {
	t.Parallel()

	accDetails := []postgres.CryptoAccount{{}, {}, {}, {}}

	testCases := []struct {
		name                 string
		path                 string
		query                string
		expectErr            bool
		accDetails           []postgres.CryptoAccount
		authValidateJWTErr   error
		authValidateJWTTimes int
		authDecryptStrErr    error
		authDecryptStrTimes  int
		fiatBalanceErr       error
		fiatBalanceTimes     int
		authEncryptStrErr    error
		authEncryptStrTimes  int
	}{
		{
			name:                 "invalid JWT",
			path:                 "/balance-all-crypto/invalid-jwt",
			query:                fmt.Sprintf(testCryptoQuery["balanceAllCrypto"], "page-cursor", 3),
			expectErr:            true,
			accDetails:           accDetails,
			authValidateJWTErr:   errors.New("invalid JWT"),
			authValidateJWTTimes: 1,
			authDecryptStrErr:    nil,
			authDecryptStrTimes:  0,
			fiatBalanceErr:       nil,
			fiatBalanceTimes:     0,
			authEncryptStrErr:    nil,
			authEncryptStrTimes:  0,
		}, {
			name:                 "decrypt cursor failure",
			path:                 "/balance-all-crypto/decrypt-cursor-failure",
			query:                fmt.Sprintf(testCryptoQuery["balanceAllCrypto"], "page-cursor", 3),
			expectErr:            true,
			accDetails:           accDetails,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			authDecryptStrErr:    errors.New("decrypt failure"),
			authDecryptStrTimes:  1,
			fiatBalanceErr:       nil,
			fiatBalanceTimes:     0,
			authEncryptStrErr:    nil,
			authEncryptStrTimes:  0,
		}, {
			name:                 "known db error",
			path:                 "/balance-all-crypto/known-db-error",
			query:                fmt.Sprintf(testCryptoQuery["balanceAllCrypto"], "page-cursor", 3),
			expectErr:            true,
			accDetails:           accDetails,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			authDecryptStrErr:    nil,
			authDecryptStrTimes:  1,
			fiatBalanceErr:       postgres.ErrNotFound,
			fiatBalanceTimes:     1,
			authEncryptStrErr:    nil,
			authEncryptStrTimes:  0,
		}, {
			name:                 "unknown db error",
			path:                 "/balance-all-crypto/unknown-db-error",
			query:                fmt.Sprintf(testCryptoQuery["balanceAllCrypto"], "page-cursor", 3),
			expectErr:            true,
			accDetails:           accDetails,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			authDecryptStrErr:    nil,
			authDecryptStrTimes:  1,
			fiatBalanceErr:       errors.New("unknown db error"),
			fiatBalanceTimes:     1,
			authEncryptStrErr:    nil,
			authEncryptStrTimes:  0,
		}, {
			name:                 "encrypt cursor failure",
			path:                 "/balance-all-crypto/encrypt-cursor-failure",
			query:                fmt.Sprintf(testCryptoQuery["balanceAllCrypto"], "page-cursor", 3),
			expectErr:            true,
			accDetails:           accDetails,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			authDecryptStrErr:    nil,
			authDecryptStrTimes:  1,
			fiatBalanceErr:       nil,
			fiatBalanceTimes:     1,
			authEncryptStrErr:    errors.New("encrypt string error"),
			authEncryptStrTimes:  1,
		}, {
			name:                 "valid without query and 10 records",
			path:                 "/balance-all-crypto/valid-no-query-10-records",
			query:                testCryptoQuery["balanceAllCryptoNoParams"],
			expectErr:            false,
			accDetails:           []postgres.CryptoAccount{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}},
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			authDecryptStrErr:    nil,
			authDecryptStrTimes:  0,
			fiatBalanceErr:       nil,
			fiatBalanceTimes:     1,
			authEncryptStrErr:    nil,
			authEncryptStrTimes:  0,
		}, {
			name:                 "valid without query and 11 records",
			path:                 "/balance-all-crypto/valid-no-query-11-records",
			query:                testCryptoQuery["balanceAllCryptoNoParams"],
			expectErr:            false,
			accDetails:           []postgres.CryptoAccount{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}},
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			authDecryptStrErr:    nil,
			authDecryptStrTimes:  0,
			fiatBalanceErr:       nil,
			fiatBalanceTimes:     1,
			authEncryptStrErr:    nil,
			authEncryptStrTimes:  1,
		}, {
			name:                 "valid without query",
			path:                 "/balance-all-crypto/valid-no-query",
			query:                testCryptoQuery["balanceAllCryptoNoParams"],
			expectErr:            false,
			accDetails:           accDetails,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			authDecryptStrErr:    nil,
			authDecryptStrTimes:  0,
			fiatBalanceErr:       nil,
			fiatBalanceTimes:     1,
			authEncryptStrErr:    nil,
			authEncryptStrTimes:  0,
		}, {
			name:                 "valid",
			path:                 "/balance-all-crypto/valid",
			query:                fmt.Sprintf(testCryptoQuery["balanceAllCrypto"], "page-cursor", 3),
			expectErr:            false,
			accDetails:           accDetails,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			authDecryptStrErr:    nil,
			authDecryptStrTimes:  1,
			fiatBalanceErr:       nil,
			fiatBalanceTimes:     1,
			authEncryptStrErr:    nil,
			authEncryptStrTimes:  1,
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
			mockRedis := mocks.NewMockRedis(mockCtrl)    // not called.
			mockQuotes := quotes.NewMockQuotes(mockCtrl) // not called.

			gomock.InOrder(
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(uuid.UUID{}, int64(0), test.authValidateJWTErr).
					Times(test.authValidateJWTTimes),

				mockAuth.EXPECT().DecryptFromString(gomock.Any()).
					Return([]byte{}, test.authDecryptStrErr).
					Times(test.authDecryptStrTimes),

				mockPostgres.EXPECT().CryptoBalancesPaginated(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(test.accDetails, test.fiatBalanceErr).
					Times(test.fiatBalanceTimes),

				mockAuth.EXPECT().EncryptToString(gomock.Any()).
					Return("encrypted-page-cursor", test.authEncryptStrErr).
					Times(test.authEncryptStrTimes),
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

func TestCryptoResolver_TransactionDetailsCrypto(t *testing.T) { //nolint:dupl
	t.Parallel()

	clientID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate client id.")

	txUUID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate transaction id.")

	txID := txUUID.String()

	testCases := []struct {
		name                 string
		path                 string
		query                string
		expectErr            bool
		authValidateJWTErr   error
		authValidateTimes    int
		fiatTxDetailsErr     error
		fiatTxDetailsTimes   int
		cryptoTxDetailsErr   error
		cryptoTxDetailsTimes int
	}{
		{
			name:                 "invalid jwt",
			path:                 "/transaction-details-crypto/invalid-jwt",
			query:                fmt.Sprintf(testCryptoQuery["transactionDetailsCrypto"], txID),
			expectErr:            true,
			authValidateTimes:    1,
			authValidateJWTErr:   errors.New("invalid jwt"),
			fiatTxDetailsErr:     nil,
			fiatTxDetailsTimes:   0,
			cryptoTxDetailsErr:   nil,
			cryptoTxDetailsTimes: 0,
		}, {
			name:                 "db failure fiat",
			path:                 "/transaction-details-crypto/db-failure-fiat",
			query:                fmt.Sprintf(testCryptoQuery["transactionDetailsCrypto"], txID),
			expectErr:            false,
			authValidateTimes:    1,
			authValidateJWTErr:   nil,
			fiatTxDetailsTimes:   1,
			fiatTxDetailsErr:     postgres.ErrTransactCryptoDetails,
			cryptoTxDetailsTimes: 0,
			cryptoTxDetailsErr:   nil,
		}, {
			name:                 "db failure crypto",
			path:                 "/transaction-details-crypto/db-failure-crypto",
			query:                fmt.Sprintf(testCryptoQuery["transactionDetailsCrypto"], txID),
			expectErr:            false,
			authValidateTimes:    1,
			authValidateJWTErr:   nil,
			fiatTxDetailsTimes:   1,
			fiatTxDetailsErr:     nil,
			cryptoTxDetailsTimes: 1,
			cryptoTxDetailsErr:   postgres.ErrTransactCryptoDetails,
		}, {
			name:                 "valid",
			path:                 "/transaction-details-crypto/valid",
			query:                fmt.Sprintf(testCryptoQuery["transactionDetailsCrypto"], txID),
			expectErr:            false,
			authValidateTimes:    1,
			authValidateJWTErr:   nil,
			fiatTxDetailsTimes:   1,
			fiatTxDetailsErr:     nil,
			cryptoTxDetailsTimes: 1,
			cryptoTxDetailsErr:   nil,
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
			mockRedis := mocks.NewMockRedis(mockCtrl)    // Not called.
			mockQuotes := quotes.NewMockQuotes(mockCtrl) // Not called.

			gomock.InOrder(
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(clientID, int64(0), test.authValidateJWTErr).
					Times(test.authValidateTimes),

				mockPostgres.EXPECT().FiatTxDetails(gomock.Any(), gomock.Any()).
					Return([]postgres.FiatJournal{{}}, test.fiatTxDetailsErr).
					Times(test.fiatTxDetailsTimes),

				mockPostgres.EXPECT().CryptoTxDetails(gomock.Any(), gomock.Any()).
					Return([]postgres.CryptoJournal{{}}, test.cryptoTxDetailsErr).
					Times(test.cryptoTxDetailsTimes),
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
