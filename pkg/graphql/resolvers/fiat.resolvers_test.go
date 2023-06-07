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
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/quotes"
	"github.com/surahman/FTeX/pkg/redis"
)

func TestFiatResolver_OpenFiat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		path                 string
		query                string
		expectErr            bool
		authValidateJWTErr   error
		authValidateJWTTimes int
		fiatCreateAccErr     error
		fiatCreateAccTimes   int
	}{
		{
			name:                 "empty request",
			path:                 "/open-fiat/empty-request",
			query:                fmt.Sprintf(testFiatQuery["openFiat"], ""),
			expectErr:            true,
			authValidateJWTErr:   errors.New("invalid token"),
			authValidateJWTTimes: 1,
			fiatCreateAccErr:     nil,
			fiatCreateAccTimes:   0,
		}, {
			name:                 "invalid currency",
			path:                 "/open-fiat/invalid-currency",
			query:                fmt.Sprintf(testFiatQuery["openFiat"], "INVALID"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatCreateAccErr:     nil,
			fiatCreateAccTimes:   0,
		}, {
			name:                 "db failure",
			path:                 "/open-fiat/db-failure",
			query:                fmt.Sprintf(testFiatQuery["openFiat"], "CAD"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatCreateAccErr:     postgres.ErrNotFound,
			fiatCreateAccTimes:   1,
		}, {
			name:                 "valid",
			path:                 "/open-fiat/valid",
			query:                fmt.Sprintf(testFiatQuery["openFiat"], "USD"),
			expectErr:            false,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatCreateAccErr:     nil,
			fiatCreateAccTimes:   1,
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
					Return(uuid.UUID{}, int64(-1), test.authValidateJWTErr).
					Times(test.authValidateJWTTimes),

				mockPostgres.EXPECT().FiatCreateAccount(gomock.Any(), gomock.Any()).
					Return(test.fiatCreateAccErr).
					Times(test.fiatCreateAccTimes),
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

func TestFiatResolver_FiatDepositResponseResolvers(t *testing.T) {
	t.Parallel()

	resolver := fiatDepositResponseResolver{}

	txID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate TxID.")

	clientID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate Client ID.")

	txTS := pgtype.Timestamptz{}
	require.NoError(t, txTS.Scan(time.Now()), "failed to generate Tx timestamp.")

	balanceFloat64 := 12345.67
	balance := decimal.NewFromFloat(balanceFloat64)

	lastTxFloat64 := 8910.11
	lastTx := decimal.NewFromFloat(lastTxFloat64)

	input := &postgres.FiatAccountTransferResult{
		TxID:     txID,
		ClientID: clientID,
		TxTS:     txTS,
		Balance:  balance,
		LastTx:   lastTx,
		Currency: postgres.CurrencyUSD,
	}

	t.Run("TxID", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.TxID(context.TODO(), input)
		require.NoError(t, err, "failed to resolve tx id.")
		require.Equal(t, txID.String(), result, "tx id mismatched.")
	})

	t.Run("ClientID", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.ClientID(context.TODO(), input)
		require.NoError(t, err, "failed to resolve client id.")
		require.Equal(t, clientID.String(), result, "client id mismatched.")
	})

	t.Run("TxTimestamp", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.TxTimestamp(context.TODO(), input)
		require.NoError(t, err, "failed to resolve tx timestamp.")
		require.Equal(t, txTS.Time.String(), result, "tx timestamp mismatched.")
	})

	t.Run("Balance", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.Balance(context.TODO(), input)
		require.NoError(t, err, "failed to resolve balance")
		require.Equal(t, balance.String(), result, "balance mismatched.")
	})

	t.Run("LastTx", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.LastTx(context.TODO(), input)
		require.NoError(t, err, "failed to resolve lastTx")
		require.Equal(t, lastTx.String(), result, "lastTx mismatched.")
	})

	t.Run("Currency", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.Currency(context.TODO(), input)
		require.NoError(t, err, "failed to resolve currency")
		require.Equal(t, "USD", result, "currency mismatched.")
	})
}

