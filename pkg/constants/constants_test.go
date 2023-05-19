package constants

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetEtcDir(t *testing.T) {
	require.Equal(t, configEtcDir, GetEtcDir(), "Incorrect etc directory")
}

func TestGetHomeDir(t *testing.T) {
	require.Equal(t, configHomeDir, GetHomeDir(), "Incorrect home directory")
}

func TestGetBaseDir(t *testing.T) {
	require.Equal(t, configBaseDir, GetBaseDir(), "Incorrect base directory")
}

func TestGetGithubCIKey(t *testing.T) {
	require.Equal(t, githubCIKey, GetGithubCIKey(), "Incorrect Github CI environment key")
}

func TestGetLoggerFileName(t *testing.T) {
	require.Equal(t, loggerConfigFileName, GetLoggerFileName(), "Incorrect logger filename")
}

func TestGetLoggerPrefix(t *testing.T) {
	require.Equal(t, loggerPrefix, GetLoggerPrefix(), "Incorrect Zap logger environment prefix")
}

func TestGetPostgresFileName(t *testing.T) {
	require.Equal(t, postgresConfigFileName, GetPostgresFileName(), "Incorrect Postgres filename")
}

func TestGetPostgresPrefix(t *testing.T) {
	require.Equal(t, postgresPrefix, GetPostgresPrefix(), "Incorrect Postgres environment prefix")
}

func TestGetPostgresDSN(t *testing.T) {
	require.Equal(t, postgresDSN, GetPostgresDSN(), "Incorrect Postgres DSN format string")
}

func TestGetTestDatabaseName(t *testing.T) {
	require.Equal(t, testDatabaseName, GetTestDatabaseName(), "Incorrect test suite database name")
}

func TestGetRedisFileName(t *testing.T) {
	require.Equal(t, redisConfigFileName, GetRedisFileName(), "Incorrect Redis filename")
}

func TestGetRedisPrefix(t *testing.T) {
	require.Equal(t, redisPrefix, GetRedisPrefix(), "Incorrect Redis prefix")
}

func TestGetQuotesFileName(t *testing.T) {
	require.Equal(t, quotesConfigFileName, GetQuotesFileName(), "Incorrect Quotes filename")
}

func TestGetQuotesPrefix(t *testing.T) {
	require.Equal(t, quotesPrefix, GetQuotesPrefix(), "Incorrect Quotes prefix")
}

func TestGetAuthFileName(t *testing.T) {
	require.Equal(t, authConfigFileName, GetAuthFileName(), "Incorrect authentication filename")
}

func TestGetAuthPrefix(t *testing.T) {
	require.Equal(t, authPrefix, GetAuthPrefix(), "Incorrect authorization environment prefix")
}

func TestGetHTTPRESTFileName(t *testing.T) {
	require.Equal(t, restConfigFileName, GetHTTPRESTFileName(), "Incorrect HTTP REST filename")
}

func TestGetHTTPRESTPrefix(t *testing.T) {
	require.Equal(t, restPrefix, GetHTTPRESTPrefix(), "Incorrect HTTP REST environment prefix")
}

func TestGetDeleteUserAccountConfirmation(t *testing.T) {
	require.Equal(t, deleteUserAccountConfirmation, GetDeleteUserAccountConfirmation(),
		"Incorrect user account deletion confirmation message.")
}

func TestGetDecimalPlacesFiat(t *testing.T) {
	require.Equal(t, fiatDecimalPlaces, GetDecimalPlacesFiat(), "Incorrect Fiat currency decimal places.")
}
func TestGetDecimalPlacesCrypto(t *testing.T) {
	require.Equal(t, cryptoDecimalPlaces, GetDecimalPlacesCrypto(), "Incorrect Cryptocurrency decimal places.")
}

func TestGetFiatOfferTTL(t *testing.T) {
	require.Equal(t, fiatOfferTTL, GetFiatOfferTTL(), "Incorrect Fiat offer TTL.")
}

func TestGetCryptoOfferTTL(t *testing.T) {
	require.Equal(t, cryptoOfferTTL, GetCryptoOfferTTL(), "Incorrect Crypto offer TTL.")
}

func TestGetMonthFormatString(t *testing.T) {
	t.Parallel()

	require.Equal(t, monthFormatString, GetMonthFormatString(), "Incorrect month format string.")

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

			actual := fmt.Sprintf(GetMonthFormatString(), test.year, test.month, test.timezone)
			require.Equal(t, test.name, actual, "actual and expected time strings mismatch.")
		})
	}
}

func TestGetNextPageRESTFormatString(t *testing.T) {
	require.Equal(t, nextPageRESTFormatString, GetNextPageRESTFormatString(),
		"next page format strings mismatched.")
}

func TestGetGraphqlConfigFileName(t *testing.T) {
	require.Equal(t, graphqlConfigFileName, GetHTTPGraphQLFileName(), "Incorrect HTTP GraphQL filename.")
}

func TestGetHTTPGraphQLPrefix(t *testing.T) {
	require.Equal(t, graphQLPrefix, GetHTTPGraphQLPrefix(), "Incorrect HTTP GraphQL environment prefix.")
}

func TestGetSpecialAccountFiat(t *testing.T) {
	require.Equal(t, specialAccountFiat, GetSpecialAccountFiat(), "Incorrect Fiat currency account name.")
}

func TestGetSpecialAccountCrypto(t *testing.T) {
	require.Equal(t, specialAccountCrypto, GetSpecialAccountCrypto(), "Incorrect Cryptocurrency account name.")
}
