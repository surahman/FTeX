package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.31

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/surahman/FTeX/pkg/common"
	"github.com/surahman/FTeX/pkg/constants"
	graphql_generated "github.com/surahman/FTeX/pkg/graphql/generated"
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
)

// Amount is the resolver for the amount field.
func (r *cryptoJournalResolver) Amount(ctx context.Context, obj *postgres.CryptoJournal) (float64, error) {
	return obj.Amount.InexactFloat64(), nil
}

// TransactedAt is the resolver for the transactedAt field.
func (r *cryptoJournalResolver) TransactedAt(ctx context.Context, obj *postgres.CryptoJournal) (string, error) {
	return obj.TransactedAt.Time.String(), nil
}

// ClientID is the resolver for the clientID field.
func (r *cryptoJournalResolver) ClientID(ctx context.Context, obj *postgres.CryptoJournal) (string, error) {
	return obj.ClientID.String(), nil
}

// TxID is the resolver for the txID field.
func (r *cryptoJournalResolver) TxID(ctx context.Context, obj *postgres.CryptoJournal) (string, error) {
	return obj.TxID.String(), nil
}

// OpenCrypto is the resolver for the openCrypto field.
func (r *mutationResolver) OpenCrypto(ctx context.Context, ticker string) (*models.CryptoOpenAccountResponse, error) {
	var (
		clientID   uuid.UUID
		errMessage string
		err        error
	)

	if clientID, _, err = AuthorizationCheck(ctx, r.auth, r.logger, r.authHeaderKey); err != nil {
		return nil, errors.New("authorization failure")
	}

	if _, errMessage, err = common.HTTPCryptoOpen(r.db, r.logger, clientID, ticker); err != nil {
		return nil, errors.New(errMessage)
	}

	return &models.CryptoOpenAccountResponse{ClientID: clientID.String(), Ticker: ticker}, nil
}

// OfferCrypto is the resolver for the offerCrypto field.
func (r *mutationResolver) OfferCrypto(ctx context.Context, input models.HTTPCryptoOfferRequest) (*models.HTTPExchangeOfferResponse, error) {
	var (
		clientID      uuid.UUID
		err           error
		offer         models.HTTPExchangeOfferResponse
		statusMessage string
	)

	if clientID, _, err = AuthorizationCheck(ctx, r.auth, r.logger, r.authHeaderKey); err != nil {
		return nil, errors.New("authorization failure")
	}

	if offer, _, statusMessage, err = common.HTTPCryptoOffer(r.auth, r.cache, r.logger, r.quotes,
		clientID, input.SourceCurrency, input.DestinationCurrency, input.SourceAmount, *input.IsPurchase); err != nil {
		if statusMessage == constants.GetInvalidRequest() {
			statusMessage = err.Error()
		}

		return nil, errors.New(statusMessage)
	}

	offer.ClientID = clientID

	return &offer, nil
}

// ExchangeCrypto is the resolver for the exchangeCrypto field.
func (r *mutationResolver) ExchangeCrypto(ctx context.Context, offerID string) (*models.HTTPCryptoTransferResponse, error) {
	var (
		clientID      uuid.UUID
		err           error
		receipt       models.HTTPCryptoTransferResponse
		statusMessage string
	)

	if clientID, _, err = AuthorizationCheck(ctx, r.auth, r.logger, r.authHeaderKey); err != nil {
		return nil, errors.New("authorization failure")
	}

	if receipt, _, statusMessage, err = common.HTTPExchangeCrypto(r.auth, r.cache, r.db, r.logger, clientID, offerID);
		err != nil {
		return nil, errors.New(statusMessage)
	}

	return &receipt, nil
}

// SourceAmount is the resolver for the sourceAmount field.
func (r *cryptoOfferRequestResolver) SourceAmount(ctx context.Context, obj *models.HTTPCryptoOfferRequest, data float64) error {
	obj.SourceAmount = decimal.NewFromFloat(data)

	return nil
}

// CryptoJournal returns graphql_generated.CryptoJournalResolver implementation.
func (r *Resolver) CryptoJournal() graphql_generated.CryptoJournalResolver {
	return &cryptoJournalResolver{r}
}

// CryptoOfferRequest returns graphql_generated.CryptoOfferRequestResolver implementation.
func (r *Resolver) CryptoOfferRequest() graphql_generated.CryptoOfferRequestResolver {
	return &cryptoOfferRequestResolver{r}
}

type cryptoJournalResolver struct{ *Resolver }
type cryptoOfferRequestResolver struct{ *Resolver }
