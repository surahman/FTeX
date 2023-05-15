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

#### Exchange

To convert between Fiat currencies, the user must maintain open accounts in both the source and destination Fiat currencies.
The amount specified will be in the source currency and the amount to deposit into the destination account will be calculated
based on the exchange rate.

The workflow will involve getting a conversion rate quote, referred to as an `Offer`. The returned rate quote `Offer` will
only be valid for a two-minute time window. The expiration time will be returned to the user as a Unix timestamp. The user
must issue a subsequent request using the encrypted `Offer ID` to complete the transaction.

##### Quote

_Request:_ All fields are required.
```graphql
mutation {
    exchangeOfferFiat(input: {
        sourceCurrency:"USD"
        destinationCurrency: "CAD"
        sourceAmount: 100.11
    }) {
        priceQuote{
            clientID,
            sourceAcc,
            destinationAcc,
            rate,
            amount
        },
        debitAmount,
        offerID,
        expires
    }
}
```

_Response:_ A rate quote with an encrypted `Offer ID`.
```json
{
  "data": {
    "exchangeOfferFiat": {
      "priceQuote": {
        "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
        "sourceAcc": "USD",
        "destinationAcc": "CAD",
        "rate": 1.355365,
        "amount": 135.69
      },
      "debitAmount": 100.11,
      "offerID": "ME0pUhmOJRescxQx7IhJYrgIxeSJ-P4dABP2QVFbr5FGlu-yI_4GoGJ0oW23KTGf",
      "expires": 1684116836
    }
  }
}
```

##### Convert

_Request:_ All fields are required.
```grpahql
mutation {
	exchangeTransferFiat(offerID: "-ptOjSHs3cw3eTw_1NuInn4w8OvI8hzFzChol7NRpKIHMDL234B_E1Fcq5Z6Zl4K") {
    sourceReceipt {
    	txId,
    	clientId,
    	txTimestamp,
    	balance,
    	lastTx,
    	currency
    },
    destinationReceipt {
    	txId,
    	clientId,
    	txTimestamp,
    	balance,
    	lastTx,
    	currency
    }
  }
}
```

_Response:_ A transaction receipt with the details of the source and destination accounts and transaction details.
```json
{
  "data": {
    "exchangeTransferFiat": {
      "sourceReceipt": {
        "txId": "043d82a9-113b-4aa7-a3e1-029cc4728926",
        "clientId": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
        "txTimestamp": "2023-05-15 16:59:24.243332 -0400 EDT",
        "balance": "13569.36",
        "lastTx": "-100.11",
        "currency": "USD"
      },
      "destinationReceipt": {
        "txId": "043d82a9-113b-4aa7-a3e1-029cc4728926",
        "clientId": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
        "txTimestamp": "2023-05-15 16:59:24.243332 -0400 EDT",
        "balance": "369283.5",
        "lastTx": "134.75",
        "currency": "CAD"
      }
    }
  }
}
```
