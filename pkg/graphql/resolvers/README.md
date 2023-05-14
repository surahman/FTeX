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
  healthcheck
}
```

_Healthy Response:_

```json
{
  "data": {
    "healthcheck": "OK"
  }
}
```

_Unhealthy Response:_

```json
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

<br/>

### User Mutations

#### Register

_Request:_ All fields are required.

```graphql
mutation {
  registerUser(input: {
    firstname: "first name"
    lastname: "last name"
    email: "email@address.com",
    userLoginCredentials: {
      username:"someusername",
      password: "somepassword"
    }
  }) {
  token,
  expires,
  threshold
  }
}
```

_Response:_ A valid JWT will be returned as an authorization response.


#### Login

_Request:_ All fields are required.

```graphql
mutation {
  loginUser(input: {
    username:"someusername",
    password: "somepassword"
  }) {
    token,
    expires,
    threshold
  }
}
```

_Response:_ A valid JWT will be returned as an authorization response.


#### Refresh

_Request:_ A valid JWT must be provided in the request header and will be validated with a fresh token issued against it.

```graphql
mutation {
  refreshToken {
    token
    expires
    threshold
  }
}
```

_Response:_ A valid JWT will be returned as an authorization response.


#### Delete

_Request:_ All fields are required and a valid JWT must be provided in the header. The user must supply their login
credentials as well as complete the confirmation message `I understand the consequences, delete my user
account **USERNAME HERE**`

```graphql
mutation {
  deleteUser(input: {
    username: "someusername"
    password: "somepassword"
    confirmation: "I understand the consequences, delete my user account <USERNAME HERE>"
  })
}
```

_Response:_ A confirmation message will be returned as a success response.


<br/>


### Fiat Account Mutations

#### Open Account

_Request:_ All fields are required.

```graphql
mutation {
    openFiat(currency: "USD") {
        clientID,
        currency
    }
}
```

_Response:_ Confirmation information containing the `Client ID` and `Currency` of the newly opened account.

```json
{
  "data": {
    "openFiat": {
      "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
      "currency": "USD"
    }
  }
}
```

#### Deposit

Deposit money into a Fiat account for a specific currency and amount. An account for the currency must already be opened
for the deposit to succeed.

_Request:_ All fields are required.
```graphql
mutation {
    depositFiat(input: {
        amount:1345.67,
        currency: "USD"
    }) {
        txId,
        clientId,
        txTimestamp,
        balance,
        lastTx,
        currency
    }
}
```

_Response:_ A confirmation of the transaction with the particulars of the transfer.
```json
{
  "data": {
    "depositFiat": {
      "txId": "8522591d-6463-4cc6-9e3c-c456c98a6755",
      "clientId": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
      "txTimestamp": "2023-05-14 11:57:47.796057 -0400 EDT",
      "balance": "14170.02",
      "lastTx": "1345.67",
      "currency": "USD"
    }
  }
}
```
