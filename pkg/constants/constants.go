package constants

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

	// Environment variables.
	githubCIKey    = "GITHUB_ACTIONS_CI"
	loggerPrefix   = "LOGGER"
	postgresPrefix = "POSTGRES"
	redisPrefix    = "REDIS"
	quotesPrefix   = "QUOTES"

	// Miscellaneous.
	postgresDSN      = "user=%s password=%s host=%s port=%d dbname=%s connect_timeout=%d sslmode=disable"
	testDatabaseName = "ftex_db_test"
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
