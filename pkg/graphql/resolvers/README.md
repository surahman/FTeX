# HTTP GraphQL API Endpoints

The GraphQL API schema can be tested and reviewed through the GraphQL Playground that is exposed when the server is started.

<br/>

## Table of contents

- [Authorization Response](#authorization-response)
- [Authorization](#authorization)
- [Healthcheck Query](#healthcheck-query)
- [User Mutations](#user-mutations)
    - [Register](#register)
    - [Login](#login)
    - [Refresh](#refresh)
    - [Delete](#delete)
- [Fiat Account Mutations and Queries](#fiat-account-mutations-and-queries)
    - [Open Account](#open-account)
    - [Deposit](#deposit)
    - [Exchange](#exchange)
        - [Quote](#quote)
        - [Convert](#convert)
    - [Info](#info)
        - [Balance for a Specific Currency](#balance-for-a-specific-currency)
        - [Balance for all Currencies for a Client](#balance-for-all-currencies-for-a-client)
        - [Transaction Details for a Specific Transaction](#transaction-details-for-a-specific-transaction)
            - [External Transfer (deposit)](#external-transfer-deposit)
            - [Internal Transfer (currency conversion/exchange)](#internal-transfer-currency-conversionexchange)
        - [Transaction Details for a Specific Currency](#transaction-details-for-a-specific-currency)
            - [Initial Page](#initial-page)
            - [Subsequent Page](#subsequent-page)
- [Crypto Account Mutations and Queries](#crypto-account-mutations-and-queries)
    - [Open Account](#open-account-1)
    - [Offer](#offer)
        - [Purchase](#purchase)
        - [Sell](#sell)
    - [Exchange](#exchange-1)
        - [Purchase](#purchase-1)
        - [Sell](#sell-1)
  - [Info](#info)
      - [Balance for a Specific Currency](#balance-for-a-specific-currency-1)
      - [Transaction Details for a Specific Transaction](#transaction-details-for-a-specific-transaction-1)
          - [Purchase](#purchase-2)
          - [Sell](#sell-2)

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

### Healthcheck Query

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
      username: "someusername",
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
    username: "someusername",
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


### Fiat Account Mutations and Queries

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
        amount: 1345.67,
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
        sourceCurrency: "USD"
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
```graphql
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
query {
	balanceFiat(currencyCode: "USD") {
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
query {
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
query {
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

_Request:_ A valid `Transaction ID` must be provided as a parameter.

```graphql
query {
  transactionDetailsFiat(transactionID: "7d2fe42b-df1e-449f-875e-e9908ff24263") {
    currency
    amount
    transactedAt
    clientID
    txID
  }
}
```

_Response:_ Transaction-related details for a specific transaction. In the event of an external deposit, there will be
a single entry reporting the deposited amount. When querying for an internal transfer, two entries will be returned -
one for the source and the other for the destination accounts.

###### External Transfer (deposit)
```json
{
  "data": {
    "transactionFiat": [
      {
        "currency": "CAD",
        "amount": 368474.77,
        "transactedAt": "2023-05-09 18:30:51.985719 -0400 EDT",
        "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
        "txID": "7d2fe42b-df1e-449f-875e-e9908ff24263"
      }
    ]
  }
}
```

###### Internal Transfer (currency conversion/exchange)
```json
{
  "data": {
    "transactionDetailsFiat": [
      {
        "currency": "AED",
        "amount": 10000,
        "transactedAt": "2023-05-09 18:33:55.453689 -0400 EDT",
        "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
        "txID": "af4467a9-7c0a-4437-acf3-e5060509a5d9"
      },
      {
        "currency": "USD",
        "amount": 2723.24,
        "transactedAt": "2023-05-09 18:33:55.453689 -0400 EDT",
        "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
        "txID": "af4467a9-7c0a-4437-acf3-e5060509a5d9"
      }
    ]
  }
}
```
##### Transaction Details for a Specific Currency

_Request:_ A valid `Currency Code` must be provided as a parameter. The parameters accepted are listed below.
If a `pageCursor` is supplied, all other parameters except for the `pageSize` are ignored.

Optional:
* `pageCursor`: Defaults to 10.

Initial Page (required):
* `month`: Month for which the transactions are being requested.
* `year`: Year for which the transactions are being requested.
* `timezone`: Timezone for which the transactions are being requested.

```graphql
query {
  transactionDetailsAllFiat(input:{
    currency: "USD"
  	pageSize: "3"
    timezone: "-04:00"
    month: "5"
    year: "2023"
  }) {
    transactions {
      currency
      amount
      transactedAt
      clientID
      txID
    }
    links {
      pageCursor
    }
  }
}
```

Subsequent Pages (required)
* `pageCursor`: Hashed page cursor for the next page of data.

```graphql
query {
  transactionDetailsAllFiat(input:{
    currency: "USD"
  	pageSize: "3"
    pageCursor: "-GQBZ1LNxWCXItw7mek5Gumc4IwzUfH7yHN0aDJMecTULYvpDAHcjdkZUaGO_gGweET2_9H78mx5_81F2JsKwXwQot9UoFlU8IlHlTWlQArP"
  }) {
    transactions {
      currency
      amount
      transactedAt
      clientID
      txID
    }
    links {
      pageCursor
    }
  }
}
```

_Response:_ All Transaction-related details for a specific currency in a given timezone and date are returned. In the
event of an external deposit, there will be a single entry reporting the deposited amount. When querying for an internal
transfer, two entries will be returned - one for the source and the other for the destination accounts.

###### Initial Page
```json
{
  "data": {
    "transactionDetailsAllFiat": {
      "transactions": [
        {
          "currency": "USD",
          "amount": 100.11,
          "transactedAt": "2023-05-15 16:59:24.243332 -0400 EDT",
          "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
          "txID": "043d82a9-113b-4aa7-a3e1-029cc4728926"
        },
        {
          "currency": "USD",
          "amount": 100.11,
          "transactedAt": "2023-05-15 16:58:54.84774 -0400 EDT",
          "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
          "txID": "04ab99a6-c054-4592-b9cb-477369e0e9d8"
        },
        {
          "currency": "USD",
          "amount": 100.11,
          "transactedAt": "2023-05-15 16:57:45.752318 -0400 EDT",
          "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
          "txID": "1c57d150-9a93-4e4d-aef3-a8c3a14ff433"
        }
      ],
      "links": {
        "pageCursor": "-GQBZ1LNxWCXItw7mek5Gumc4IwzUfH7yHN0aDJMecTULYvpDAHcjdkZUaGO_gGweET2_9H78mx5_81F2JsKwXwQot9UoFlU8IlHlTWlQArP"
      }
    }
  }
}
```

###### Subsequent Page
```json
{
  "data": {
    "transactionDetailsAllFiat": {
      "transactions": [
        {
          "currency": "USD",
          "amount": 1345.67,
          "transactedAt": "2023-05-14 11:57:47.796057 -0400 EDT",
          "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
          "txID": "8522591d-6463-4cc6-9e3c-c456c98a6755"
        },
        {
          "currency": "USD",
          "amount": 2723.24,
          "transactedAt": "2023-05-09 18:33:55.453689 -0400 EDT",
          "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
          "txID": "af4467a9-7c0a-4437-acf3-e5060509a5d9"
        },
        {
          "currency": "USD",
          "amount": 10101.11,
          "transactedAt": "2023-05-09 18:29:48.729195 -0400 EDT",
          "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
          "txID": "1d7e1e70-0f9d-41b4-9f85-6dc310aa8f2d"
        }
      ],
      "links": {
        "pageCursor": ""
      }
    }
  }
}
```


<br/>


### Crypto Account Mutations and Queries

#### Open Account

_Request:_ All fields are required.

```graphql
mutation {
    openCrypto(ticker: "ETH") {
        clientID,
        ticker
    }
}
```

_Response:_ Confirmation information containing the `Client ID` and `Ticker` of the newly opened account.

```json
{
  "data": {
    "openCrypto": {
      "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe",
      "ticker": "ETH"
    }
  }
}
```

#### Offer

To convert between a Cryptocurrency and a Fiat currencies, the user must maintain open accounts in both the source and
destination currencies. The amount specified will be in the source currency and the amount to deposit into the
destination account will be calculated based on the exchange rate.

The workflow will involve getting a conversion rate quote, referred to as an `Offer`. The returned rate quote `Offer`
will only be valid for a two-minute time window. The expiration time will be returned to the user as a Unix timestamp.
The user must issue a subsequent request using the encrypted `Offer ID` to complete the transaction.

##### Purchase

_Request:_ All fields are required.
```graphql
mutation {
    offerCrypto(input: {
        sourceAmount: 1234.56
        sourceCurrency: "USD"
        destinationCurrency: "BTC"
        isPurchase: true
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
    "offerCrypto": {
      "priceQuote": {
        "clientID": "a83a2506-f812-476b-8e14-9fa100126518",
        "sourceAcc": "USD",
        "destinationAcc": "BTC",
        "rate": 0.00003779753759799514,
        "amount": 0.04666333
      },
      "debitAmount": 1234.56,
      "offerID": "VltcBxmGjFcDL4YV8-xWVSp3WEnuF5oVVyPI9p7DV-A5WGrXTmPvwa11VbJRoElt",
      "expires": 1686255413
    }
  }
}
```

##### Sell

_Request:_ All fields are required.
```graphql
mutation {
    offerCrypto(input: {
        sourceAmount: 1234.56
        sourceCurrency: "BTC"
        destinationCurrency: "USD"
        isPurchase: false
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
    "offerCrypto": {
      "priceQuote": {
        "clientID": "a83a2506-f812-476b-8e14-9fa100126518",
        "sourceAcc": "BTC",
        "destinationAcc": "USD",
        "rate": 26455.3975169303,
        "amount": 32660775.56
      },
      "debitAmount": 1234.56,
      "offerID": "YzLpRLex_bWKuNhXBji2wd0VkIxNnn3eYvBwRp204wjJIO2lDXv3jz73lr3LsL--",
      "expires": 1686255663
    }
  }
}
```

#### Exchange

Execute a Cryptocurrency purchase or sale using a valid exchange offer that must be obtained prior using the
`crypto/offer` mutation.

######  Purchase

_Request:_ All fields are required.
```graphql
mutation {
    exchangeCrypto(offerID: "roqzjmgIxlHMHWdSmJcRVby7RPvLEIzuMJ3ajH3bIr0YRukzd8XIL-rcUYRsE10R") {
        fiatTxReceipt{
            currency,
            amount,
            transactedAt,
            clientID,
            txID,
        },
        cryptoTxReceipt{
            ticker,
            amount,
            transactedAt,
            clientID,
            txID,
        },
    }
}
```

_Response:_ A receipt with the Fiat and Cryptocurrency transaction information.
```json
{
  "data": {
    "exchangeCrypto": {
      "fiatTxReceipt": {
        "currency": "USD",
        "amount": -1234.56,
        "transactedAt": "2023-06-08 17:44:27.766461 -0400 EDT",
        "clientID": "a83a2506-f812-476b-8e14-9fa100126518",
        "txID": "4650fa28-1ad5-46fc-97a8-15c21ee8608e"
      },
      "cryptoTxReceipt": {
        "ticker": "BTC",
        "amount": 0.04653972,
        "transactedAt": "2023-06-08 17:44:27.766461 -0400 EDT",
        "clientID": "a83a2506-f812-476b-8e14-9fa100126518",
        "txID": "4650fa28-1ad5-46fc-97a8-15c21ee8608e"
      }
    }
  }
}
```

######  Sell

_Request:_ All fields are required.
```graphql
mutation {
    exchangeCrypto(offerID: "LQq07LHQdqCbwuXuxkH-rW6-WMcBhi2RG9q9HSKOwh8TcxzG_DWg_iOW9m9xdZy8") {
        fiatTxReceipt{
            currency,
            amount,
            transactedAt,
            clientID,
            txID,
        },
        cryptoTxReceipt{
            ticker,
            amount,
            transactedAt,
            clientID,
            txID,
        },
    }
}
```

_Response:_ A receipt with the Fiat and Cryptocurrency transaction information.
```json
{
  "data": {
    "exchangeCrypto": {
      "fiatTxReceipt": {
        "currency": "USD",
        "amount": 864247.73,
        "transactedAt": "2023-06-08 17:06:03.192364 -0400 EDT",
        "clientID": "a83a2506-f812-476b-8e14-9fa100126518",
        "txID": "b4df7d86-36b0-407b-8acf-21cccbc88386"
      },
      "cryptoTxReceipt": {
        "ticker": "BTC",
        "amount": -32.45,
        "transactedAt": "2023-06-08 17:06:03.192364 -0400 EDT",
        "clientID": "a83a2506-f812-476b-8e14-9fa100126518",
        "txID": "b4df7d86-36b0-407b-8acf-21cccbc88386"
      }
    }
  }
}
```

#### Info

##### Balance for a Specific Currency

_Request:_ A valid Cryptocurrency ticker must be provided as a query parameter.
```graphql
query {
    balanceCrypto(ticker:"BTC") {
        ticker,
        balance,
        lastTx,
        lastTxTs,
        createdAt,
        clientID,
    }
}
```

_Response:_ Account balance related details associated with the currency.
```json
{
  "data": {
    "balanceCrypto": {
      "ticker": "BTC",
      "balance": 46.69881177,
      "lastTx": 46.69881177,
      "lastTxTs": "2023-06-09 16:51:55.520098 -0400 EDT",
      "createdAt": "2023-06-09 16:51:03.466403 -0400 EDT",
      "clientID": "6bc1d17e-68c6-4b82-80fd-542c4d3aba9b"
    }
  }
}
```

##### Transaction Details for a Specific Transaction

_Request:_ A valid `Transaction ID` must be provided as a query parameter.
```graphql
query {
  transactionDetailsCrypto(transactionID: "05cef33f-2082-48c4-ad08-e0f8dc5d4444")
}
```

_Response:_ Transaction-related details for a specific transaction. There will be one entry for the Fiat currency
account and another for the Cryptocurrency account.

###### Purchase
```json
{
  "data": {
    "transactionDetailsCrypto": [
      {
        "currency": "USD",
        "amount": "-0.32",
        "transactedAt": "2023-06-09T17:25:01.62373-04:00",
        "clientID": "6bc1d17e-68c6-4b82-80fd-542c4d3aba9b",
        "txID": "05cef33f-2082-48c4-ad08-e0f8dc5d4444"
      },
      {
        "ticker": "BTC",
        "amount": "0.0000121",
        "transactedAt": "2023-06-09T17:25:01.62373-04:00",
        "clientID": "6bc1d17e-68c6-4b82-80fd-542c4d3aba9b",
        "txID": "05cef33f-2082-48c4-ad08-e0f8dc5d4444"
      }
    ]
  }
}
```

###### Sell
```json
{
  "data": {
    "exchangeCrypto": {
      "fiatTxReceipt": {
        "currency": "USD",
        "amount": 9410.35,
        "transactedAt": "2023-06-09 17:34:27.727458 -0400 EDT",
        "clientID": "6bc1d17e-68c6-4b82-80fd-542c4d3aba9b",
        "txID": "0cadcb76-8d26-4a1a-bf03-d3392c80d57b"
      },
      "cryptoTxReceipt": {
        "ticker": "BTC",
        "amount": -0.356,
        "transactedAt": "2023-06-09 17:34:27.727458 -0400 EDT",
        "clientID": "6bc1d17e-68c6-4b82-80fd-542c4d3aba9b",
        "txID": "0cadcb76-8d26-4a1a-bf03-d3392c80d57b"
      }
    }
  }
}
```
