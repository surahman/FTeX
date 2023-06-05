package utilities

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/postgres"
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
