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
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/quotes"
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

func TestFiatResolver_FiatDepositResponseResolver(t *testing.T) {
	t.Parallel()

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

	resolver := fiatDepositResponseResolver{}

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
			name:                 "empty request",
			path:                 "/deposit-fiat/empty-request",
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
			authValidateJWTTimes: 0,
			fiatDepositAccErr:    nil,
			fiatDepositAccTimes:  0,
		}, {
			name:                 "too many decimal places",
			path:                 "/deposit-fiat/too-many-decimal-places",
			query:                fmt.Sprintf(testFiatQuery["depositFiat"], 1234.567, "USD"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 0,
			fiatDepositAccErr:    nil,
			fiatDepositAccTimes:  0,
		}, {
			name:                 "negative amount",
			path:                 "/deposit-fiat/negative-amount",
			query:                fmt.Sprintf(testFiatQuery["depositFiat"], -1234.56, "USD"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 0,
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
