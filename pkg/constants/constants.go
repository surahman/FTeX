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

	// Environment variables.
	githubCIKey    = "GITHUB_ACTIONS_CI"
	loggerPrefix   = "LOGGER"
	postgresPrefix = "POSTGRES"
	redisPrefix    = "REDIS"
	quotesPrefix   = "QUOTES"
	authPrefix     = "AUTH"
	restPrefix     = "REST"

	// Miscellaneous.
	postgresDSN                   = "user=%s password=%s host=%s port=%d dbname=%s connect_timeout=%d sslmode=disable"
	testDatabaseName              = "ftex_db_test"
	deleteUserAccountConfirmation = "I understand the consequences, delete my user account %s"
	fiatDecimalPlaces             = int32(2)
	fiatOfferTTL                  = 2 * time.Minute
	cryptoOfferTTL                = 2 * time.Minute
	monthFormatString             = "%d-%02d-01T00:00:00%s" // YYYY-MM-DDTHH:MM:SS+HH:MM (last section is +/- timezone.)
	nextPageRESTFormatString      = "?pageCursor=%s&pageSize=%d"
)

// GetEtcDir returns the configuration directory in Etc.
func GetEtcDir() string {
	return configEtcDir
}

// GetHomeDir returns the configuration directory in users home.
func GetHomeDir() string {
	return configHomeDir
}

// GetBaseDir returns the configuration base directory in the root of the application.
func GetBaseDir() string {
	return configBaseDir
}

// GetGithubCIKey is the key for the environment variable expected to be present in the GH CI runner.
func GetGithubCIKey() string {
	return githubCIKey
}

// GetLoggerFileName returns the Zap logger configuration file name.
func GetLoggerFileName() string {
	return loggerConfigFileName
}

// GetLoggerPrefix returns the environment variable prefix for the Zap logger.
func GetLoggerPrefix() string {
	return loggerPrefix
}

// GetPostgresFileName returns the Postgres configuration file name.
func GetPostgresFileName() string {
	return postgresConfigFileName
}

// GetPostgresPrefix returns the environment variable prefix for Postgres.
func GetPostgresPrefix() string {
	return postgresPrefix
}

// GetPostgresDSN returns the format string for the Postgres Data Source Name used to connect to the database.
func GetPostgresDSN() string {
	return postgresDSN
}

// GetTestDatabaseName returns the name of the database used in test suites.
func GetTestDatabaseName() string {
	return testDatabaseName
}

// GetRedisFileName returns the Redis server configuration file name.
func GetRedisFileName() string {
	return redisConfigFileName
}

// GetRedisPrefix returns the environment variable prefix for the Redis server.
func GetRedisPrefix() string {
	return redisPrefix
}

// GetQuotesFileName returns the quotes configuration file name.
func GetQuotesFileName() string {
	return quotesConfigFileName
}

// GetQuotesPrefix returns the environment variable prefix for the quotes.
func GetQuotesPrefix() string {
	return quotesPrefix
}

// GetAuthFileName returns the authentication configuration file name.
func GetAuthFileName() string {
	return authConfigFileName
}

// GetAuthPrefix returns the environment variable prefix for authentication.
func GetAuthPrefix() string {
	return authPrefix
}

// GetHTTPRESTFileName returns the HTTP REST endpoint configuration file name.
func GetHTTPRESTFileName() string {
	return restConfigFileName
}

// GetHTTPRESTPrefix returns the environment variable prefix for the HTTP REST endpoint.
func GetHTTPRESTPrefix() string {
	return restPrefix
}

// GetDeleteUserAccountConfirmation is the format string template confirmation message used to delete a user account.
func GetDeleteUserAccountConfirmation() string {
	return deleteUserAccountConfirmation
}

// GetDecimalPlacesFiat the number of decimal places Fiat currency can have.
func GetDecimalPlacesFiat() int32 {
	return fiatDecimalPlaces
}

// GetFiatOfferTTL is the time duration that a Fiat conversion rate offer will be valid for.
func GetFiatOfferTTL() time.Duration {
	return fiatOfferTTL
}

// GetCryptoOfferTTL is the time duration that a Crypto conversion rate offer will be valid for.
func GetCryptoOfferTTL() time.Duration {
	return cryptoOfferTTL
}

// GetMonthFormatString is the base RFC3339 format string for a configurable month, year, and timezone.
func GetMonthFormatString() string {
	return monthFormatString
}

// GetNextPageRESTFormatString is the format for the naked next page link for REST requests responses.
func GetNextPageRESTFormatString() string {
	return nextPageRESTFormatString
}
