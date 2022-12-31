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
  max_connection_attempts: 1
  port: 6432
  timeout: 5
pool:
  health_check_period: 30s
  max_conns: 4
  min_conns: 4`,

		"github-ci-runner": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db_test
  max_connection_attempts: 3
  host: 127.0.0.1
  port: 5432
  timeout: 5
pool:
  health_check_period: 30s
  max_conns: 8
  min_conns: 4`,

		"valid": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db
  host: 127.0.0.1
  max_connection_attempts: 5
  port: 6432
  timeout: 5
pool:
  health_check_period: 30s
  max_conns: 8
  min_conns: 4`,

		"bad_health_check": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db
  host: 127.0.0.1
  max_connection_attempts: 5
  port: 6432
  timeout: 5
pool:
  health_check_period: 3s
  max_conns: 8
  min_conns: 4`,

		"invalid_conns": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db
  host: 127.0.0.1
  max_connection_attempts: 5
  port: 6432
  timeout: 5
pool:
  health_check_period: 30s
  max_conns: 2
  min_conns: 2`,

		"invalid_max_conn_attempts": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db
  host: 127.0.0.1
  max_connection_attempts: 0
  port: 6432
  timeout: 5
pool:
  health_check_period: 30s
  max_conns: 8
  min_conns: 4`,

		"invalid_timeout": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db
  host: 127.0.0.1
  max_connection_attempts: 5
  port: 6432
  timeout: 2
pool:
  health_check_period: 30s
  max_conns: 8
  min_conns: 4`,
	}
}