func TestFiatResolver_DepositFiat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		path                 string
		query                string
		expectErr            bool
		authValidateJWTErr   error
		authValidateJWTTimes int
		fiatDepositAccErr    error
		fiatDepositAccTimes  int
	}{
		{
			name:                 "invalid jwt",
			path:                 "/deposit-fiat/invalid-jwt",
			query:                fmt.Sprintf(testFiatQuery["depositFiat"], 1234.56, "USD"),
			expectErr:            true,
			authValidateJWTErr:   errors.New("authorization failure"),
			authValidateJWTTimes: 1,
			fiatDepositAccErr:    nil,
			fiatDepositAccTimes:  0,
		}, {
			name:                 "invalid currency",
			path:                 "/deposit-fiat/invalid-currency",
			query:                fmt.Sprintf(testFiatQuery["depositFiat"], 1234.56, "INVALID"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatDepositAccErr:    nil,
			fiatDepositAccTimes:  0,
		}, {
			name:                 "too many decimal places",
			path:                 "/deposit-fiat/too-many-decimal-places",
			query:                fmt.Sprintf(testFiatQuery["depositFiat"], 1234.567, "USD"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatDepositAccErr:    nil,
			fiatDepositAccTimes:  0,
		}, {
			name:                 "negative amount",
			path:                 "/deposit-fiat/negative-amount",
			query:                fmt.Sprintf(testFiatQuery["depositFiat"], -1234.56, "USD"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatDepositAccErr:    nil,
			fiatDepositAccTimes:  0,
		}, {
			name:                 "valid",
			path:                 "/deposit-fiat/valid",
			query:                fmt.Sprintf(testFiatQuery["depositFiat"], 1234.56, "USD"),
			expectErr:            false,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatDepositAccErr:    nil,
			fiatDepositAccTimes:  1,
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
					Return(uuid.UUID{}, int64(-1), test.authValidateJWTErr).
					Times(test.authValidateJWTTimes),

				mockPostgres.EXPECT().FiatExternalTransfer(gomock.Any(), gomock.Any()).
					Return(&postgres.FiatAccountTransferResult{}, test.fiatDepositAccErr).
					Times(test.fiatDepositAccTimes),
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

func TestFiatResolver_FiatExchangeOfferRequestResolver(t *testing.T) {
	t.Parallel()

	resolver := fiatExchangeOfferRequestResolver{}
	expected := 9876.54

	exchangeOfferRequest := &models.HTTPExchangeOfferRequest{
		SourceCurrency:      "",
		DestinationCurrency: "",
		SourceAmount:        decimal.NewFromFloat(123456.78),
	}

	t.Run("SourceAmount", func(t *testing.T) {
		t.Parallel()

		err := resolver.SourceAmount(context.TODO(), exchangeOfferRequest, expected)
		require.NoError(t, err, "failed to resolve debit amount")
		require.InDelta(t, expected, exchangeOfferRequest.SourceAmount.InexactFloat64(), 0.01, "debit amount mismatched.")
	})
}

func TestFiatResolver_FiatExchangeOfferResponseResolver(t *testing.T) {
	t.Parallel()

	resolver := fiatExchangeOfferResponseResolver{}

	debitAmount := decimal.NewFromFloat(123456.78)

	exchangeOfferResponse := &models.HTTPExchangeOfferResponse{
		PriceQuote:  models.PriceQuote{},
		DebitAmount: debitAmount,
		OfferID:     "",
		Expires:     0,
	}

	t.Run("DebitAmount", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.DebitAmount(context.TODO(), exchangeOfferResponse)
		require.NoError(t, err, "failed to resolve debit amount")
		require.InDelta(t, debitAmount.InexactFloat64(), result, 0.01, "debit amount mismatched.")
	})
}

func TestFiatResolver_ExchangeOfferFiat(t *testing.T) {
	t.Parallel()

	amountValid, err := decimal.NewFromString("999")
	require.NoError(t, err, "failed to convert valid amount")

	testCases := []struct {
		name                 string
		path                 string
		query                string
		expectErr            bool
		authValidateJWTErr   error
		authValidateJWTTimes int
		quotesErr            error
		quotesTimes          int
		authEncryptErr       error
		authEncryptTimes     int
		redisErr             error
		redisTimes           int
	}{
		{
			name:                 "invalid jwt",
			path:                 "/exchange-offer-fiat/invalid-jwt",
			query:                fmt.Sprintf(testFiatQuery["exchangeOfferFiat"], "USD", "CAD", 101.11),
			expectErr:            true,
			authValidateJWTErr:   errors.New("invalid jwt"),
			authValidateJWTTimes: 1,
			quotesErr:            nil,
			quotesTimes:          0,
			authEncryptErr:       nil,
			authEncryptTimes:     0,
			redisErr:             nil,
			redisTimes:           0,
		}, {
			name:                 "empty request",
			path:                 "/exchange-offer-fiat/empty-request",
			query:                fmt.Sprintf(testFiatQuery["exchangeOfferFiat"], "", "", 101.11),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			quotesErr:            nil,
			quotesTimes:          0,
			authEncryptErr:       nil,
			authEncryptTimes:     0,
			redisErr:             nil,
			redisTimes:           0,
		}, {
			name:                 "invalid source currency",
			path:                 "/exchange-offer-fiat/invalid-src-currency",
			query:                fmt.Sprintf(testFiatQuery["exchangeOfferFiat"], "INVALID", "CAD", 101.11),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			quotesErr:            nil,
			quotesTimes:          0,
			authEncryptErr:       nil,
			authEncryptTimes:     0,
			redisErr:             nil,
			redisTimes:           0,
		}, {
			name:                 "invalid destination currency",
			path:                 "/exchange-offer-fiat/invalid-dst-currency",
			query:                fmt.Sprintf(testFiatQuery["exchangeOfferFiat"], "USD", "INVALID", 101.11),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			quotesErr:            nil,
			quotesTimes:          0,
			authEncryptErr:       nil,
			authEncryptTimes:     0,
			redisErr:             nil,
			redisTimes:           0,
		}, {
			name:                 "too many decimal places",
			path:                 "/exchange-offer-fiat/too-many-decimal-places",
			query:                fmt.Sprintf(testFiatQuery["exchangeOfferFiat"], "USD", "CAD", 101.111),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			quotesErr:            nil,
			quotesTimes:          0,
			authEncryptErr:       nil,
			authEncryptTimes:     0,
			redisErr:             nil,
			redisTimes:           0,
		}, {
			name:                 "negative",
			path:                 "/exchange-offer-fiat/negative",
			query:                fmt.Sprintf(testFiatQuery["exchangeOfferFiat"], "USD", "CAD", -101.11),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			quotesErr:            nil,
			quotesTimes:          0,
			authEncryptErr:       nil,
			authEncryptTimes:     0,
			redisErr:             nil,
			redisTimes:           0,
		}, {
			name:                 "fiat conversion error",
			path:                 "/exchange-offer-fiat/fiat-conversion-error",
			query:                fmt.Sprintf(testFiatQuery["exchangeOfferFiat"], "USD", "CAD", 101.11),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			quotesErr:            errors.New(""),
			quotesTimes:          1,
			authEncryptErr:       nil,
			authEncryptTimes:     0,
			redisErr:             nil,
			redisTimes:           0,
		}, {
			name:                 "encryption error",
			path:                 "/exchange-offer-fiat/encryption-error",
			query:                fmt.Sprintf(testFiatQuery["exchangeOfferFiat"], "USD", "CAD", 101.11),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			quotesErr:            nil,
			quotesTimes:          1,
			authEncryptErr:       errors.New(""),
			authEncryptTimes:     1,
			redisErr:             nil,
			redisTimes:           0,
		}, {
			name:                 "redis error",
			path:                 "/exchange-offer-fiat/redis-error",
			query:                fmt.Sprintf(testFiatQuery["exchangeOfferFiat"], "USD", "CAD", 101.11),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			quotesErr:            nil,
			quotesTimes:          1,
			authEncryptErr:       nil,
			authEncryptTimes:     1,
			redisErr:             errors.New(""),
			redisTimes:           1,
		}, {
			name:                 "valid",
			path:                 "/exchange-offer-fiat/valid",
			query:                fmt.Sprintf(testFiatQuery["exchangeOfferFiat"], "USD", "CAD", 101.11),
			expectErr:            false,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			quotesErr:            nil,
			quotesTimes:          1,
			authEncryptErr:       nil,
			authEncryptTimes:     1,
			redisErr:             nil,
			redisTimes:           1,
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
					Times(test.authValidateJWTTimes),

				mockQuotes.EXPECT().FiatConversion(gomock.Any(), gomock.Any(), gomock.Any(), nil).
					Return(amountValid, amountValid, test.quotesErr).
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

func TestFiatResolver_FiatExchangeTransferResponseResolver(t *testing.T) {
	t.Parallel()

	resolver := fiatExchangeTransferResponseResolver{}

	response := &models.HTTPFiatTransferResponse{
		SrcTxReceipt: &postgres.FiatAccountTransferResult{
			TxID:     uuid.UUID{},
			ClientID: uuid.UUID{},
			TxTS:     pgtype.Timestamptz{},
			Balance:  decimal.Decimal{},
			LastTx:   decimal.Decimal{},
			Currency: "",
		},
		DstTxReceipt: &postgres.FiatAccountTransferResult{
			TxID:     uuid.UUID{},
			ClientID: uuid.UUID{},
			TxTS:     pgtype.Timestamptz{},
			Balance:  decimal.Decimal{},
			LastTx:   decimal.Decimal{},
			Currency: "",
		},
	}

	source, err := resolver.SourceReceipt(context.TODO(), response)
	require.NoError(t, err, "source should always return a nil error.")
	require.Equal(t, response.SrcTxReceipt, source, "source and returned struct addresses mismatched.")

	destination, err := resolver.SourceReceipt(context.TODO(), response)
	require.NoError(t, err, "destinations should always return a nil error.")
	require.Equal(t, response.DstTxReceipt, destination, "destination and returned struct addresses mismatched.")
}

func TestFiatResolver_ExchangeTransferFiat(t *testing.T) { //nolint:maintidx
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
		name                 string
		path                 string
		query                string
		expectErr            bool
		authValidateJWTErr   error
		authValidateJWTTimes int
		authDecryptErr       error
		authDecryptTimes     int
		redisGetData         models.HTTPExchangeOfferResponse
		redisGetErr          error
		redisGetTimes        int
		redisDelErr          error
		redisDelTimes        int
		internalXferErr      error
		internalXferTimes    int
	}{
		{
			name:                 "invalid JWT",
			path:                 "/exchange-xfer-fiat/invalid-jwt",
			query:                fmt.Sprintf(testFiatQuery["exchangeTransferFiat"], "some-encrypted-offer-id"),
			expectErr:            true,
			authValidateJWTErr:   errors.New("bad auth"),
			authValidateJWTTimes: 1,
			authDecryptErr:       nil,
			authDecryptTimes:     0,
			redisGetData:         validOffer,
			redisGetErr:          nil,
			redisGetTimes:        0,
			redisDelErr:          nil,
			redisDelTimes:        0,
			internalXferErr:      nil,
			internalXferTimes:    0,
		}, {
			name:                 "decrypt offer ID",
			path:                 "/exchange-xfer-fiat/decrypt-offer-id",
			query:                fmt.Sprintf(testFiatQuery["exchangeTransferFiat"], "some-encrypted-offer-id"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			authDecryptErr:       errors.New("decrypt offer id"),
			authDecryptTimes:     1,
			redisGetData:         validOffer,
			redisGetErr:          nil,
			redisGetTimes:        0,
			redisDelErr:          nil,
			redisDelTimes:        0,
			internalXferErr:      nil,
			internalXferTimes:    0,
		}, {
			name:                 "cache unknown error",
			path:                 "/exchange-xfer-fiat/cache-unknown-error",
			query:                fmt.Sprintf(testFiatQuery["exchangeTransferFiat"], "some-encrypted-offer-id"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			authDecryptErr:       nil,
			authDecryptTimes:     1,
			redisGetData:         validOffer,
			redisGetErr:          errors.New("unknown error"),
			redisGetTimes:        1,
			redisDelErr:          nil,
			redisDelTimes:        0,
			internalXferErr:      nil,
			internalXferTimes:    0,
		}, {
			name:                 "cache expired",
			path:                 "/exchange-xfer-fiat/cache-expired",
			query:                fmt.Sprintf(testFiatQuery["exchangeTransferFiat"], "some-encrypted-offer-id"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			authDecryptErr:       nil,
			authDecryptTimes:     1,
			redisGetData:         validOffer,
			redisGetErr:          redis.ErrCacheMiss,
			redisGetTimes:        1,
			redisDelErr:          nil,
			redisDelTimes:        0,
			internalXferErr:      nil,
			internalXferTimes:    0,
		}, {
			name:                 "cache del failure",
			path:                 "/exchange-xfer-fiat/cache-del-failure",
			query:                fmt.Sprintf(testFiatQuery["exchangeTransferFiat"], "some-encrypted-offer-id"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			authDecryptErr:       nil,
			authDecryptTimes:     1,
			redisGetData:         validOffer,
			redisGetErr:          nil,
			redisGetTimes:        1,
			redisDelErr:          redis.ErrCacheUnknown,
			redisDelTimes:        1,
			internalXferErr:      nil,
			internalXferTimes:    0,
		}, {
			name:                 "cache del expired",
			path:                 "/exchange-xfer-fiat/cache-del-expired",
			query:                fmt.Sprintf(testFiatQuery["exchangeTransferFiat"], "some-encrypted-offer-id"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			authDecryptErr:       nil,
			authDecryptTimes:     1,
			redisGetData:         validOffer,
			redisGetErr:          nil,
			redisGetTimes:        1,
			redisDelErr:          redis.ErrCacheMiss,
			redisDelTimes:        1,
			internalXferErr:      nil,
			internalXferTimes:    1,
		}, {
			name:                 "client id mismatch",
			path:                 "/exchange-xfer-fiat/client-id-mismatch",
			query:                fmt.Sprintf(testFiatQuery["exchangeTransferFiat"], "some-encrypted-offer-id"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			authDecryptErr:       nil,
			authDecryptTimes:     1,
			redisGetData:         invalidOfferClientID,
			redisGetErr:          nil,
			redisGetTimes:        1,
			redisDelErr:          nil,
			redisDelTimes:        1,
			internalXferErr:      nil,
			internalXferTimes:    0,
		}, {
			name:                 "invalid source destination amount",
			path:                 "/exchange-xfer-fiat/invalid-source-destination-amount",
			query:                fmt.Sprintf(testFiatQuery["exchangeTransferFiat"], "some-encrypted-offer-id"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			authDecryptErr:       nil,
			authDecryptTimes:     1,
			redisGetData:         invalidOfferSource,
			redisGetErr:          nil,
			redisGetTimes:        1,
			redisDelErr:          nil,
			redisDelTimes:        1,
			internalXferErr:      nil,
			internalXferTimes:    0,
		}, {
			name:                 "transaction failure",
			path:                 "/exchange-xfer-fiat/transaction-failure",
			query:                fmt.Sprintf(testFiatQuery["exchangeTransferFiat"], "some-encrypted-offer-id"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			authDecryptErr:       nil,
			authDecryptTimes:     1,
			redisGetData:         validOffer,
			redisGetErr:          nil,
			redisGetTimes:        1,
			redisDelErr:          nil,
			redisDelTimes:        1,
			internalXferErr:      errors.New("transaction failure"),
			internalXferTimes:    1,
		}, {
			name:                 "valid",
			path:                 "/exchange-transfer-fiat/valid",
			query:                fmt.Sprintf(testFiatQuery["exchangeTransferFiat"], "some-encrypted-offer-id"),
			expectErr:            false,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			authDecryptErr:       nil,
			authDecryptTimes:     1,
			redisGetData:         validOffer,
			redisGetErr:          nil,
			redisGetTimes:        1,
			redisDelErr:          nil,
			redisDelTimes:        1,
			internalXferErr:      nil,
			internalXferTimes:    1,
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
			mockQuotes := quotes.NewMockQuotes(mockCtrl) // not called.

			gomock.InOrder(
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(validClientID, int64(0), test.authValidateJWTErr).
					Times(test.authValidateJWTTimes),

				mockAuth.EXPECT().DecryptFromString(gomock.Any()).
					Return(validOfferID, test.authDecryptErr).
					Times(test.authDecryptTimes),

				mockRedis.EXPECT().Get(gomock.Any(), gomock.Any()).
					Return(test.redisGetErr).
					SetArg(1, test.redisGetData).
					Times(test.redisGetTimes),

				mockRedis.EXPECT().Del(gomock.Any()).
					Return(test.redisDelErr).
					Times(test.redisDelTimes),

				mockPostgres.EXPECT().FiatInternalTransfer(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil, test.internalXferErr).
					Times(test.internalXferTimes),
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

func TestFiatResolver_FiatAccountResolvers(t *testing.T) {
	t.Parallel()

	resolver := fiatAccountResolver{}

	clientID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate client id.")

	balanceAmount := decimal.NewFromFloat(123456.78)
	lastTxAmount := decimal.NewFromFloat(91011.12)

	lastTxTS := time.Now().Add(-15 * time.Second)
	lastTxTSPG := pgtype.Timestamptz{}
	require.NoError(t, lastTxTSPG.Scan(lastTxTS), "failed to generate lastTxTs.")

	createdAt := time.Now().Add(-15 * time.Minute)
	createdAtPG := pgtype.Timestamptz{}
	require.NoError(t, createdAtPG.Scan(createdAt), "failed to generate createdAt.")

	fiatAccount := &postgres.FiatAccount{
		Currency:  postgres.CurrencyUSD,
		Balance:   balanceAmount,
		LastTx:    lastTxAmount,
		LastTxTs:  lastTxTSPG,
		CreatedAt: createdAtPG,
		ClientID:  clientID,
	}

	t.Run("Currency", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.Currency(context.TODO(), fiatAccount)
		require.NoError(t, err, "failed to resolve currency")
		require.Equal(t, "USD", result, "currency mismatched.")
	})

	t.Run("BalanceAmount", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.Balance(context.TODO(), fiatAccount)
		require.NoError(t, err, "failed to resolve balance amount")
		require.InDelta(t, balanceAmount.InexactFloat64(), result, 0.01, "balance amount mismatched.")
	})

	t.Run("LastTxAmount", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.LastTx(context.TODO(), fiatAccount)
		require.NoError(t, err, "failed to resolve lastTx amount")
		require.InDelta(t, lastTxAmount.InexactFloat64(), result, 0.01, "lastTx amount mismatched.")
	})

	t.Run("LastTxTS", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.LastTxTs(context.TODO(), fiatAccount)
		require.NoError(t, err, "failed to resolve lastTx timestamp")
		require.Equal(t, lastTxTS.String(), result, "lastTx timestamp mismatched.")
	})

	t.Run("CreatedAtTS", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.CreatedAt(context.TODO(), fiatAccount)
		require.NoError(t, err, "failed to resolve created at timestamp.")
		require.Equal(t, createdAt.String(), result, "created at timestamp mismatched.")
	})
}

func TestFiatResolver_BalanceFiat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		path                 string
		query                string
		expectErr            bool
		authValidateJWTErr   error
		authValidateJWTTimes int
		fiatBalanceErr       error
		fiatBalanceTimes     int
	}{
		{
			name:                 "invalid JWT",
			path:                 "/balance-fiat/invalid-jwt",
			query:                fmt.Sprintf(testFiatQuery["balanceFiat"], "USD"),
			authValidateJWTErr:   errors.New("invalid JWT"),
			authValidateJWTTimes: 1,
			fiatBalanceErr:       nil,
			fiatBalanceTimes:     0,
		}, {
			name:                 "invalid currency",
			path:                 "/balance-fiat/invalid-currency",
			query:                fmt.Sprintf(testFiatQuery["balanceFiat"], "INVALID"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatBalanceErr:       nil,
			fiatBalanceTimes:     0,
		}, {
			name:                 "unknown db error",
			path:                 "/balance-fiat/unknown-db-error",
			query:                fmt.Sprintf(testFiatQuery["balanceFiat"], "USD"),
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatBalanceErr:       errors.New("unknown error"),
			fiatBalanceTimes:     1,
		}, {
			name:                 "known db error",
			path:                 "/balance-fiat/known-db-error",
			query:                fmt.Sprintf(testFiatQuery["balanceFiat"], "USD"),
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatBalanceErr:       postgres.ErrNotFound,
			fiatBalanceTimes:     1,
		}, {
			name:                 "valid",
			path:                 "/balance-fiat/valid",
			query:                fmt.Sprintf(testFiatQuery["balanceFiat"], "USD"),
			expectErr:            false,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatBalanceErr:       nil,
			fiatBalanceTimes:     1,
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

				mockPostgres.EXPECT().FiatBalance(gomock.Any(), gomock.Any()).
					Return(postgres.FiatAccount{}, test.fiatBalanceErr).
					Times(test.fiatBalanceTimes),
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

func TestFiatResolver_BalanceAllFiat(t *testing.T) {
	t.Parallel()

	accDetails := []postgres.FiatAccount{{}, {}, {}, {}}

	testCases := []struct {
		name                 string
		path                 string
		query                string
		expectErr            bool
		accDetails           []postgres.FiatAccount
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
			path:                 "/balance-all-fiat/invalid-jwt",
			query:                fmt.Sprintf(testFiatQuery["balanceAllFiat"], "page-cursor", 3),
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
			path:                 "/balance-all-fiat/decrypt-cursor-failure",
			query:                fmt.Sprintf(testFiatQuery["balanceAllFiat"], "page-cursor", 3),
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
			path:                 "/balance-all-fiat/known-db-error",
			query:                fmt.Sprintf(testFiatQuery["balanceAllFiat"], "page-cursor", 3),
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
			path:                 "/balance-all-fiat/unknown-db-error",
			query:                fmt.Sprintf(testFiatQuery["balanceAllFiat"], "page-cursor", 3),
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
			path:                 "/balance-all-fiat/encrypt-cursor-failure",
			query:                fmt.Sprintf(testFiatQuery["balanceAllFiat"], "page-cursor", 3),
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
			path:                 "/balance-all-fiat/valid-no-query-10-records",
			query:                testFiatQuery["balanceAllFiatNoParams"],
			expectErr:            false,
			accDetails:           []postgres.FiatAccount{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}},
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
			path:                 "/balance-all-fiat/valid-no-query-11-records",
			query:                testFiatQuery["balanceAllFiatNoParams"],
			expectErr:            false,
			accDetails:           []postgres.FiatAccount{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}},
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
			path:                 "/balance-all-fiat/valid-no-query",
			query:                testFiatQuery["balanceAllFiatNoParams"],
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
			path:                 "/balance-all-fiat/valid",
			query:                fmt.Sprintf(testFiatQuery["balanceAllFiat"], "page-cursor", 3),
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

				mockPostgres.EXPECT().FiatBalancePaginated(gomock.Any(), gomock.Any(), gomock.Any()).
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

func TestFiatResolver_FiatJournalResolvers(t *testing.T) {
	t.Parallel()

	resolver := fiatJournalResolver{}

	amount := decimal.NewFromFloat(123456.78)

	timestamp := time.Now()
	transactedAt := pgtype.Timestamptz{}
	require.NoError(t, transactedAt.Scan(timestamp), "failed to generate timestamp.")

	clientID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate client id.")

	txID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate tx id.")

	fiatJournal := &postgres.FiatJournal{
		Currency:     postgres.CurrencyUSD,
		Amount:       amount,
		TransactedAt: transactedAt,
		ClientID:     clientID,
		TxID:         txID,
	}

	t.Run("Currency", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.Currency(context.TODO(), fiatJournal)
		require.NoError(t, err, "failed to resolve currency")
		require.Equal(t, string(postgres.CurrencyUSD), result, "currency mismatched.")
	})

	t.Run("Amount", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.Amount(context.TODO(), fiatJournal)
		require.NoError(t, err, "failed to resolve amount")
		require.InDeltaf(t, amount.InexactFloat64(), result, 0.01, "amount mismatched.")
	})

	t.Run("TransactedAt", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.TransactedAt(context.TODO(), fiatJournal)
		require.NoError(t, err, "failed to resolve transacted at timestamp.")
		require.Equal(t, timestamp.String(), result, "transacted at timestamp mismatched.")
	})

	t.Run("ClientID", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.ClientID(context.TODO(), fiatJournal)
		require.NoError(t, err, "failed to resolve client id.")
		require.Equal(t, clientID.String(), result, "client id mismatched.")
	})

	t.Run("TxID", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.TxID(context.TODO(), fiatJournal)
		require.NoError(t, err, "failed to resolve tx id.")
		require.Equal(t, txID.String(), result, "tx id mismatched.")
	})
}

func TestFiatResolver_TransactionDetailsFiat(t *testing.T) {
	t.Parallel()

	txID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate new UUID.")

	journalEntries := []postgres.FiatJournal{{}}

	testCases := []struct {
		name                 string
		path                 string
		query                string
		expectErr            bool
		journalEntries       []postgres.FiatJournal
		authValidateJWTErr   error
		authValidateJWTTimes int
		fiatTxDetailsErr     error
		fiatTxDetailsTimes   int
	}{
		{
			name:                 "invalid JWT",
			path:                 "/transaction-details-fiat/invalid-jwt",
			query:                fmt.Sprintf(testFiatQuery["transactionDetailsFiat"], txID),
			expectErr:            true,
			journalEntries:       journalEntries,
			authValidateJWTErr:   errors.New("invalid JWT"),
			authValidateJWTTimes: 1,
			fiatTxDetailsErr:     nil,
			fiatTxDetailsTimes:   0,
		}, {
			name:                 "invalid transaction ID",
			path:                 "/transaction-details-fiat/invalid-transaction-id",
			query:                fmt.Sprintf(testFiatQuery["transactionDetailsFiat"], "invalid-tx-id"),
			expectErr:            true,
			journalEntries:       journalEntries,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatTxDetailsErr:     nil,
			fiatTxDetailsTimes:   0,
		}, {
			name:                 "unknown db error",
			path:                 "/transaction-details-fiat/unknown-db-error",
			query:                fmt.Sprintf(testFiatQuery["transactionDetailsFiat"], txID),
			expectErr:            true,
			journalEntries:       journalEntries,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatTxDetailsErr:     errors.New("unknown error"),
			fiatTxDetailsTimes:   1,
		}, {
			name:                 "known db error",
			path:                 "/transaction-details-fiat/known-db-error",
			query:                fmt.Sprintf(testFiatQuery["transactionDetailsFiat"], txID),
			expectErr:            true,
			journalEntries:       journalEntries,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatTxDetailsErr:     postgres.ErrNotFound,
			fiatTxDetailsTimes:   1,
		}, {
			name:                 "transaction id not found",
			path:                 "/transaction-details-fiat/transaction-id-not-found",
			query:                fmt.Sprintf(testFiatQuery["transactionDetailsFiat"], txID),
			expectErr:            true,
			journalEntries:       nil,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatTxDetailsErr:     nil,
			fiatTxDetailsTimes:   1,
		}, {
			name:                 "valid",
			path:                 "/transaction-details-fiat/valid",
			query:                fmt.Sprintf(testFiatQuery["transactionDetailsFiat"], txID),
			expectErr:            false,
			journalEntries:       journalEntries,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatTxDetailsErr:     nil,
			fiatTxDetailsTimes:   1,
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

				mockPostgres.EXPECT().FiatTxDetails(gomock.Any(), gomock.Any()).
					Return(test.journalEntries, test.fiatTxDetailsErr).
					Times(test.fiatTxDetailsTimes),
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

func TestFiatResolver_TransactionDetailsAllFiat(t *testing.T) {
	t.Parallel()

	decryptedCursor := fmt.Sprintf("%s,%s,%d",
		fmt.Sprintf(constants.GetMonthFormatString(), 2023, 6, "-04:00"),
		fmt.Sprintf(constants.GetMonthFormatString(), 2023, 7, "-04:00"),
		10)

	journalEntries := []postgres.FiatJournal{{}, {}, {}, {}}

	testCases := []struct {
		name                   string
		path                   string
		query                  string
		expectErr              bool
		journalEntries         []postgres.FiatJournal
		authValidateJWTErr     error
		authValidateJWTTimes   int
		authDecryptCursorErr   error
		authDecryptCursorTimes int
		authEncryptCursorErr   error
		authEncryptCursorTimes int
		fiatTxPaginatedErr     error
		fiatTxPaginatedTimes   int
	}{
		{
			name: "auth failure",
			path: "/transaction-details-all-fiat/auth-failure",
			query: fmt.Sprintf(testFiatQuery["transactionDetailsAllFiatSubsequent"],
				"USD", 3, "page-cusror"),
			expectErr:              true,
			journalEntries:         journalEntries,
			authValidateJWTErr:     errors.New("auth failure"),
			authValidateJWTTimes:   1,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 0,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 0,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   0,
		}, {
			name: "bad currency",
			path: "/transaction-details-all-fiat/bad-currency",
			query: fmt.Sprintf(testFiatQuery["transactionDetailsAllFiatSubsequent"],
				"INVALID", 3, "page-cusror"),
			expectErr:              true,
			journalEntries:         journalEntries,
			authValidateJWTErr:     nil,
			authValidateJWTTimes:   1,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 0,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 0,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   0,
		}, {
			name: "no cursor or params",
			path: "/transaction-details-all-fiat/no-cursor-or-params",
			query: fmt.Sprintf(testFiatQuery["transactionDetailsAllFiatSubsequent"],
				"", 3, ""),
			expectErr:              true,
			journalEntries:         journalEntries,
			authValidateJWTErr:     nil,
			authValidateJWTTimes:   1,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 0,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 0,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   0,
		}, {
			name: "db failure",
			path: "/transaction-details-all-fiat/db-failure",
			query: fmt.Sprintf(testFiatQuery["transactionDetailsAllFiatSubsequent"],
				"USD", 3, "page-cusror"),
			expectErr:              true,
			journalEntries:         journalEntries,
			authValidateJWTErr:     nil,
			authValidateJWTTimes:   1,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 1,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     postgres.ErrNotFound,
			fiatTxPaginatedTimes:   1,
		}, {
			name: "unknown db failure",
			path: "/transaction-details-all-fiat/unknown-db-failure",
			query: fmt.Sprintf(testFiatQuery["transactionDetailsAllFiatSubsequent"],
				"USD", 3, "page-cusror"),
			expectErr:              true,
			journalEntries:         journalEntries,
			authValidateJWTErr:     nil,
			authValidateJWTTimes:   1,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 1,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     errors.New("db failure"),
			fiatTxPaginatedTimes:   1,
		}, {
			name: "no transactions",
			path: "/transaction-details-all-fiat/no-transactions",
			query: fmt.Sprintf(testFiatQuery["transactionDetailsAllFiatSubsequent"],
				"USD", 3, "page-cusror"),
			expectErr:              false,
			journalEntries:         []postgres.FiatJournal{},
			authValidateJWTErr:     nil,
			authValidateJWTTimes:   1,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 1,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   1,
		}, {
			name: "valid with cursor",
			path: "/transaction-details-all-fiat/valid-with-cursor",
			query: fmt.Sprintf(testFiatQuery["transactionDetailsAllFiatSubsequent"],
				"USD", 3, "page-cusror"),
			expectErr:              false,
			journalEntries:         journalEntries,
			authValidateJWTErr:     nil,
			authValidateJWTTimes:   1,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 1,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   1,
		}, {
			name: "valid with query",
			path: "/transaction-details-all-fiat/valid-with-query",
			query: fmt.Sprintf(testFiatQuery["transactionDetailsAllFiatInit"],
				"USD", 3, "-04:00", 6, 2023),
			expectErr:              false,
			journalEntries:         journalEntries,
			authValidateJWTErr:     nil,
			authValidateJWTTimes:   1,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 0,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   1,
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
					Return([]byte(decryptedCursor), test.authDecryptCursorErr).
					Times(test.authDecryptCursorTimes),

				mockAuth.EXPECT().EncryptToString(gomock.Any()).
					Return("encrypted-cursor", test.authEncryptCursorErr).
					Times(test.authEncryptCursorTimes),

				mockPostgres.EXPECT().FiatTransactionsPaginated(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any()).
					Return(test.journalEntries, test.fiatTxPaginatedErr).
					Times(test.fiatTxPaginatedTimes),
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

func TestFiatResolver_FiatTransactionsPaginatedResolver(t *testing.T) {
	t.Parallel()

	resolver := fiatTransactionsPaginatedResolver{}

	transactions := &models.HTTPFiatTransactionsPaginated{}

	actual, err := resolver.Transactions(context.TODO(), transactions)
	require.NoError(t, err, "error should always be nil.")
	require.Equal(t, transactions.TransactionDetails, actual, "actual and returned addresses do not match.")
}
