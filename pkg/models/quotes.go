package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// FiatQuote is the quote returned from the Exchange Rates Data API.
type FiatQuote struct {
	Info    FiatInfo        `json:"info"`
	Query   FiatQuery       `json:"query"`
	Date    string          `json:"date"`
	Result  decimal.Decimal `json:"result"`
	Success bool            `json:"success"`
}

// FiatInfo contains the rate and the UTC timestamp of the price quote.
type FiatInfo struct {
	Rate      decimal.Decimal `json:"rate"`
	Timestamp time.Time       `json:"timestamp"`
}

// FiatQuery contains the information used to query for the quote.
type FiatQuery struct {
	From   string          `json:"from"`
	To     string          `json:"to"`
	Amount decimal.Decimal `json:"amount"`
}
