package rest

// configTestData will return a map of test data containing valid and invalid Authorization configs.
func configTestData() map[string]string {
	return map[string]string{
		"empty": ``,

		"valid": `
server:
  portNumber: 44243
  shutdownDelay: 5
  basePath: api/rest/v1
  swaggerPath: /swagger/*any
authorization:
  headerKey: Authorization`,

		"out of range port": `
server:
  portNumber: 99
  shutdownDelay: 5
  basePath: api/rest/v1
  swaggerPath: /swagger/*any
authorization:
  headerKey: Authorization`,

		"out of range shutdown delay": `
server:
  portNumber: 44243
  shutdownDelay: -1
  basePath: api/rest/v1
  swaggerPath: /swagger/*any
authorization:
  headerKey: Authorization`,

		"no base path": `
server:
  portNumber: 44243
  shutdownDelay: 5
  swaggerPath: /swagger/*any
authorization:
  headerKey: Authorization`,

		"no swagger path": `
server:
  portNumber: 44243
  shutdownDelay: 5
  basePath: api/rest/v1
authorization:
  headerKey: Authorization`,

		"no auth header": `
server:
  portNumber: 44243
  shutdownDelay: 5
  basePath: api/rest/v1
  swaggerPath:  /swagger/*any`,
	}
}
