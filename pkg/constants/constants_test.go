package constants

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEtcDir(t *testing.T) {
	require.Equal(t, configEtcDir, EtcDir(), "Incorrect etc directory")
}

func TestHomeDir(t *testing.T) {
	require.Equal(t, configHomeDir, HomeDir(), "Incorrect home directory")
}

func TestBaseDir(t *testing.T) {
	require.Equal(t, configBaseDir, BaseDir(), "Incorrect base directory")
}

func TestGithubCIKey(t *testing.T) {
	require.Equal(t, githubCIKey, GithubCIKey(), "Incorrect Github CI environment key")
}

func TestLoggerFileName(t *testing.T) {
	require.Equal(t, loggerConfigFileName, LoggerFileName(), "Incorrect logger filename")
}

func TestLoggerPrefix(t *testing.T) {
	require.Equal(t, loggerPrefix, LoggerPrefix(), "Incorrect Zap logger environment prefix")
}

func TestPostgresFileName(t *testing.T) {
	require.Equal(t, postgresConfigFileName, PostgresFileName(), "Incorrect Postgres filename")
}

func TestGetPostgresPrefix(t *testing.T) {
	require.Equal(t, postgresPrefix, PostgresPrefix(), "Incorrect Postgres environment prefix")
}

func TestPostgresDSN(t *testing.T) {
	require.Equal(t, postgresDSN, PostgresDSN(), "Incorrect Postgres DSN format string")
}

func TestTestDatabaseName(t *testing.T) {
	require.Equal(t, testDatabaseName, TestDatabaseName(), "Incorrect test suite database name")
}

func TestRedisFileName(t *testing.T) {
	require.Equal(t, redisConfigFileName, RedisFileName(), "Incorrect Redis filename")
}

func TestRedisPrefix(t *testing.T) {
	require.Equal(t, redisPrefix, RedisPrefix(), "Incorrect Redis prefix")
}

func TestQuotesFileName(t *testing.T) {
	require.Equal(t, quotesConfigFileName, QuotesFileName(), "Incorrect Quotes filename")
}

func TestQuotesPrefix(t *testing.T) {
	require.Equal(t, quotesPrefix, QuotesPrefix(), "Incorrect Quotes prefix")
}

func TestAuthFileName(t *testing.T) {
	require.Equal(t, authConfigFileName, AuthFileName(), "Incorrect authentication filename")
}

func TestAuthPrefix(t *testing.T) {
	require.Equal(t, authPrefix, AuthPrefix(), "Incorrect authorization environment prefix")
}

func TestHTTPRESTFileName(t *testing.T) {
	require.Equal(t, restConfigFileName, HTTPRESTFileName(), "Incorrect HTTP REST filename")
}

func TestHTTPRESTPrefix(t *testing.T) {
	require.Equal(t, restPrefix, HTTPRESTPrefix(), "Incorrect HTTP REST environment prefix")
}

func TestDeleteUserAccountConfirmation(t *testing.T) {
	require.Equal(t, deleteUserAccountConfirmation, DeleteUserAccountConfirmation(),
		"Incorrect user account deletion confirmation message.")
}

func TestDecimalPlacesFiat(t *testing.T) {
	require.Equal(t, fiatDecimalPlaces, DecimalPlacesFiat(), "Incorrect Fiat currency decimal places.")
}
func TestDecimalPlacesCrypto(t *testing.T) {
	require.Equal(t, cryptoDecimalPlaces, DecimalPlacesCrypto(), "Incorrect Cryptocurrency decimal places.")
}

func TestFiatOfferTTL(t *testing.T) {
	require.Equal(t, fiatOfferTTL, FiatOfferTTL(), "Incorrect Fiat offer TTL.")
}

func TestCryptoOfferTTL(t *testing.T) {
	require.Equal(t, cryptoOfferTTL, CryptoOfferTTL(), "Incorrect Crypto offer TTL.")
}

func TestMonthFormatString(t *testing.T) {
	t.Parallel()

	require.Equal(t, monthFormatString, MonthFormatString(), "Incorrect month format string.")

	testCases := []struct {
		name     string
		timezone string
		month    int
		year     int
	}{
		{
			name:     "2023-01-01T00:00:00-04:00",
			timezone: "-04:00",
			month:    1,
			year:     2023,
		}, {
			name:     "2023-01-01T00:00:00+04:00",
			timezone: "+04:00",
			month:    1,
			year:     2023,
		}, {
			name:     "2023-12-01T00:00:00+04:00",
			timezone: "+04:00",
			month:    12,
			year:     2023,
		}, {
			name:     "2023-06-01T00:00:00+00:00",
			timezone: "+00:00",
			month:    06,
			year:     2023,
		}, {
			name:     "9999-99-01T00:00:00+99:99",
			timezone: "+99:99",
			month:    99,
			year:     9999,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := fmt.Sprintf(MonthFormatString(), test.year, test.month, test.timezone)
			require.Equal(t, test.name, actual, "actual and expected time strings mismatch.")
		})
	}
}

func TestNextPageRESTFormatString(t *testing.T) {
	require.Equal(t, nextPageRESTFormatString, NextPageRESTFormatString(),
		"next page format strings mismatched.")
}

func TestGraphqlConfigFileName(t *testing.T) {
	require.Equal(t, graphqlConfigFileName, HTTPGraphQLFileName(), "Incorrect HTTP GraphQL filename.")
}

func TestHTTPGraphQLPrefix(t *testing.T) {
	require.Equal(t, graphQLPrefix, HTTPGraphQLPrefix(), "Incorrect HTTP GraphQL environment prefix.")
}

func TestSpecialAccountFiat(t *testing.T) {
	require.Equal(t, specialAccountFiat, SpecialAccountFiat(), "Incorrect Fiat currency account name.")
}

func TestSpecialAccountCrypto(t *testing.T) {
	require.Equal(t, specialAccountCrypto, SpecialAccountCrypto(), "Incorrect Cryptocurrency account name.")
}

func TestInvalidRequest(t *testing.T) {
	require.Equal(t, invalidRequestString, InvalidRequestString(), "Incorrect invalid request string.")
}

func TestValidationString(t *testing.T) {
	require.Equal(t, validationSting, ValidationString(), "Incorrect validation string.")
}

func TestInvalidCurrencyString(t *testing.T) {
	require.Equal(t, invalidCurrencyString, InvalidCurrencyString(), "Incorrect invalid currency string.")
}

func TestRetryMessageString(t *testing.T) {
	require.Equal(t, retryMessageString, RetryMessageString(), "Incorrect retry message string.")
}

func TestTwoSeconds(t *testing.T) {
	require.Equal(t, twoSecondDuration, TwoSeconds(), "Incorrect two second duration.")
}
func TestThreeSeconds(t *testing.T) {
	require.Equal(t, threeSecondDuration, ThreeSeconds(), "Incorrect three second duration.")
}

func TestClientIDCtxKey(t *testing.T) {
	require.Equal(t, clientIDCtxKey, ClientIDCtxKey(), "Incorrect clientID context key.")
}

func TestExpiresAtCtxKey(t *testing.T) {
	require.Equal(t, expiresAtCtxKey, ExpiresAtCtxKey(), "Incorrect expiration deadline context key.")
}

func TestErrorFormatMessage(t *testing.T) {
	require.Equal(t, errorFormatMessage, ErrorFormatMessage(), "Incorrect error format string.")
}
