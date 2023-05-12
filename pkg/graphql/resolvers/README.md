# HTTP GraphQL API Endpoints

The GraphQL API schema can be tested and reviewed through the GraphQL Playground that is exposed when the server is started.

<br/>

## Table of contents


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

<br/>

### Authorization

A valid JSON Web Token must be included in the header in the HTTP request for all endpoints that require authorization.
The `Authorization` header key is customizable through the GraphQL endpoint configurations.

The following `queries` and `mutations` do not require authorization:
- Register User: `registerUser`
- Login User: `loginUser`
- Healthcheck: `healthcheck`

```json
{
  "Authorization":
  "JSON Web Token goes here"
}
```

<br/>

### Healthcheck Query.

The health check endpoint is exposed to facilitate liveness checks on the service. The check will verify whether the
service is connected to all the ancillary services and responds appropriately.

This check is essential for load balancers and container orchestrators to determine whether to route traffic or restart
the container.

```graphql
query {
  healthcheck()
}
```

_Healthy Response:_

```graphql
{
  "data": {
    "healthcheck": "OK"
  }
}
```

_Unhealthy Response:_

```graphql
{
  "errors": [
    {
      "message": "[Postgres|Redis] healthcheck failed",
      "path": [
        "healthcheck"
      ]
    }
  ],
  "data": null
}
```
