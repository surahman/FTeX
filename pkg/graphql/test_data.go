package graphql

// configTestData will return a map of test data containing valid and invalid GraphQL configs.
func configTestData() map[string]string {
	return map[string]string{
		"empty": ``,

		"valid": `
server:
  portNumber: 33723
  shutdownDelay: 5s
  basePath: api/rest/v1
  playgroundPath: /playground
  queryPath: /query
  readTimeout: 3s
  writeTimeout: 3s
  readHeaderTimeout: 3s
authorization:
  headerKey: Authorization`,

		"out of range port": `
server:
  portNumber: 99
  shutdownDelay: 5s
  basePath: api/rest/v1
  playgroundPath: /playground
  queryPath: /query
  readTimeout: 1s
  writeTimeout: 1s
  readHeaderTimeout: 1s
authorization:
  headerKey: Authorization`,

		"out of range time delay": `
server:
  portNumber: 44243
  shutdownDelay: 0s
  basePath: api/rest/v1
  playgroundPath: /playground
  queryPath: /query
  readTimeout: 0s
  writeTimeout: 0s
  readHeaderTimeout: 0s
authorization:
  headerKey: Authorization`,

		"no base path": `
server:
  portNumber: 44243
  shutdownDelay: 5s
  playgroundPath: /playground
  queryPath: /query
  readTimeout: 1s
  writeTimeout: 1s
  readHeaderTimeout: 1s
authorization:
  headerKey: Authorization`,

		"no playground path": `
server:
  portNumber: 44243
  shutdownDelay: 5s
  basePath: api/rest/v1
  queryPath: /query
  readTimeout: 1s
  writeTimeout: 1s
  readHeaderTimeout: 1s
authorization:
  headerKey: Authorization`,

		"no query path": `
server:
  portNumber: 33723
  shutdownDelay: 5s
  basePath: api/rest/v1
  playgroundPath: /playground
  readTimeout: 3s
  writeTimeout: 3s
  readHeaderTimeout: 3s
authorization:
  headerKey: Authorization`,

		"no read timeout": `
server:
  portNumber: 33723
  shutdownDelay: 5s
  basePath: api/rest/v1
  playgroundPath: /playground
  queryPath: /query
  writeTimeout: 1s
  readHeaderTimeout: 1s
authorization:
  headerKey: Authorization`,

		"no write timeout": `
server:
  portNumber: 33723
  shutdownDelay: 5s
  basePath: api/rest/v1
  playgroundPath: /playground
  queryPath: /query
  readTimeout: 1s
  readHeaderTimeout: 1s
authorization:
  headerKey: Authorization`,

		"no read header timeout": `
server:
  portNumber: 33723
  shutdownDelay: 5s
  basePath: api/rest/v1
  playgroundPath: /playground
  queryPath: /query
  readTimeout: 1s
  writeTimeout: 1s
authorization:
  headerKey: Authorization`,

		"no auth header": `
server:
  portNumber: 44243
  shutdownDelay: 5s
  basePath: api/rest/v1
  playgroundPath: /playground
  queryPath: /query
  readTimeout: 1s
  writeTimeout: 1s
  readHeaderTimeout: 1s
authorization:
  headerKey:`,
	}
}
