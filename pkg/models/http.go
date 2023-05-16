package models

import (
	"github.com/shopspring/decimal"
	modelsPostgres "github.com/surahman/FTeX/pkg/models/postgres"
	"github.com/surahman/FTeX/pkg/postgres"
)

// JWTAuthResponse is the response to a successful login and token refresh. The expires field is used on by the client to
// know when to refresh the token.
//
//nolint:lll
type JWTAuthResponse struct {
	Token     string `json:"token" yaml:"token" validate:"required"`         // JWT string sent to and validated by the server.
	Expires   int64  `json:"expires" yaml:"expires" validate:"required"`     // Expiration time as unix time stamp. Strictly used by client to gauge when to refresh the token.
	Threshold int64  `json:"threshold" yaml:"threshold" validate:"required"` // The window in seconds before expiration during which the token can be refreshed.
}

// HTTPError is a generic error message that is returned to the requester.
type HTTPError struct {
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
	Payload any    `json:"payload,omitempty" yaml:"payload,omitempty"`
}

// HTTPSuccess is a generic success message that is returned to the requester.
type HTTPSuccess struct {
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
	Payload any    `json:"payload,omitempty" yaml:"payload,omitempty"`
}

// HTTPDeleteUserRequest is the request to mark a user account as deleted. The user must supply their login credentials
// as well as a confirmation message.
type HTTPDeleteUserRequest struct {
	modelsPostgres.UserLoginCredentials
	Confirmation string `json:"confirmation" yaml:"confirmation" validate:"required"`
}

// HTTPOpenCurrencyAccountRequest is a request to open an account in a specified Fiat or Cryptocurrency.
type HTTPOpenCurrencyAccountRequest struct {
	Currency string `json:"currency" yaml:"currency" validate:"required"`
}

// HTTPDepositCurrencyRequest is a request to deposit currency in to a specified Fiat or Cryptocurrency.
type HTTPDepositCurrencyRequest struct {
	Amount   decimal.Decimal `json:"amount" yaml:"amount" validate:"required,gt=0"`
	Currency string          `json:"currency" yaml:"currency" validate:"required"`
}

// HTTPFiatExchangeOfferRequest is a request to convert a source to destination currency in the source currency amount.
type HTTPFiatExchangeOfferRequest struct {
	SourceCurrency      string          `json:"sourceCurrency" yaml:"sourceCurrency" validate:"required"`
	DestinationCurrency string          `json:"destinationCurrency" yaml:"destinationCurrency" validate:"required"`
	SourceAmount        decimal.Decimal `json:"sourceAmount" yaml:"sourceAmount" validate:"required,gt=0"`
}

// HTTPFiatExchangeOfferResponse is an offer to convert a source to destination currency in the source currency amount.
type HTTPFiatExchangeOfferResponse struct {
	PriceQuote  `json:"offer" yaml:"offer"`
	DebitAmount decimal.Decimal `json:"debitAmount" yaml:"debitAmount"`
	OfferID     string          `json:"offerId" yaml:"offerId"`
	Expires     int64           `json:"expires" yaml:"expires"`
}

// HTTPFiatTransferRequest is the request to accept and execute an existing exchange offer.
type HTTPFiatTransferRequest struct {
	OfferID string `json:"offerId" yaml:"offerId" validate:"required"`
}

// HTTPFiatTransferResponse is the response to a successful exchange conversion request.
type HTTPFiatTransferResponse struct {
	SrcTxReceipt *postgres.FiatAccountTransferResult `json:"sourceReceipt" yaml:"sourceReceipt"`
	DstTxReceipt *postgres.FiatAccountTransferResult `json:"destinationReceipt" yaml:"destinationReceipt"`
}

// HTTPFiatDetailsPaginated is the response to paginated account details request. It returns a link to the next page of
// information.
type HTTPFiatDetailsPaginated struct {
	AccountBalances []postgres.FiatAccount `json:"accountBalances"`
	Links           HTTPLinks              `json:"links"`
}

// HTTPFiatTransactionsPaginated is the response to paginated account transactions request. It returns a link to the
// next page of information.
type HTTPFiatTransactionsPaginated struct {
	TransactionDetails []postgres.FiatJournal `json:"transactionDetails"`
	Links              HTTPLinks              `json:"links,omitempty"`
}

// HTTPLinks are links used in HTTP responses to retrieve pages of information.
type HTTPLinks struct {
	NextPage   string `json:"nextPage,omitempty"`
	PageCursor string `json:"pageCursor,omitempty"`
}
