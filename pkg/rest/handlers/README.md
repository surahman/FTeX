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
  - [Delete](#delete)
- [Fiat Accounts Endpoints `/fiat/`](#fiat-accounts-endpoints-fiat)
  - [Open](#open)
  - [Deposit](#deposit)

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

#### Login

Log into a valid user account by providing valid user credentials.

_Request:_ All fields are required.
```json
{
  "password": "string",
  "username": "string"
}
```

_Response:_ A valid JWT will be returned as an authorization response.

#### Refresh

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

#### Delete

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

### Fiat Accounts Endpoints `/fiat/`

Fiat accounts endpoints provide access to deposit money into and across Fiat accounts belonging to the same client.

#### Open

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

#### Deposit

Deposit money into a Fiat account for a specific currency and amount. An account for the currency must already be opened for the deposit to succeed.

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
    "txTimestamp": "2023-04-23T11:09:07.468161-04:00",
    "balance": "3259.57",
    "lastTx": "1921.68",
    "currency": "USD"
  }
}
```

### Convert

To convert between Fiat currencies, the user must maintain open accounts in both the source and destination Fiat currencies.

The workflow will involve getting a conversion rate quote, referred to as an `Offer`. The returned rate quote `Offer` will
only be valid for a two-minute time window. The expiration time will be returned to the user as a Unix timestamp. The user
must issue a subsequent request using the encrypted `Offer ID` to complete the transaction.

##### Quote `/fiat/exchange/offer`

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
      "sourceAcc": "USD",
      "destinationAcc": "CAD",
      "rate": "1.35463",
      "amount": "1354.63"
    },
    "offerId": "8nXnmfvKTNL5dgxNIjmnPNO7e7vcBxCL6iHcqwGxdE1VOPhfTWwMblZk-kAJiWhc",
    "expires": 1682716375000
  }
}
```

##### Convert `/fiat/exchange/convert`
