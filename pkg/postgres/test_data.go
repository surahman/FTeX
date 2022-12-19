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
  database: ft-ex-db-test
  host: 127.0.0.1
  port: 6432
  timeout: 5
  ssl_enabled: false
pool:
  health_check_period: 30s
  max_conns: 4
  min_conns: 4
  lazy_connect: true`,

		"valid": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ft-ex-db
  host: 127.0.0.1
  port: 6432
  timeout: 5
  ssl_enabled: false
pool:
  health_check_period: 30s
  max_conns: 8
  min_conns: 4
  lazy_connect: false`,

		"valid_true_bool": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ft-ex-db
  host: 127.0.0.1
  port: 6432
  timeout: 5
  ssl_enabled: true
pool:
  health_check_period: 30s
  max_conns: 8
  min_conns: 4
  lazy_connect: true`,

		"valid_prod_no_bool": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ft-ex-db
  host: 127.0.0.1
  port: 6432
  timeout: 5
  ssl_enabled:
pool:
  health_check_period: 30s
  max_conns: 8
  min_conns: 4
  lazy_connect:`,

		"bad_health_check": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ft-ex-db
  host: 127.0.0.1
  port: 6432
  timeout: 5
  ssl_enabled: false
pool:
  health_check_period: 3s
  max_conns: 8
  min_conns: 4
  lazy_connect: false`,

		"invalid_conns": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ft-ex-db
  host: 127.0.0.1
  port: 6432
  timeout: 5
  ssl_enabled: false
pool:
  health_check_period: 30s
  max_conns: 2
  min_conns: 2
  lazy_connect: false`,
	}
}
