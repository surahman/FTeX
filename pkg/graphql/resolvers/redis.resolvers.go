package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.31

import (
	"context"

	graphql_generated "github.com/surahman/FTeX/pkg/graphql/generated"
	"github.com/surahman/FTeX/pkg/models"
)

// ClientID is the resolver for the ClientID field.
func (r *priceQuoteResolver) ClientID(ctx context.Context, obj *models.PriceQuote) (string, error) {
	return obj.ClientID.String(), nil
}

// Rate is the resolver for the Rate field.
func (r *priceQuoteResolver) Rate(ctx context.Context, obj *models.PriceQuote) (float64, error) {
	return obj.Rate.InexactFloat64(), nil
}

// Amount is the resolver for the Amount field.
func (r *priceQuoteResolver) Amount(ctx context.Context, obj *models.PriceQuote) (float64, error) {
	return obj.Amount.InexactFloat64(), nil
}

// PriceQuote returns graphql_generated.PriceQuoteResolver implementation.
func (r *Resolver) PriceQuote() graphql_generated.PriceQuoteResolver { return &priceQuoteResolver{r} }

type priceQuoteResolver struct{ *Resolver }
