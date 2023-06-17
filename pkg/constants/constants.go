package constants

import "time"

const (
	// Configuration file directories.
	configEtcDir  = "/etc/FTeX.conf/"
	configHomeDir = "$HOME/.FTeX/"
	configBaseDir = "./configs/"

	// Configuration file names.
	loggerConfigFileName   = "LoggerConfig.yaml"
	postgresConfigFileName = "PostgresConfig.yaml"
	redisConfigFileName    = "RedisConfig.yaml"
	quotesConfigFileName   = "QuotesConfig.yaml"
	authConfigFileName     = "AuthConfig.yaml"
	restConfigFileName     = "HTTPRESTConfig.yaml"
	graphqlConfigFileName  = "HTTPGraphQLConfig.yaml"

	// Environment variables.
	githubCIKey    = "GITHUB_ACTIONS_CI"
	loggerPrefix   = "LOGGER"
	postgresPrefix = "POSTGRES"
	redisPrefix    = "REDIS"
	quotesPrefix   = "QUOTES"
	authPrefix     = "AUTH"
	restPrefix     = "REST"
	graphQLPrefix  = "GRAPHQL"

	// Miscellaneous.
	postgresDSN                   = "user=%s password=%s host=%s port=%d dbname=%s connect_timeout=%d sslmode=disable"
	testDatabaseName              = "ftex_db_test"
	deleteUserAccountConfirmation = "I understand the consequences, delete my user account %s"
	fiatDecimalPlaces             = int32(2)
	cryptoDecimalPlaces           = int32(8)
	fiatOfferTTL                  = 2 * time.Minute
	cryptoOfferTTL                = 2 * time.Minute
	monthFormatString             = "%d-%02d-01T00:00:00%s" // YYYY-MM-DDTHH:MM:SS+HH:MM (last section is +/- timezone.)
	nextPageRESTFormatString      = "?pageCursor=%s&pageSize=%d"
	specialAccountFiat            = "fiat-currencies"
	specialAccountCrypto          = "crypto-currencies"
	invalidRequestString          = "invalid request"
	validationSting               = "validation"
	invalidCurrencyString         = "invalid currency"
	retryMessageString            = "please retry your request later"
)

// EtcDir returns the configuration directory in Etc.
func EtcDir() string {
	return configEtcDir
}

// HomeDir returns the configuration directory in users home.
func HomeDir() string {
	return configHomeDir
}

// BaseDir returns the configuration base directory in the root of the application.
func BaseDir() string {
	return configBaseDir
}

// GithubCIKey is the key for the environment variable expected to be present in the GH CI runner.
func GithubCIKey() string {
	return githubCIKey
}

// LoggerFileName returns the Zap logger configuration file name.
func LoggerFileName() string {
	return loggerConfigFileName
}

// LoggerPrefix returns the environment variable prefix for the Zap logger.
func LoggerPrefix() string {
	return loggerPrefix
}

// PostgresFileName returns the Postgres configuration file name.
func PostgresFileName() string {
	return postgresConfigFileName
}

// PostgresPrefix returns the environment variable prefix for Postgres.
func PostgresPrefix() string {
	return postgresPrefix
}

// PostgresDSN returns the format string for the Postgres Data Source Name used to connect to the database.
func PostgresDSN() string {
	return postgresDSN
}

// TestDatabaseName returns the name of the database used in test suites.
func TestDatabaseName() string {
	return testDatabaseName
}

// RedisFileName returns the Redis server configuration file name.
func RedisFileName() string {
	return redisConfigFileName
}

// RedisPrefix returns the environment variable prefix for the Redis server.
func RedisPrefix() string {
	return redisPrefix
}

// QuotesFileName returns the quotes configuration file name.
func QuotesFileName() string {
	return quotesConfigFileName
}

// QuotesPrefix returns the environment variable prefix for the quotes.
func QuotesPrefix() string {
	return quotesPrefix
}

// AuthFileName returns the authentication configuration file name.
func AuthFileName() string {
	return authConfigFileName
}

// AuthPrefix returns the environment variable prefix for authentication.
func AuthPrefix() string {
	return authPrefix
}

// HTTPRESTFileName returns the HTTP REST endpoint configuration file name.
func HTTPRESTFileName() string {
	return restConfigFileName
}

// HTTPRESTPrefix returns the environment variable prefix for the HTTP REST endpoint.
func HTTPRESTPrefix() string {
	return restPrefix
}

// DeleteUserAccountConfirmation is the format string template confirmation message used to delete a user account.
func DeleteUserAccountConfirmation() string {
	return deleteUserAccountConfirmation
}

// DecimalPlacesFiat the number of decimal places Fiat currency can have.
func DecimalPlacesFiat() int32 {
	return fiatDecimalPlaces
}

// DecimalPlacesCrypto the number of decimal places Cryptocurrency can have.
func DecimalPlacesCrypto() int32 {
	return cryptoDecimalPlaces
}

// FiatOfferTTL is the time duration that a Fiat conversion rate offer will be valid for.
func FiatOfferTTL() time.Duration {
	return fiatOfferTTL
}

// CryptoOfferTTL is the time duration that a Crypto conversion rate offer will be valid for.
func CryptoOfferTTL() time.Duration {
	return cryptoOfferTTL
}

// MonthFormatString is the base RFC3339 format string for a configurable month, year, and timezone.
func MonthFormatString() string {
	return monthFormatString
}

// NextPageRESTFormatString is the format for the naked next page link for REST requests responses.
func NextPageRESTFormatString() string {
	return nextPageRESTFormatString
}

// HTTPGraphQLFileName returns the HTTP GraphQL endpoint configuration file name.
func HTTPGraphQLFileName() string {
	return graphqlConfigFileName
}

// HTTPGraphQLPrefix returns the environment variable prefix for the HTTP GraphQL endpoint.
func HTTPGraphQLPrefix() string {
	return graphQLPrefix
}

// SpecialAccountFiat special purpose account for Fiat currency related operations in the database.
func SpecialAccountFiat() string {
	return specialAccountFiat
}

// SpecialAccountCrypto special purpose account for Cryptocurrency related operations in the database.
func SpecialAccountCrypto() string {
	return specialAccountCrypto
}

// InvalidRequestString is the error string message for an invalid request.
func InvalidRequestString() string {
	return invalidRequestString
}

// ValidationString is the error message for a struct validation failure.
func ValidationString() string {
	return validationSting
}

// InvalidCurrencyString is the error message for an invalid currency.
func InvalidCurrencyString() string {
	return invalidCurrencyString
}

// RetryMessageString is the error message requesting a retry.
func RetryMessageString() string {
	return retryMessageString
}
