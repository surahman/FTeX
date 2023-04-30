package redis

// configTestData will return a map of test data containing valid and invalid Redis configs.
func configTestData() map[string]string {
	return map[string]string{
		"empty": ``,

		"valid": `
authentication:
  username: root
  password: root
connection:
  addr: 127.0.0.1:7379
  maxConnAttempts: 5
  maxRetries: 3
  poolSize: 4
  minIdleConns: 1
  maxIdleConns: 20`,

		"username_empty": `
authentication:
  username:
  password: root
connection:
  addr: 127.0.0.1:7379
  maxConnAttempts: 5
  maxRetries: 3
  poolSize: 4
  minIdleConns: 1
  maxIdleConns: 20`,

		"password_empty": `
authentication:
  username: root
  password:
connection:
  addr: 127.0.0.1:7379
  maxConnAttempts: 5
  maxRetries: 3
  poolSize: 4
  minIdleConns: 1
  maxIdleConns: 20`,

		"no_addr": `
authentication:
  username: root
  password: root
connection:
  addr:
  maxConnAttempts: 5
  maxRetries: 3
  poolSize: 4
  minIdleConns: 1
  maxIdleConns: 20`,

		"invalid_max_retries": `
authentication:
  username: root
  password: root
connection:
  addr: 127.0.0.1:7379
  maxConnAttempts: 5
  maxRetries: 0
  poolSize: 4
  minIdleConns: 1
  maxIdleConns: 20`,

		"invalid_pool_size": `
authentication:
  username: root
  password: root
connection:
  addr: 127.0.0.1:7379
  maxConnAttempts: 5
  maxRetries: 3
  poolSize: 0
  minIdleConns: 1
  maxIdleConns: 20`,

		"invalid_min_idle_conns": `
authentication:
  username: root
  password: root
connection:
  addr: 127.0.0.1:7379
  maxConnAttempts: 5
  maxRetries: 3
  poolSize: 4
  minIdleConns: 0
  maxIdleConns: 20`,

		"no_max_idle_conns": `
authentication:
  username: root
  password: root
connection:
  addr: 127.0.0.1:7379
  maxConnAttempts: 5
  maxRetries: 3
  poolSize: 4
  minIdleConns: 10
  maxIdleConns:`,

		"test_suite": `
authentication:
  username: ftex_service
  password: ZoF1bncLLyYT1agKfWQY
connection:
  addr: 127.0.0.1:7379
  maxConnAttempts: 1
  maxRetries: 3
  poolSize: 4
  minIdleConns: 1
  maxIdleConns: 20`,

		"github-ci-runner": `
authentication:
  username: default
  password:
connection:
  addr: 127.0.0.1:6379
  maxConnAttempts: 3
  maxRetries: 3
  poolSize: 4
  minIdleConns: 1
  maxIdleConns: 20`,
	}
}
