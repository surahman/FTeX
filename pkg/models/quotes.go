package models

import (
	"github.com/shopspring/decimal"
)

// FiatQuote is the quote returned from the Fiat Exchange Rates Data API.
type FiatQuote struct {
	Info    FiatInfo        `json:"info"`
	Query   FiatQuery       `json:"query"`
	Error   FiatError       `json:"error,omitempty"`
	Date    string          `json:"date"`
	Result  decimal.Decimal `json:"result"`
	Success bool            `json:"success"`
}

// FiatInfo contains the rate and the UTC timestamp of the price quote.
type FiatInfo struct {
	Rate      decimal.Decimal `json:"rate"`
	Timestamp int64           `json:"timestamp"`
}

// FiatQuery contains the information used to query for the quote.
type FiatQuery struct {
	From   string          `json:"from"`
	To     string          `json:"to"`
	Amount decimal.Decimal `json:"amount"`
}

// FiatError is populated if there is an error whilst running a query.
type FiatError struct {
	Code int    `json:"code"`
	Type string `json:"type"`
	Info string `json:"info"`
}

// CryptoQuote is the quote returned from the Crypto Exchange Rates Data API.
//
//nolint:tagliatelle
type CryptoQuote struct {
	BaseCurrency  string          `json:"asset_id_base"`
	QuoteCurrency string          `json:"asset_id_quote"`
	Time          string          `json:"time"`
	Rate          decimal.Decimal `json:"rate"`
}
