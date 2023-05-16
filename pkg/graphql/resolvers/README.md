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

#### Info

##### Balance for a Specific Currency

_Request:_ A valid currency code must be provided as a parameter.
```graphql
mutation {
	balanceFiat(currencyCode:"USD") {
    currency,
    balance,
    lastTx,
    lastTxTs,
    createdAt,
    clientID
  }
}
```

_Response:_ Account balance related details associated with the currency.
```json
{
  "data": {
    "balanceFiat": {
      "currency": "USD",
      "balance": 13569.36,
      "lastTx": -100.11,
      "lastTxTs": "2023-05-15 14:59:24.243332 -0400 EDT",
      "createdAt": "2023-05-09 18:29:04.345387 -0400 EDT",
      "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe"
    }
  }
}
```

##### Balance for all Currencies for a Client

_Request:_ The initial request can only contain an optional `page size`, which if not provided will default to 10. The
subsequent responses will contain encrypted page cursors that must be specified to retrieve the following page of data.

Initial request: The `pageCursor` will not be provided and the `pageSize` is optional and will default to 10.
```graphql
mutation {
  balanceAllFiat(pageSize: 3) {
    accountBalances{
      currency
      balance
      lastTx
      lastTxTs
      createdAt
      clientID
    }
    links{
      pageCursor
    }
  }
}
```

Subsequent requests: The `pageCursor` must be provided but the `pageSize` is optional.
```graphql
mutation {
  balanceAllFiat(pageCursor: "G4dGbYhcNY8ByNNpdgYJq-jK1eRXHD7lBp56-IeiAQ==", pageSize: 3) {
    accountBalances{
      currency
      balance
      lastTx
      lastTxTs
      createdAt
      clientID
    }
    links{
      pageCursor
    }
  }
}
```


_Response:_ The number of account balances for the Client will be limited to the `Page Size` specified and is `10` by
default. A `Page Cursor` link will be supplied if there are subsequent pages of data to be retrieved in the
`links.pageCursor` JSON field.

```json
{
  "data": {
    "balanceAllFiat": {
      "accountBalances": [
        {
          "currency": "AED",
          "balance": 30903.7,
          "lastTx": -10000,
          "lastTxTs": "2023-05-09 18:33:55.453689 -0400 EDT",
          "createdAt": "2023-05-09 18:29:16.74704 -0400 EDT",
          "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe"
        },
        {
          "currency": "CAD",
          "balance": 369283.5,
          "lastTx": 134.75,
          "lastTxTs": "2023-05-15 16:59:24.243332 -0400 EDT",
          "createdAt": "2023-05-09 18:29:08.746285 -0400 EDT",
          "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe"
        },
        {
          "currency": "EUR",
          "balance": 1536.45,
          "lastTx": 1536.45,
          "lastTxTs": "2023-05-09 18:31:32.213239 -0400 EDT",
          "createdAt": "2023-05-09 18:29:21.365991 -0400 EDT",
          "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe"
        }
      ],
      "links": {
        "pageCursor": "iaguqIObr8FvtimV4k1uHJtZ2DHGPgTxNZVmsyEKKA=="
      }
    }
  }
}
```

```json
{
  "data": {
    "balanceAllFiat": {
      "accountBalances": [
        {
          "currency": "USD",
          "balance": 13569.36,
          "lastTx": -100.11,
          "lastTxTs": "2023-05-15 16:59:24.243332 -0400 EDT",
          "createdAt": "2023-05-09 18:29:04.345387 -0400 EDT",
          "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe"
        }
      ],
      "links": {
        "pageCursor": ""
      }
    }
  }
}
```
##### Transaction Details for a Specific Transaction

_Request:_ A valid `Transaction ID` must be provided as a path parameter.

```graphql

```

_Response:_ Transaction-related details for a specific transaction. In the event of an external deposit, there will be
a single entry reporting the deposited amount. When querying for an internal transfer, two entries will be returned -
one for the source and the other for the destination accounts.

