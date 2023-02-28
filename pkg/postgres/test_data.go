package postgres

// configTestData will return a map of test data containing valid and invalid logger configs.
func configTestData() map[string]string {
	return map[string]string{
		"empty": ``,

		"test_suite": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db_test
  host: 127.0.0.1
  maxConnectionAttempts: 1
  port: 6432
  timeout: 5
pool:
  healthCheckPeriod: 30s
  maxConns: 4
  minConns: 4`,

		"github-ci-runner": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db_test
  maxConnectionAttempts: 3
  host: 127.0.0.1
  port: 5432
  timeout: 5
pool:
  healthCheckPeriod: 30s
  maxConns: 8
  minConns: 4`,

		"valid": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db
  host: 127.0.0.1
  maxConnectionAttempts: 5
  port: 6432
  timeout: 5
pool:
  healthCheckPeriod: 30s
  maxConns: 8
  minConns: 4`,

		"bad_health_check": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db
  host: 127.0.0.1
  maxConnectionAttempts: 5
  port: 6432
  timeout: 5
pool:
  healthCheckPeriod: 3s
  maxConns: 8
  minConns: 4`,

		"invalid_conns": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db
  host: 127.0.0.1
  maxConnectionAttempts: 5
  port: 6432
  timeout: 5
pool:
  healthCheckPeriod: 30s
  maxConns: 2
  minConns: 2`,

		"invalid_max_conn_attempts": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db
  host: 127.0.0.1
  maxConnectionAttempts: 0
  port: 6432
  timeout: 5
pool:
  healthCheckPeriod: 30s
  maxConns: 8
  minConns: 4`,

		"invalid_timeout": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db
  host: 127.0.0.1
  maxConnectionAttempts: 5
  port: 6432
  timeout: 2
pool:
  healthCheckPeriod: 30s
  maxConns: 8
  minConns: 4`,
	}
}
