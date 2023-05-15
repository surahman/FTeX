package graphql

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/models"
)

func TestResolver_PriceQuoteResolvers(t *testing.T) {
	t.Parallel()

	resolver := priceQuoteResolver{}

	clientID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate client ID.")

	rate := decimal.NewFromFloat(1234.56)
	amount := decimal.NewFromFloat(78910.11)

	priceQuote := &models.PriceQuote{
		ClientID:       clientID,
		SourceAcc:      "",
		DestinationAcc: "",
		Rate:           rate,
		Amount:         amount,
	}

	t.Run("Client ID", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.ClientID(context.TODO(), priceQuote)
		require.NoError(t, err, "failed to resolve client id.")
		require.Equal(t, clientID.String(), result, "client id mismatched.")
	})

	t.Run("Rate", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.Rate(context.TODO(), priceQuote)
		require.NoError(t, err, "failed to resolve rate.")
		require.Equal(t, rate.InexactFloat64(), result, "rate mismatched.")
	})

	t.Run("Amount", func(t *testing.T) {
		t.Parallel()

		result, err := resolver.Amount(context.TODO(), priceQuote)
		require.NoError(t, err, "failed to resolve amount.")
		require.Equal(t, amount.InexactFloat64(), result, "amount mismatched.")
	})
}
