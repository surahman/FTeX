package models

import (
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

// PriceQuote is the quote provided to the end-user requesting a transfer and will be stored in the Redis cache.
type PriceQuote struct {
	ClientID       uuid.UUID       `json:"clientId" validate:"required"`
	SourceAcc      string          `json:"sourceAcc" validate:"required"`
	DestinationAcc string          `json:"destinationAcc" validate:"required"`
	Rate           decimal.Decimal `json:"rate" validate:"required"`
	Amount         decimal.Decimal `json:"amount" validate:"required"`
}
