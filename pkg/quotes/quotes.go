package quotes

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/imroc/req/v3"
	"github.com/shopspring/decimal"
	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/models"
	"go.uber.org/zap"
)

// Mock Quotes interface stub generation. This is local to the Quotes package.
//go:generate mockgen -destination=quotes_mocks.go -package=quotes github.com/surahman/FTeX/pkg/quotes Quotes

// Mock Quotes interface stub generation.
//go:generate mockgen -destination=../mocks/mock_quotes.go -package=mocks github.com/surahman/FTeX/pkg/quotes Quotes

// Quotes is the interface through which the currency quote services can be accessed. Created to support mock testing.
type Quotes interface {
	// fiatQuote will retrieve a quote for a Fiat currency price.
	fiatQuote(source, destination string, sourceAmount decimal.Decimal) (models.FiatQuote, error)

	// cryptoQuote will retrieve a quote for a Cryptocurrency price.
	cryptoQuote(source, destination string) (models.CryptoQuote, error)

	// FiatConversion will convert a source currency, in a given amount, to the destination currency.
	FiatConversion(source, destination string, amount decimal.Decimal,
		fiatQuote func(source, destination string, amount decimal.Decimal) (models.FiatQuote, error)) (
		decimal.Decimal, decimal.Decimal, error)
}

// Check to ensure the Redis interface has been implemented.
var _ Quotes = &quotesImpl{}

// quoteImpl implements the Quote interface and contains the logic to interface with currency price services.
type quotesImpl struct {
	clientCrypto *req.Client
	clientFiat   *req.Client
	conf         *config
	logger       *logger.Logger
}

// NewQuote will create a new Quote configuration by loading it.
func NewQuote(fs *afero.Fs, logger *logger.Logger) (Quotes, error) {
	if fs == nil || logger == nil {
		return nil, errors.New("nil file system or logger supplied")
	}

	return newQuotesImpl(fs, logger)
}

// newQuoteImpl will create a new quoteImpl configuration and load it from disk.
func newQuotesImpl(fs *afero.Fs, logger *logger.Logger) (q *quotesImpl, err error) {
	q = &quotesImpl{conf: newConfig(), logger: logger}
	if err = q.conf.Load(*fs); err != nil {
		q.logger.Error("failed to load Quote configurations from disk", zap.Error(err))

		return nil, err
	}

	// Fiat Client configuration.
	q.clientFiat, err = configFiatClient(q.conf)
	if err != nil {
		q.logger.Error("failed to configure Fiat client", zap.Error(err))

		return nil, err
	}

	// Crypto Client configuration.
	q.clientCrypto, err = configCryptoClient(q.conf)
	if err != nil {
		q.logger.Error("failed to configure Crypto client", zap.Error(err))

		return nil, err
	}

	return
}

// configFiatClient will setup the global configurations for the Fiat client.
func configFiatClient(conf *config) (*req.Client, error) {
	if conf == nil {
		return nil, fmt.Errorf("configurations not loaded")
	}

	return req.C().
			SetUserAgent(conf.Connection.UserAgent).
			SetTimeout(conf.Connection.Timeout).
			SetCommonHeader(conf.FiatCurrency.HeaderKey, conf.FiatCurrency.APIKey),
		nil
}

// configCryptoClient will setup the global configurations for the Crypto client.
func configCryptoClient(conf *config) (*req.Client, error) {
	if conf == nil {
		return nil, fmt.Errorf("configurations not loaded")
	}

	return req.C().
			SetUserAgent(conf.Connection.UserAgent).
			SetTimeout(conf.Connection.Timeout).
			SetCommonHeader(conf.CryptoCurrency.HeaderKey, conf.CryptoCurrency.APIKey),
		nil
}

// FiatQuote will access the Fiat currency price quote service and get the latest exchange rate.
func (q *quotesImpl) fiatQuote(source, destination string, sourceAmount decimal.Decimal) (models.FiatQuote, error) {
	result := models.FiatQuote{}

	_, err := q.clientFiat.R().
		SetQueryParam("from", source).
		SetQueryParam("to", destination).
		SetQueryParam("amount", sourceAmount.String()).
		SetSuccessResult(&result).
		Get(q.conf.FiatCurrency.Endpoint)

	// Failed to query endpoint for price.
	if err != nil {
		msg := "failed to get Fiat currency price quote"
		q.logger.Warn(msg, zap.Error(err))

		return result, NewError("please try again later").SetStatus(http.StatusServiceUnavailable)
	}

	// Check for a successful rate retrieval.
	if !result.Success {
		return result, NewError("invalid Fiat currency code").SetStatus(http.StatusBadRequest)
	}

	return result, nil
}

// FiatConversion will convert a source currency, in a given amount, to the destination currency.
func (q *quotesImpl) FiatConversion(
	source,
	destination string,
	amount decimal.Decimal,
	fiatQuote func(source, destination string, amount decimal.Decimal) (models.FiatQuote, error)) (
	decimal.Decimal, decimal.Decimal, error) {
	var (
		err      error
		rawQuote models.FiatQuote
	)

	// The fiatQuote parameter is exposed for stub injection used for testing.
	if fiatQuote == nil {
		fiatQuote = q.fiatQuote
	}

	rawQuote, err = fiatQuote(source, destination, amount)
	if err != nil {
		q.logger.Warn("failed to convert Fiat currency", zap.Error(err))

		return decimal.Decimal{}, decimal.Decimal{}, fmt.Errorf("%w", err)
	}

	// For precision related concerns the amount to be posted will be recalculated here. We only rely on the quote
	// provider for rate quote's precision and not the amount converted precision.
	convertedAmount := rawQuote.Info.Rate.
		Mul(amount).
		RoundBank(constants.GetDecimalPlacesFiat())

	return rawQuote.Info.Rate, convertedAmount, nil
}

// CryptoQuote will access the Fiat currency price quote service and get the latest exchange rate.
func (q *quotesImpl) cryptoQuote(source, destination string) (models.CryptoQuote, error) {
	result := models.CryptoQuote{}

	resp, err := q.clientCrypto.R().
		SetPathParam("base_symbol", source).
		SetPathParam("quote_symbol", destination).
		SetSuccessResult(&result).
		Get(q.conf.CryptoCurrency.Endpoint)

	// Failed to query endpoint for price.
	if err != nil {
		q.logger.Warn("failed to get Fiat currency price quote", zap.Error(err))

		return result, NewError("crypto price service unreachable").SetStatus(http.StatusInternalServerError)
	}

	if !resp.IsSuccessState() {
		// Invalid cryptocurrency codes.
		if resp.StatusCode == 550 { //nolint:gomnd
			return result, NewError("invalid Crypto currency code").SetStatus(http.StatusBadRequest)
		}

		// Log and other API related errors and return an internal server error to user.
		q.logger.Error("API error", zap.String("Response", resp.String()))

		return result, NewError("please try again later").SetStatus(http.StatusInternalServerError)
	}

	return result, nil
}