###### External Transfer (deposit)
```json
{
  "message": "transaction details",
  "payload": [
    {
      "currency": "USD",
      "amount": "10101.11",
      "transactedAt": "2023-04-28T17:24:53.396603-04:00",
      "clientID": "a8d55c17-09cc-4805-a7f7-4c5038a97b32",
      "txID": "de7456cb-1dde-4b73-941d-252a1fb1d337"
    }
  ]
}
```

###### Internal Transfer (currency conversion/exchange)
```json
{
  "message": "transaction details",
  "payload": [
    {
      "currency": "CAD",
      "amount": "-100.26",
      "transactedAt": "2023-04-30T17:06:54.654345-04:00",
      "clientID": "a8d55c17-09cc-4805-a7f7-4c5038a97b32",
      "txID": "da3f100a-2f47-4879-a3b7-bb0517c3b1ac"
    },
    {
      "currency": "USD",
      "amount": "73.44",
      "transactedAt": "2023-04-30T17:06:54.654345-04:00",
      "clientID": "a8d55c17-09cc-4805-a7f7-4c5038a97b32",
      "txID": "da3f100a-2f47-4879-a3b7-bb0517c3b1ac"
    }
  ]
}
```
##### Transaction Details for a Specific Currency

_Request:_ A valid `Currency Code` must be provided as a path parameter. The path parameters accepted are listed below.
If a `pageCursor` is supplied, all other parameters except for the `pageSize` are ignored.

Optional:
* `pageCursor`: Defaults to 10.

Initial Page (required):
* `month`: Month for which the transactions are being requested.
* `year`: Year for which the transactions are being requested.
* `timezone`: Timezone for which the transactions are being requested.

Subsequent Pages (required)
* `pageCursor`: Hashed page cursor for the next page of data.

```graphql

```

_Response:_ All Transaction-related details for a specific currency in a given timezone and date are returned. In the
event of an external deposit, there will be a single entry reporting the deposited amount. When querying for an internal
transfer, two entries will be returned - one for the source and the other for the destination accounts.

###### Initial Page
```json
{
  "message": "account transactions",
  "payload": {
    "transactionDetails": [
      {
        "currency": "AED",
        "amount": "10000",
        "transactedAt": "2023-05-09T18:33:55.453689-04:00",
        "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
        "txID": "af4467a9-7c0a-4437-acf3-e5060509a5d9"
      },
      {
        "currency": "AED",
        "amount": "8180.74",
        "transactedAt": "2023-05-09T18:32:16.38917-04:00",
        "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
        "txID": "b6a760ba-a189-4222-9897-4a783c799953"
      },
      {
        "currency": "AED",
        "amount": "4396.12",
        "transactedAt": "2023-05-09T18:32:16.004549-04:00",
        "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
        "txID": "7108d3e5-257e-45a8-ace1-d7e86c84556e"
      }
    ],
    "links": {
      "nextPage": "?pageCursor=xft0C3AaJwShw6Du5tr0d8FKXYedyFd1cgPp13W2LvU9U8ii3svtRn2Tt7Pd3LI6nQvO3AUI0NioM18v6XGFXuC4jpFDA8AsqFnXqSZMwMSk&pageSize=3"
    }
  }
}
```

###### Subsequent Page
```json
{
  "message": "account transactions",
  "payload": {
    "transactionDetails": [
      {
        "currency": "AED",
        "amount": "4561.01",
        "transactedAt": "2023-05-09T18:32:15.547456-04:00",
        "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
        "txID": "525ea850-916b-4761-ae28-a34a63613212"
      },
      {
        "currency": "AED",
        "amount": "3323.22",
        "transactedAt": "2023-05-09T18:32:15.137486-04:00",
        "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
        "txID": "77278e19-5a1b-46fe-a106-d2f21ad72839"
      },
      {
        "currency": "AED",
        "amount": "4242.43",
        "transactedAt": "2023-05-09T18:31:49.872366-04:00",
        "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
        "txID": "6c930c8c-fef8-4711-8961-2d101bfb7a5e"
      }
    ],
    "links": {
      "nextPage": ""
    }
  }
}
```
