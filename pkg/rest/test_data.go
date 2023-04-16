package rest

// configTestData will return a map of test data containing valid and invalid Authorization configs.
func configTestData() map[string]string {
	return map[string]string{
		"empty": ``,

		"valid": `
server:
  portNumber: 33723
  shutdownDelay: 5
  basePath: api/rest/v1
  swaggerPath: /swagger/*any
  readTimeout: 1s
  writeTimeout: 1s
  idleTimeout: 30s
  readHeaderTimeout: 1s
authorization:
  headerKey: Authorization`,

		"out of range port": `
server:
  portNumber: 99
  shutdownDelay: 5
  basePath: api/rest/v1
  swaggerPath: /swagger/*any
  readTimeout: 1s
  writeTimeout: 1s
  idleTimeout: 30s
  readHeaderTimeout: 1s
authorization:
  headerKey: Authorization`,

		"out of range shutdown delay": `
server:
  portNumber: 44243
  shutdownDelay: -1
  basePath: api/rest/v1
  swaggerPath: /swagger/*any
  readTimeout: 1s
  writeTimeout: 1s
  idleTimeout: 30s
  readHeaderTimeout: 1s
authorization:
  headerKey: Authorization`,

		"no base path": `
server:
  portNumber: 44243
  shutdownDelay: 5
  swaggerPath: /swagger/*any
  readTimeout: 1s
  writeTimeout: 1s
  idleTimeout: 30s
  readHeaderTimeout: 1s
authorization:
  headerKey: Authorization`,

		"no swagger path": `
server:
  portNumber: 44243
  shutdownDelay: 5
  basePath: api/rest/v1
  readTimeout: 1s
  writeTimeout: 1s
  idleTimeout: 30s
  readHeaderTimeout: 1s
authorization:
  headerKey: Authorization`,

		"no read timeout": `
server:
  portNumber: 33723
  shutdownDelay: 5
  basePath: api/rest/v1
  swaggerPath: /swagger/*any
  writeTimeout: 1s
  idleTimeout: 30s
  readHeaderTimeout: 1s
authorization:
  headerKey: Authorization`,

		"no write timeout": `
server:
  portNumber: 33723
  shutdownDelay: 5
  basePath: api/rest/v1
  swaggerPath: /swagger/*any
  readTimeout: 1s
  idleTimeout: 30s
  readHeaderTimeout: 1s
authorization:
  headerKey: Authorization`,

		"no idle timeout": `
server:
  portNumber: 33723
  shutdownDelay: 5
  basePath: api/rest/v1
  swaggerPath: /swagger/*any
  readTimeout: 1s
  writeTimeout: 1s
  readHeaderTimeout: 1s
authorization:
  headerKey: Authorization`,

		"no read header timeout": `
server:
  portNumber: 33723
  shutdownDelay: 5
  basePath: api/rest/v1
  swaggerPath: /swagger/*any
  readTimeout: 1s
  writeTimeout: 1s
  idleTimeout: 30s
authorization:
  headerKey: Authorization`,

		"no auth header": `
server:
  portNumber: 44243
  shutdownDelay: 5
  basePath: api/rest/v1
  swaggerPath:  /swagger/*any
  readTimeout: 1s
  writeTimeout: 1s
  idleTimeout: 30s
  readHeaderTimeout: 1s
authorization:
  headerKey:`,
	}
}
