# HTTP REST API Endpoints

The REST API schema can be tested and reviewed through the Swagger UI that is exposed when the server is started.

<br/>

## Table of contents

- [Authorization Response](#authorization-response)
- [Error Response](#error-response)
- [Success Response](#success-response)
- [Healthcheck Endpoint `/health`](#healthcheck-endpoint-health)
- [User Endpoints `/user/`](#user-endpoints-user)
  - [Register](#register)
  - [Login](#login)
  - [Refresh](#refresh)

<br/>

### Authorization Response

Authorization is implemented using JSON Web Tokens. An expiration deadline for the JWT is returned in response. It is
the client's responsibility to refresh the token before, but no sooner than 60 seconds, before the deadline.

The returned token schema is below.

```json
{
  "expires": "expiration time integer in seconds, Unix time stamp",
  "token": "token string",
  "threshold": "threshold in integer seconds before expiration when the token can be refreshed"
}
```

<br/>

### Error Response

There is a generic error response with a message and optional payload. If there is a validation error of some sort the
details of the failures will be enclosed within the payload section of the response.

```json
{
  "message": "message string",
  "payload": "string or JSON object"
}
```

<br/>

### Success Response

A successful request _may_ result in a response object when appropriate. In such an event, a message and an optional
payload will be returned.

```json
{
  "message": "message string",
  "payload": "string or JSON object"
}
```

<br/>

### Healthcheck Endpoint `/health`

The health check endpoint is exposed to facilitate liveness checks on the service. The check will verify whether the
service is connected to all the ancillary services and responds appropriately.

This check is essential for load balancers and container orchestrators to determine whether to route traffic or restart
the container.

_Healthy Response:_ HTTP 200 OK

_Unhealthy Response:_ HTTP 503 Service Unavailable


<br/>

### User Endpoints `/user/`

#### Register

_Request:_ All fields are required.
_Response:_ A valid JWT will be returned as an authorization response.

```json
{
  "email": "string",
  "first_name": "string",
  "last_name": "string",
  "password": "string",
  "username": "string"
}
```

#### Login

_Request:_ All fields are required.
_Response:_ A valid JWT will be returned as an authorization response.

```json
{
  "password": "string",
  "username": "string"
}
```
#### Refresh

_Request:_ A valid JWT must be provided in the request header and will be validated with a fresh token issued against it.
_Response:_ A valid JWT will be returned as an authorization response.

```json
{
  "expires": "expiration time string",
  "token": "token string"
}
```
