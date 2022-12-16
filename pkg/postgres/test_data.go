package postgres

// configTestData will return a map of test data containing valid and invalid logger configs.
func configTestData() map[string]string {
	return map[string]string{
		"empty": ``,

		"valid_prod": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ft-ex-db
  host: 127.0.0.1
  port: 6432
  Timeout: 5
  ssl_enabled: false
pool:
  health_check_period: 30
  max_conns: 8
  min_conns: 4
  lazy_connect: false`,

		"valid_prod_true_bool": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ft-ex-db
  host: 127.0.0.1
  port: 6432
  Timeout: 5
  ssl_enabled: true
pool:
  health_check_period: 30
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
  Timeout: 5
pool:
  health_check_period: 30
  max_conns: 8
  min_conns: 4`,

		"bad_health_check": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ft-ex-db
  host: 127.0.0.1
  port: 6432
  Timeout: 5
  ssl_enabled: false
pool:
  health_check_period: 3
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
  Timeout: 5
  ssl_enabled: false
pool:
  health_check_period: 30
  max_conns: 2
  min_conns: 2
  lazy_connect: false`,
	}
}
