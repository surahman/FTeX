package models

import (
	"github.com/shopspring/decimal"
	modelsPostgres "github.com/surahman/FTeX/pkg/models/postgres"
	"github.com/surahman/FTeX/pkg/postgres"
)

// JWTAuthResponse is the response to a successful login and token refresh.
// The client uses the expires field on to know when to refresh the token.
//
//nolint:lll
type JWTAuthResponse struct {
	Token     string `json:"token"     validate:"required" yaml:"token"`     // JWT string sent to and validated by the server.
	Expires   int64  `json:"expires"   validate:"required" yaml:"expires"`   // Expiration time as unix time stamp. Strictly used by client to gauge when to refresh the token.
	Threshold int64  `json:"threshold" validate:"required" yaml:"threshold"` // The window in seconds before expiration during which the token can be refreshed.
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
	Confirmation string `json:"confirmation" validate:"required" yaml:"confirmation"`
}

// HTTPOpenCurrencyAccountRequest is a request to open an account in a specified Fiat currency.
type HTTPOpenCurrencyAccountRequest struct {
	Currency string `json:"currency" validate:"required" yaml:"currency"`
}

// HTTPDepositCurrencyRequest is a request to deposit currency in to a specified Fiat currency.
type HTTPDepositCurrencyRequest struct {
	Amount   decimal.Decimal `json:"amount"   validate:"required,gt=0" yaml:"amount"`
	Currency string          `json:"currency" validate:"required"      yaml:"currency"`
}

// HTTPExchangeOfferRequest is a request to convert a source to destination currency in the source currency amount.
type HTTPExchangeOfferRequest struct {
	SourceCurrency      string          `json:"sourceCurrency"      validate:"required"      yaml:"sourceCurrency"`
	DestinationCurrency string          `json:"destinationCurrency" validate:"required"      yaml:"destinationCurrency"`
	SourceAmount        decimal.Decimal `json:"sourceAmount"        validate:"required,gt=0" yaml:"sourceAmount"`
}

// HTTPCryptoOfferRequest is a request to convert a source to destination currency in the source currency amount.
type HTTPCryptoOfferRequest struct {
	HTTPExchangeOfferRequest `json:"request"    validate:"required" yaml:"request"`
	IsPurchase               *bool `json:"isPurchase" validate:"required" yaml:"isPurchase"`
}

// HTTPExchangeOfferResponse is an offer to convert a source to destination currency in the source currency amount.
type HTTPExchangeOfferResponse struct {
	PriceQuote       `json:"offer"                      yaml:"offer"`
	DebitAmount      decimal.Decimal `json:"debitAmount"                yaml:"debitAmount"`
	OfferID          string          `json:"offerId"                    yaml:"offerId"`
	Expires          int64           `json:"expires"                    yaml:"expires"`
	IsCryptoPurchase bool            `json:"isCryptoPurchase,omitempty" yaml:"isCryptoPurchase,omitempty"`
	IsCryptoSale     bool            `json:"isCryptoSale,omitempty"     yaml:"isCryptoSale,omitempty"`
}

// HTTPTransferRequest is the request to accept and execute an existing exchange offer.
type HTTPTransferRequest struct {
	OfferID string `json:"offerId" validate:"required" yaml:"offerId"`
}

// HTTPFiatTransferResponse is the response to a successful Fiat exchange conversion request.
type HTTPFiatTransferResponse struct {
	SrcTxReceipt *postgres.FiatAccountTransferResult `json:"sourceReceipt"      yaml:"sourceReceipt"`
	DstTxReceipt *postgres.FiatAccountTransferResult `json:"destinationReceipt" yaml:"destinationReceipt"`
}

// HTTPCryptoTransferResponse is the response to a successful Cryptocurrency purchase/sale request.
type HTTPCryptoTransferResponse struct {
	FiatTxReceipt   *postgres.FiatJournal   `json:"fiatReceipt"   yaml:"fiatReceipt"`
	CryptoTxReceipt *postgres.CryptoJournal `json:"cryptoReceipt" yaml:"cryptoReceipt"`
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

// HTTPCryptoDetailsPaginated is the response to paginated account details request. It returns a link to the next page
// of information.
type HTTPCryptoDetailsPaginated struct {
	AccountBalances []postgres.CryptoAccount `json:"accountBalances"`
	Links           HTTPLinks                `json:"links"`
}

// HTTPCryptoTransactionsPaginated is the response to paginated account transactions request. It returns a link to the
// next page of information.
type HTTPCryptoTransactionsPaginated struct {
	TransactionDetails []postgres.CryptoJournal `json:"transactionDetails"`
	Links              HTTPLinks                `json:"links,omitempty"`
}
