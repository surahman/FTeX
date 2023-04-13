package quotes

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/imroc/req/v3"
	"github.com/shopspring/decimal"
	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/models"
	"go.uber.org/zap"
)

// Mock Quotes interface stub generation.
//go:generate mockgen -destination=../mocks/mock_quotes.go -package=mocks github.com/surahman/FTeX/pkg/quotes Quotes

// Quotes is the interface through which the currency quote services can be accessed. Created to support mock testing.
type Quotes interface {
	// FiatQuote will retrieve a quote for a fiat currency price.
	FiatQuote(source, destination string, sourceAmount decimal.Decimal) (models.FiatQuote, error)
	CryptoQuote(source, destination string) (models.CryptoQuote, error)
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
			SetCommonHeader("X-RapidAPI-Key", conf.FiatCurrency.APIKey),
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
			SetCommonHeader("X-CoinAPI-Key", conf.CryptoCurrency.APIKey),
		nil
}

// FiatQuote will access the Fiat currency price quote service and get the latest exchange rate.
func (q *quotesImpl) FiatQuote(source, destination string, sourceAmount decimal.Decimal) (models.FiatQuote, error) {
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

		return result, fmt.Errorf(msg+" %w", err)
	}

	// Check for a successful rate retrieval.
	if !result.Success {
		return result, fmt.Errorf("invalid Fiat currency code")
	}

	return result, nil
}

// CryptoQuote will access the Fiat currency price quote service and get the latest exchange rate.
func (q *quotesImpl) CryptoQuote(source, destination string) (models.CryptoQuote, error) {
	result := models.CryptoQuote{}

	resp, err := q.clientCrypto.R().
		SetPathParam("base_symbol", source).
		SetPathParam("quote_symbol", destination).
		SetSuccessResult(&result).
		Get(q.conf.CryptoCurrency.Endpoint)

	// Failed to query endpoint for price.
	if err != nil {
		msg := "failed to get Fiat currency price quote"
		q.logger.Warn(msg, zap.Error(err))

		return result, fmt.Errorf(msg+" %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("invalid Crypto currency code")
	}

	return result, nil
}
