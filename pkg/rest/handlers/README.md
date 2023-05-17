# HTTP REST API Endpoints

The REST API schema can be tested and reviewed through the Swagger UI that is exposed when the server is started.

<br/>

## Table of contents

# Table of contents

- [Authorization Response](#authorization-response)
- [Error Response](#error-response)
- [Success Response](#success-response)
- [Healthcheck Endpoint `/health`](#healthcheck-endpoint-health)
- [User Endpoints `/user`](#user-endpoints-user)
  - [Register `/register`](#register-register)
  - [Login `/login`](#login-login)
  - [Refresh `/refresh`](#refresh-refresh)
  - [Delete `/delete`](#delete-delete)
- [Fiat Accounts Endpoints `/fiat`](#fiat-accounts-endpoints-fiat)
  - [Open `/open`](#open-open)
  - [Deposit `/deposit`](#deposit-deposit)
  - [Exchange `/exchange`](#exchange-exchange)
    - [Quote `/offer`](#quote-offer)
    - [Convert `/convert`](#convert-convert)
  - [Info `/info`](#info-info)
    - [Balance for a Specific Currency `/balance/{currencyCode}`](#balance-for-a-specific-currency-balancecurrencycode)
    - [Balance for all Currencies for a Client `/fiat/info/balance/?pageCursor=PaGeCuRs0R==&pageSize=3`](#balance-for-all-currencies-for-a-client-fiatinfobalancepagecursorpagecurs0rpagesize3)
    - [Transaction Details for a Specific Transaction `/transaction/{transactionID}`](#transaction-details-for-a-specific-transaction-transactiontransactionid)
      - [External Transaction (deposit)](#external-transaction-deposit)
      - [Internal Transfer (currency conversion/exchange)](#internal-transfer-currency-conversionexchange)
    - [Transaction Details for a Specific Currency `/transaction/all/{currencyCode}`](#transaction-details-for-a-specific-currency-transactionallcurrencycode)
      - [Initial Page](#initial-page)
      - [Subsequent Page](#subsequent-page)

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

### User Endpoints `/user`

#### Register `/register`

Register a new user account.

_Request:_ All fields are required.
```json
{
  "email": "string",
  "first_name": "string",
  "last_name": "string",
  "password": "string",
  "username": "string"
}
```

_Response:_ A valid JWT will be returned as an authorization response.

#### Login `/login`

Log into a valid user account by providing valid user credentials.

_Request:_ All fields are required.
```json
{
  "password": "string",
  "username": "string"
}
```

_Response:_ A valid JWT will be returned as an authorization response.

#### Refresh `/refresh`

Refresh a valid but expiring JWT within the refresh threshold window. The client must refresh the token before
expiration but within the refresh threshold specified in the `JWT` authorization response.

_Request:_ A valid JWT must be provided in the request header and will be validated with a fresh token issued against it.
_Response:_ A valid JWT will be returned as an authorization response.
```json
{
  "expires": "expiration time string",
  "token": "token string"
}
```

#### Delete `/delete`

Soft-delete an active and valid user account by completing the acknowledgment confirmation correctly and providing
valid user credentials.

_Request:_ All fields are required and a valid JWT must be provided in the header. The user must supply their login
credentials as well as complete the confirmation message `I understand the consequences, delete my user
account **USERNAME HERE**`
```json
{
  "confirmation": "I understand the consequences, delete my user account <USERNAME HERE>",
  "password": "password string",
  "username": "username string"
}
```

_Response:_ A confirmation message will be returned as a success response.


<br/>

### Fiat Accounts Endpoints `/fiat`

Fiat accounts endpoints provide access to deposit money into and across Fiat accounts belonging to the same client.

#### Open `/open`

Open a Fiat account with an empty balance for a logged-in user in a specific currency. The
[`ISO 4217`](https://www.iso.org/iso-4217-currency-codes.html) currency code for the new account to be opened must be
provided in the request.

_Request:_ All fields are required.
```json
{
  "currency": "USD"
}
```
_Response:_ The Client ID and `ISO 4217` currency code that the Fiat account was set up for.
```json
{
  "message": "account created",
  "payload": [
    "cbe0d46b-7668-45f4-8519-6f291914b14c",
    "USD"
  ]
}
```

#### Deposit `/deposit`

Deposit money into a Fiat account for a specific currency and amount. An account for the currency must already be opened
for the deposit to succeed.

_Request:_ All fields are required.
```json
{
  "currency": "USD",
  "amount": 1921.68
}
```

_Response:_ A confirmation of the transaction with the particulars of the transfer.
```json
{
  "message": "funds successfully transferred",
  "payload": {
    "txId": "f9a3bfe1-de43-47cc-a634-508181652d75",
    "clientId": "cbe0d46b-7668-45f4-8519-6f291914b14c",
    "txTimestamp": "2023-04-23T17:09:07.468161-04:00",
    "balance": "3259.57",
    "lastTx": "1921.68",
    "currency": "USD"
  }
}
```

#### Exchange `/exchange`

To convert between Fiat currencies, the user must maintain open accounts in both the source and destination Fiat currencies.
The amount specified will be in the source currency and the amount to deposit into the destination account will be calculated
based on the exchange rate.

The workflow will involve getting a conversion rate quote, referred to as an `Offer`. The returned rate quote `Offer` will
only be valid for a two-minute time window. The expiration time will be returned to the user as a Unix timestamp. The user
must issue a subsequent request using the encrypted `Offer ID` to complete the transaction.

##### Quote `/offer`

_Request:_ All fields are required.
```json
{
  "destinationCurrency": "CAD",
  "sourceAmount": 1000,
  "sourceCurrency": "USD"
}
```

_Response:_ A rate quote with an encrypted `Offer ID`.
```json
{
  "message": "conversion rate offer",
  "payload": {
    "offer": {
      "clientId": "a8d55c17-09cc-4805-a7f7-4c5038a97b32",
      "sourceAcc": "CAD",
      "destinationAcc": "USD",
      "rate": "0.732467",
      "amount": "73.44"
    },
    "debitAmount": "100.26",
    "offerId": "m45QsqDVbzi2bVasVzWJ3cKPKy98BUDhyicK4cOwIbZXdydUXXMzW9PFx82OAz7y",
    "expires": 1682878564
  }
}
```

##### Convert `/convert`

_Request:_ All fields are required.
```json
{
  "offerId": "m45QsqDVbzi2bVasVzWJ3cKPKy98BUDhyicK4cOwIbZXdydUXXMzW9PFx82OAz7y"
}
```

_Response:_ A transaction receipt with the details of the source and destination accounts and transaction details.
```json
{
  "message": "funds exchange transfer successful",
  "payload": {
    "sourceReceipt": {
      "txId": "da3f100a-2f47-4879-a3b7-bb0517c3b1ac",
      "clientId": "a8d55c17-09cc-4805-a7f7-4c5038a97b32",
      "txTimestamp": "2023-04-30T17:06:54.654345-04:00",
      "balance": "1338.43",
      "lastTx": "-100.26",
      "currency": "CAD"
    },
    "destinationReceipt": {
      "txId": "da3f100a-2f47-4879-a3b7-bb0517c3b1ac",
      "clientId": "a8d55c17-09cc-4805-a7f7-4c5038a97b32",
      "txTimestamp": "2023-04-30T17:06:54.654345-04:00",
      "balance": "21714.35",
      "lastTx": "73.44",
      "currency": "USD"
    }
  }
}
```

#### Info `/info`

##### Balance for a Specific Currency `/balance/{currencyCode}`

_Request:_ A valid currency code must be provided as a path parameter.

_Response:_ Account balance related details associated with the currency.
```json
{
  "message": "account balance",
  "payload": {
    "currency": "USD",
    "balance": "22813.05",
    "lastTx": "1098.7",
    "lastTxTs": "2023-04-30T17:15:43.605776-04:00",
    "createdAt": "2023-04-28T17:24:11.540235-04:00",
    "clientID": "a8d55c17-09cc-4805-a7f7-4c5038a97b32"
  }
}
```

##### Balance for all Currencies for a Client `/fiat/info/balance/?pageCursor=PaGeCuRs0R==&pageSize=3`

_Request:_ The initial request can only contain an optional `page size`, which if not provided will default to 10. The
subsequent responses will contain encrypted page cursors that must be specified to retrieve the following page of data.

> fiat/info/balance/?pageCursor=QW9bg6pXqXdwegEf7PVEuqoPzAJ28tO0r4TSh-t8qQ==&pageSize=3


_Response:_ Account balances for the Client will be limited to the `Page Size` specified and is `10` by default. A
`Page Cursor` link will be supplied if there are subsequent pages of data to be retrieved in the `links.nextPage` JSON
field.

```json
{
  "message": "account balances",
  "payload": {
    "accountBalances": [
      {
        "currency": "AED",
        "balance": "30903.7",
        "lastTx": "-10000",
        "lastTxTs": "2023-05-09T18:33:55.453689-04:00",
        "createdAt": "2023-05-09T18:29:16.74704-04:00",
        "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe"
      },
      {
        "currency": "CAD",
        "balance": "368474.77",
        "lastTx": "368474.77",
        "lastTxTs": "2023-05-09T18:30:51.985719-04:00",
        "createdAt": "2023-05-09T18:29:08.746285-04:00",
        "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe"
      },
      {
        "currency": "EUR",
        "balance": "1536.45",
        "lastTx": "1536.45",
        "lastTxTs": "2023-05-09T18:31:32.213239-04:00",
        "createdAt": "2023-05-09T18:29:21.365991-04:00",
        "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe"
      }
    ],
    "links": {
      "nextPage": "?pageCursor=zTrzwXDqdxG-9aQ6sWVCwfJNs--anH9mQEMVKlDsvA==&pageSize=3"
    }
  }
}
```
```json
{
  "message": "account balances",
  "payload": {
    "accountBalances": [
      {
        "currency": "USD",
        "balance": "12824.35",
        "lastTx": "2723.24",
        "lastTxTs": "2023-05-09T18:33:55.453689-04:00",
        "createdAt": "2023-05-09T18:29:04.345387-04:00",
        "clientID": "70a0caf3-3fb2-4a96-b6e8-991252a88efe"
      }
    ],
    "links": {}
  }
}
```

##### Transaction Details for a Specific Transaction `/transaction/{transactionID}`

_Request:_ A valid `Transaction ID` must be provided as a path parameter.

_Response:_ Transaction-related details for a specific transaction. In the event of an external deposit, there will be
a single entry reporting the deposited amount. When querying for an internal transfer, two entries will be returned -
one for the source and the other for the destination accounts.

###### External Transaction (deposit)
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
##### Transaction Details for a Specific Currency `/transaction/all/{currencyCode}`

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
    "links": {}
  }
}
```
