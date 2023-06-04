package utilities

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/quotes"
)

func TestUtilities_HTTPCryptoTxPaginated(t *testing.T) {
	t.Parallel()

	var (
		decryptedCursor = fmt.Sprintf("%s,%s,%d",
			fmt.Sprintf(constants.GetMonthFormatString(), 2023, 6, "-04:00"),
			fmt.Sprintf(constants.GetMonthFormatString(), 2023, 7, "-04:00"),
			10)
		journalEntries   = []postgres.CryptoJournal{{}, {}, {}, {}}
		paramsPageCursor = &HTTPPaginatedTxParams{PageCursorStr: "?pageCursor=page-cursor"}
	)

	testCases := []struct {
		name                   string
		path                   string
		ticker                 string
		expectedMsg            string
		expectedStatus         int
		journalEntries         []postgres.CryptoJournal
		params                 *HTTPPaginatedTxParams
		authDecryptCursorErr   error
		authDecryptCursorTimes int
		authEncryptCursorErr   error
		authEncryptCursorTimes int
		fiatTxPaginatedErr     error
		fiatTxPaginatedTimes   int
		expectErr              require.ErrorAssertionFunc
		expectNextPage         require.BoolAssertionFunc
	}{
		{
			name:                   "no cursor or params",
			path:                   "no-cursor-or-params/",
			ticker:                 "ETH",
			expectedMsg:            "missing required parameters",
			params:                 &HTTPPaginatedTxParams{PageCursorStr: ""},
			journalEntries:         journalEntries,
			expectedStatus:         http.StatusBadRequest,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 0,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 0,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   0,
			expectErr:              require.Error,
			expectNextPage:         require.False,
		}, {
			name:                   "db failure",
			path:                   "db-failure/",
			ticker:                 "ETH",
			expectedMsg:            "records not found",
			params:                 paramsPageCursor,
			journalEntries:         journalEntries,
			expectedStatus:         http.StatusNotFound,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 1,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     postgres.ErrNotFound,
			fiatTxPaginatedTimes:   1,
			expectErr:              require.Error,
			expectNextPage:         require.False,
		}, {
			name:                   "unknown db failure",
			path:                   "unknown-db-failure/",
			ticker:                 "ETH",
			expectedMsg:            "retry",
			params:                 paramsPageCursor,
			journalEntries:         journalEntries,
			expectedStatus:         http.StatusInternalServerError,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 1,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     errors.New("db failure"),
			fiatTxPaginatedTimes:   1,
			expectErr:              require.Error,
			expectNextPage:         require.False,
		}, {
			name:                   "no transactions",
			path:                   "no-transactions/",
			ticker:                 "ETH",
			expectedMsg:            "no transactions",
			params:                 paramsPageCursor,
			journalEntries:         []postgres.CryptoJournal{},
			expectedStatus:         http.StatusRequestedRangeNotSatisfiable,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 1,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   1,
			expectErr:              require.Error,
			expectNextPage:         require.False,
		}, {
			name:                   "valid with cursor",
			path:                   "valid-with-cursor/",
			ticker:                 "ETH",
			expectedMsg:            "",
			params:                 paramsPageCursor,
			journalEntries:         journalEntries,
			expectedStatus:         0,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 1,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   1,
			expectErr:              require.NoError,
			expectNextPage:         require.False,
		}, {
			name:        "valid with query",
			path:        "valid-with-query/",
			ticker:      "ETH",
			expectedMsg: "",
			params: &HTTPPaginatedTxParams{
				PageSizeStr: "3",
				TimezoneStr: "-04:00",
				MonthStr:    "6",
				YearStr:     "2023",
			},
			journalEntries:         journalEntries,
			expectedStatus:         0,
			authDecryptCursorErr:   nil,
			authDecryptCursorTimes: 0,
			authEncryptCursorErr:   nil,
			authEncryptCursorTimes: 1,
			fiatTxPaginatedErr:     nil,
			fiatTxPaginatedTimes:   1,
			expectErr:              require.NoError,
			expectNextPage:         require.True,
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockAuth := mocks.NewMockAuth(mockCtrl)
			mockDB := mocks.NewMockPostgres(mockCtrl)

			gomock.InOrder(
				mockAuth.EXPECT().DecryptFromString(gomock.Any()).
					Return([]byte(decryptedCursor), test.authDecryptCursorErr).
					Times(test.authDecryptCursorTimes),

				mockAuth.EXPECT().EncryptToString(gomock.Any()).
					Return("encrypted-cursor", test.authEncryptCursorErr).
					Times(test.authEncryptCursorTimes),

				mockDB.EXPECT().CryptoTransactionsPaginated(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any()).
					Return(test.journalEntries, test.fiatTxPaginatedErr).
					Times(test.fiatTxPaginatedTimes),
			)

			actual, httpStatus, httpMessage, err :=
				HTTPCryptoTXPaginated(mockAuth, mockDB, zapLogger, test.params, uuid.UUID{}, test.ticker)
			test.expectErr(t, err, "error expectation failed.")
			require.Equal(t, test.expectedStatus, httpStatus, "http status codes mismatched.")
			require.Contains(t, httpMessage, test.expectedMsg, "http message mismatched.")

			if err != nil {
				return
			}

			test.expectNextPage(t, len(actual.Links.NextPage) > 0, "next page link expectation failed.")
		})
	}
}

func TestUtilities_HTTPPrepareCryptoOffer(t *testing.T) { //nolint: maintidx
	t.Parallel()

	var (
		sourceAmount = decimal.NewFromFloat(23123.12)
		quotesRate   = decimal.NewFromFloat(23100)
	)

	testCases := []struct {
		name             string
		source           string
		destination      string
		expectErrMsg     string
		httpMessage      string
		httpStatus       int
		isPurchase       bool
		quotesAmount     decimal.Decimal
		quotesTimes      int
		quotesErr        error
		authEncryptTimes int
		authEncryptErr   error
		redisTimes       int
		redisErr         error
		expectErr        require.ErrorAssertionFunc
	}{
		{
			name:             "validate offer - purchase",
			source:           "INVALID",
			destination:      "BTC",
			expectErrMsg:     "INVALID",
			httpMessage:      constants.GetInvalidRequest(),
			httpStatus:       http.StatusBadRequest,
			isPurchase:       true,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      0,
			quotesErr:        nil,
			authEncryptTimes: 0,
			authEncryptErr:   nil,
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
		}, {
			name:             "crypto conversion - purchase",
			source:           "USD",
			destination:      "BTC",
			expectErrMsg:     "quote failure",
			httpMessage:      retryMessage,
			httpStatus:       http.StatusInternalServerError,
			isPurchase:       true,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      1,
			quotesErr:        errors.New("quote failure"),
			authEncryptTimes: 0,
			authEncryptErr:   nil,
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
		}, {
			name:             "zero amount - purchase",
			source:           "USD",
			destination:      "BTC",
			expectErrMsg:     "too small",
			httpMessage:      "too small",
			httpStatus:       http.StatusBadRequest,
			isPurchase:       true,
			quotesAmount:     decimal.NewFromFloat(0),
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 0,
			authEncryptErr:   nil,
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
		}, {
			name:             "encryption failure - purchase",
			source:           "USD",
			destination:      "BTC",
			expectErrMsg:     "failed to encrypt",
			httpMessage:      retryMessage,
			httpStatus:       http.StatusInternalServerError,
			isPurchase:       true,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 1,
			authEncryptErr:   errors.New("encryption failure"),
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
		}, {
			name:             "cache failure - purchase",
			source:           "USD",
			destination:      "BTC",
			expectErrMsg:     "failed to store",
			httpMessage:      retryMessage,
			httpStatus:       http.StatusInternalServerError,
			isPurchase:       true,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisTimes:       1,
			redisErr:         errors.New("cache failure"),
			expectErr:        require.Error,
		}, {
			name:             "valid - purchase",
			source:           "USD",
			destination:      "BTC",
			expectErrMsg:     "",
			httpMessage:      "",
			httpStatus:       0,
			isPurchase:       true,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisTimes:       1,
			redisErr:         nil,
			expectErr:        require.NoError,
		}, {
			name:             "validate offer - sale",
			source:           "BTC",
			destination:      "INVALID",
			expectErrMsg:     "INVALID",
			httpMessage:      constants.GetInvalidRequest(),
			httpStatus:       http.StatusBadRequest,
			isPurchase:       false,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      0,
			quotesErr:        nil,
			authEncryptTimes: 0,
			authEncryptErr:   nil,
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
		}, {
			name:             "crypto conversion - sale",
			source:           "BTC",
			destination:      "USD",
			expectErrMsg:     "quote failure",
			httpMessage:      retryMessage,
			httpStatus:       http.StatusInternalServerError,
			isPurchase:       false,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      1,
			quotesErr:        errors.New("quote failure"),
			authEncryptTimes: 0,
			authEncryptErr:   nil,
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
		}, {
			name:             "zero amount - sale",
			source:           "BTC",
			destination:      "USD",
			expectErrMsg:     "too small",
			httpMessage:      "too small",
			httpStatus:       http.StatusBadRequest,
			isPurchase:       false,
			quotesAmount:     decimal.NewFromFloat(0),
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 0,
			authEncryptErr:   nil,
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
		}, {
			name:             "encryption failure - sale",
			source:           "BTC",
			destination:      "USD",
			expectErrMsg:     "failed to encrypt",
			httpMessage:      retryMessage,
			httpStatus:       http.StatusInternalServerError,
			isPurchase:       false,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 1,
			authEncryptErr:   errors.New("encryption failure"),
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
		}, {
			name:             "cache failure - sale",
			source:           "BTC",
			destination:      "USD",
			expectErrMsg:     "failed to store",
			httpMessage:      retryMessage,
			httpStatus:       http.StatusInternalServerError,
			isPurchase:       false,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisTimes:       1,
			redisErr:         errors.New("cache failure"),
			expectErr:        require.Error,
		}, {
			name:             "valid - sell",
			source:           "BTC",
			destination:      "USD",
			expectErrMsg:     "",
			httpMessage:      "",
			httpStatus:       0,
			isPurchase:       false,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisTimes:       1,
			redisErr:         nil,
			expectErr:        require.NoError,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockAuth := mocks.NewMockAuth(mockCtrl)
			mockCache := mocks.NewMockRedis(mockCtrl)
			mockQuotes := quotes.NewMockQuotes(mockCtrl)

			gomock.InOrder(
				mockQuotes.EXPECT().CryptoConversion(
					test.source, test.destination, sourceAmount, test.isPurchase, nil).
					Return(quotesRate, test.quotesAmount, test.quotesErr).
					Times(test.quotesTimes),

				mockAuth.EXPECT().EncryptToString(gomock.Any()).
					Return("OFFER-ID", test.authEncryptErr).
					Times(test.authEncryptTimes),

				mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(test.redisErr).
					Times(test.redisTimes),
			)

			offer, status, msg, err := HTTPPrepareCryptoOffer(mockAuth, mockCache, zapLogger, mockQuotes,
				uuid.UUID{}, test.source, test.destination, sourceAmount, test.isPurchase)
			test.expectErr(t, err, "error expectation failed.")

			if err != nil {
				require.Contains(t, err.Error(), test.expectErrMsg, "error message is incorrect.")
				require.Contains(t, msg, test.httpMessage, "http error message mismatched.")
				require.Equal(t, test.httpStatus, status, "http status mismatched.")

				return
			}

			require.Equal(t, test.source, offer.SourceAcc, "source account mismatch.")
			require.Equal(t, test.destination, offer.DestinationAcc, "destination account mismatch.")
			require.Equal(t, sourceAmount, offer.DebitAmount, "debit amount mismatch.")
			require.Equal(t, quotesRate, offer.Rate, "offer rate mismatch.")
			require.Equal(t, test.quotesAmount, offer.Amount, "offer amount mismatch.")
		})
	}
}

func TestUtilities_HTTPExchangeCrypto(t *testing.T) { //nolint: maintidx
	t.Parallel()

	validClientID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate a valid uuid.")

	invalidClientID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate a valid uuid.")

	cryptoAmount := decimal.NewFromFloat(1234.56)
	fiatAmount := decimal.NewFromFloat(78910.11)

	validFiat := models.HTTPExchangeOfferResponse{
		PriceQuote: models.PriceQuote{
			ClientID:       validClientID,
			SourceAcc:      "BTC",
			DestinationAcc: "USD",
			Rate:           decimal.Decimal{},
			Amount:         decimal.Decimal{},
		},
		DebitAmount:      decimal.Decimal{},
		OfferID:          "OFFER-ID",
		Expires:          0,
		IsCryptoPurchase: false,
		IsCryptoSale:     false,
	}

	invalidSale := models.HTTPExchangeOfferResponse{
		PriceQuote: models.PriceQuote{
			ClientID:       validClientID,
			SourceAcc:      "USD",
			DestinationAcc: "BTC", // should be Fiat currency ticker.
			Rate:           decimal.Decimal{},
			Amount:         fiatAmount,
		},
		DebitAmount:      cryptoAmount,
		OfferID:          "OFFER-ID",
		Expires:          0,
		IsCryptoPurchase: false,
		IsCryptoSale:     true,
	}

	validSale := models.HTTPExchangeOfferResponse{
		PriceQuote: models.PriceQuote{
			ClientID:       validClientID,
			SourceAcc:      "BTC",
			DestinationAcc: "USD",
			Rate:           decimal.Decimal{},
			Amount:         fiatAmount,
		},
		DebitAmount:      cryptoAmount,
		OfferID:          "OFFER-ID",
		Expires:          0,
		IsCryptoPurchase: false,
		IsCryptoSale:     true,
	}

	validPurchase := models.HTTPExchangeOfferResponse{
		PriceQuote: models.PriceQuote{
			ClientID:       validClientID,
			SourceAcc:      "USD",
			DestinationAcc: "BTC",
			Rate:           decimal.Decimal{},
			Amount:         cryptoAmount,
		},
		DebitAmount:      fiatAmount,
		OfferID:          "OFFER-ID",
		Expires:          0,
		IsCryptoPurchase: true,
		IsCryptoSale:     false,
	}

	testCases := []struct {
		name             string
		expectErrMsg     string
		clientID         uuid.UUID
		httpStatus       int
		authEncryptTimes int
		authEncryptErr   error
		redisGetData     models.HTTPExchangeOfferResponse
		redisGetTimes    int
		redisGetErr      error
		redisDelTimes    int
		redisDelErr      error
		purchaseTimes    int
		purchaseErr      error
		sellTimes        int
		sellErr          error
		expectErr        require.ErrorAssertionFunc
	}{
		{
			name:             "valid - purchase",
			clientID:         validClientID,
			expectErrMsg:     "",
			httpStatus:       0,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisGetData:     validPurchase,
			redisGetTimes:    1,
			redisGetErr:      nil,
			redisDelTimes:    1,
			redisDelErr:      nil,
			purchaseTimes:    1,
			purchaseErr:      nil,
			sellTimes:        0,
			sellErr:          nil,
			expectErr:        require.NoError,
		}, {
			name:             "decrypt failure - sell",
			clientID:         validClientID,
			expectErrMsg:     "retry",
			httpStatus:       http.StatusInternalServerError,
			authEncryptTimes: 1,
			authEncryptErr:   errors.New("decrypt failure"),
			redisGetData:     validSale,
			redisGetTimes:    0,
			redisGetErr:      nil,
			redisDelTimes:    0,
			redisDelErr:      nil,
			purchaseTimes:    0,
			purchaseErr:      nil,
			sellTimes:        0,
			sellErr:          nil,
			expectErr:        require.Error,
		}, {
			name:             "cache get failure - sell",
			clientID:         validClientID,
			expectErrMsg:     "retry",
			httpStatus:       http.StatusInternalServerError,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisGetData:     validSale,
			redisGetTimes:    1,
			redisGetErr:      errors.New("cache get failure"),
			redisDelTimes:    0,
			redisDelErr:      nil,
			purchaseTimes:    0,
			purchaseErr:      nil,
			sellTimes:        0,
			sellErr:          nil,
			expectErr:        require.Error,
		}, {
			name:             "cache del failure - sell",
			clientID:         validClientID,
			expectErrMsg:     "retry",
			httpStatus:       http.StatusInternalServerError,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisGetData:     validSale,
			redisGetTimes:    1,
			redisGetErr:      nil,
			redisDelTimes:    1,
			redisDelErr:      errors.New("cache del failure"),
			purchaseTimes:    0,
			purchaseErr:      nil,
			sellTimes:        0,
			sellErr:          nil,
			expectErr:        require.Error,
		}, {
			name:             "cache del failure - sell",
			clientID:         validClientID,
			expectErrMsg:     "invalid",
			httpStatus:       http.StatusBadRequest,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisGetData:     validFiat,
			redisGetTimes:    1,
			redisGetErr:      nil,
			redisDelTimes:    1,
			redisDelErr:      nil,
			purchaseTimes:    0,
			purchaseErr:      nil,
			sellTimes:        0,
			sellErr:          nil,
			expectErr:        require.Error,
		}, {
			name:             "clientID mismatch - sell",
			clientID:         invalidClientID,
			expectErrMsg:     "retry",
			httpStatus:       http.StatusInternalServerError,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisGetData:     validSale,
			redisGetTimes:    1,
			redisGetErr:      nil,
			redisDelTimes:    1,
			redisDelErr:      nil,
			purchaseTimes:    0,
			purchaseErr:      nil,
			sellTimes:        0,
			sellErr:          nil,
			expectErr:        require.Error,
		}, {
			name:             "validation failure - sell",
			clientID:         validClientID,
			expectErrMsg:     "Fiat",
			httpStatus:       http.StatusBadRequest,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisGetData:     invalidSale,
			redisGetTimes:    1,
			redisGetErr:      nil,
			redisDelTimes:    1,
			redisDelErr:      nil,
			purchaseTimes:    0,
			purchaseErr:      nil,
			sellTimes:        0,
			sellErr:          nil,
			expectErr:        require.Error,
		}, {
			name:             "transaction failure - sell",
			clientID:         validClientID,
			expectErrMsg:     "sell failure",
			httpStatus:       http.StatusInternalServerError,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisGetData:     validSale,
			redisGetTimes:    1,
			redisGetErr:      nil,
			redisDelTimes:    1,
			redisDelErr:      nil,
			purchaseTimes:    0,
			purchaseErr:      nil,
			sellTimes:        1,
			sellErr:          errors.New("sell failure"),
			expectErr:        require.Error,
		}, {
			name:             "valid - sell",
			clientID:         validClientID,
			expectErrMsg:     "",
			httpStatus:       0,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisGetData:     validSale,
			redisGetTimes:    1,
			redisGetErr:      nil,
			redisDelTimes:    1,
			redisDelErr:      nil,
			purchaseTimes:    0,
			purchaseErr:      nil,
			sellTimes:        1,
			sellErr:          nil,
			expectErr:        require.NoError,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockAuth := mocks.NewMockAuth(mockCtrl)
			mockCache := mocks.NewMockRedis(mockCtrl)
			mockPostgres := mocks.NewMockPostgres(mockCtrl)

			gomock.InOrder(
				mockAuth.EXPECT().DecryptFromString(gomock.Any()).
					Return([]byte("OFFER-ID"), test.authEncryptErr).
					Times(test.authEncryptTimes),

				mockCache.EXPECT().Get(gomock.Any(), gomock.Any()).
					Return(test.redisGetErr).
					SetArg(1, test.redisGetData).
					Times(test.redisGetTimes),

				mockCache.EXPECT().Del(gomock.Any()).
					Return(test.redisDelErr).
					Times(test.redisDelTimes),

				mockPostgres.EXPECT().CryptoPurchase(
					gomock.Any(), postgres.CurrencyUSD, fiatAmount, "BTC", cryptoAmount).
					Return(&postgres.FiatJournal{}, &postgres.CryptoJournal{}, test.purchaseErr).
					Times(test.purchaseTimes),

				mockPostgres.EXPECT().CryptoSell(
					gomock.Any(), postgres.CurrencyUSD, fiatAmount, "BTC", cryptoAmount).
					Return(&postgres.FiatJournal{}, &postgres.CryptoJournal{}, test.sellErr).
					Times(test.sellTimes),
			)

			_, status, errMsg, err :=
				HTTPExchangeCrypto(mockAuth, mockCache, mockPostgres, zapLogger, test.clientID, "offer-id")
			test.expectErr(t, err, "error expectation failed.")

			require.Equal(t, test.httpStatus, status, "http status code mismatched.")
			require.Contains(t, errMsg, test.expectErrMsg, "http error message mismatched.")
		})
	}
}

func TestUtilities_CryptoBalancePaginatedRequest(t *testing.T) {
	t.Parallel()

	encBTC, err := testAuth.EncryptToString([]byte("BTC"))
	require.NoError(t, err, "failed to encrypt BTC currency.")

	encETH, err := testAuth.EncryptToString([]byte("ETH"))
	require.NoError(t, err, "failed to encrypt ETH currency.")

	encXRP, err := testAuth.EncryptToString([]byte("XRP"))
	require.NoError(t, err, "failed to encrypt XRP currency.")

	testCases := []struct {
		name            string
		encryptedTicker string
		limitStr        string
		expectTicker    string
		expectLimit     int32
	}{
		{
			name:            "empty currency",
			encryptedTicker: "",
			limitStr:        "5",
			expectTicker:    "BTC",
			expectLimit:     5,
		}, {
			name:            "BTC",
			encryptedTicker: encBTC,
			limitStr:        "5",
			expectTicker:    "BTC",
			expectLimit:     5,
		}, {
			name:            "ETH",
			encryptedTicker: encETH,
			limitStr:        "5",
			expectTicker:    "ETH",
			expectLimit:     5,
		}, {
			name:            "XRP",
			encryptedTicker: encXRP,
			limitStr:        "5",
			expectTicker:    "XRP",
			expectLimit:     5,
		}, {
			name:            "base bound limit",
			encryptedTicker: encXRP,
			limitStr:        "0",
			expectTicker:    "XRP",
			expectLimit:     10,
		}, {
			name:            "above base bound limit",
			encryptedTicker: encXRP,
			limitStr:        "999",
			expectTicker:    "XRP",
			expectLimit:     999,
		}, {
			name:            "empty request",
			encryptedTicker: "",
			limitStr:        "",
			expectTicker:    "BTC",
			expectLimit:     10,
		}, {
			name:            "empty currency",
			encryptedTicker: "",
			limitStr:        "999",
			expectTicker:    "BTC",
			expectLimit:     999,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actualTicker, actualLimit, err := cryptoBalancePaginatedRequest(testAuth, test.encryptedTicker, test.limitStr)
			require.NoError(t, err, "error returned from query unpacking")
			require.Equal(t, test.expectTicker, actualTicker, "tickers mismatched.")
			require.Equal(t, test.expectLimit, actualLimit, "request limit size mismatched.")
		})
	}
}

func TestUtilities_HTTPCryptoBalancePaginated(t *testing.T) {
	t.Parallel()

	var (
		pageCursor  = "some-page-cursor"
		fourRecords = []postgres.CryptoAccount{{}, {}, {}, {}}
	)

	testCases := []struct {
		name               string
		pageSize           string
		expectNextPageSize string
		expectErrMsg       string
		httpStatus         int
		expectedRecordsLen int
		expectedPageSize   int32
		decryptStringErr   error
		decryptStringTimes int
		balanceData        []postgres.CryptoAccount
		balanceErr         error
		balanceTimes       int
		encryptStringErr   error
		encryptStingTimes  int
		expectErr          require.ErrorAssertionFunc
		expectNextPage     require.BoolAssertionFunc
	}{
		{
			name:               "cursor bad page failure",
			pageSize:           "bad-page-size",
			expectNextPageSize: "pageSize=3",
			expectErrMsg:       "page size",
			httpStatus:         http.StatusBadRequest,
			expectedRecordsLen: 0,
			expectedPageSize:   3,
			decryptStringErr:   nil,
			decryptStringTimes: 1,
			balanceData:        fourRecords,
			balanceErr:         nil,
			balanceTimes:       0,
			encryptStringErr:   nil,
			encryptStingTimes:  0,
			expectErr:          require.Error,
			expectNextPage:     require.True,
		}, {
			name:               "cursor decryption failure",
			pageSize:           "3",
			expectNextPageSize: "pageSize=3",
			expectErrMsg:       "invalid page cursor",
			httpStatus:         http.StatusBadRequest,
			expectedRecordsLen: 0,
			expectedPageSize:   3,
			decryptStringErr:   errors.New("decrypt failure"),
			decryptStringTimes: 1,
			balanceData:        fourRecords,
			balanceErr:         nil,
			balanceTimes:       0,
			encryptStringErr:   nil,
			encryptStingTimes:  0,
			expectErr:          require.Error,
			expectNextPage:     require.True,
		}, {
			name:               "db failure - known error",
			pageSize:           "3",
			expectNextPageSize: "pageSize=3",
			expectErrMsg:       "not found",
			httpStatus:         http.StatusNotFound,
			expectedRecordsLen: 0,
			expectedPageSize:   3,
			decryptStringErr:   nil,
			decryptStringTimes: 1,
			balanceData:        fourRecords,
			balanceErr:         postgres.ErrNotFound,
			balanceTimes:       1,
			encryptStringErr:   nil,
			encryptStingTimes:  0,
			expectErr:          require.Error,
			expectNextPage:     require.True,
		}, {
			name:               "db failure - unknown error",
			pageSize:           "3",
			expectNextPageSize: "pageSize=3",
			expectErrMsg:       retryMessage,
			httpStatus:         http.StatusInternalServerError,
			expectedRecordsLen: 0,
			expectedPageSize:   3,
			decryptStringErr:   nil,
			decryptStringTimes: 1,
			balanceData:        fourRecords,
			balanceErr:         errors.New("unknown db failure"),
			balanceTimes:       1,
			encryptStringErr:   nil,
			encryptStingTimes:  0,
			expectErr:          require.Error,
			expectNextPage:     require.True,
		}, {
			name:               "valid - default page size",
			pageSize:           "0",
			expectNextPageSize: "",
			expectErrMsg:       "",
			httpStatus:         0,
			expectedRecordsLen: 4,
			expectedPageSize:   10,
			decryptStringErr:   nil,
			decryptStringTimes: 1,
			balanceData:        fourRecords,
			balanceErr:         nil,
			balanceTimes:       1,
			encryptStringErr:   nil,
			encryptStingTimes:  0,
			expectErr:          require.NoError,
			expectNextPage:     require.False,
		}, {
			name:               "valid - no next page",
			pageSize:           "4",
			expectNextPageSize: "",
			expectErrMsg:       "",
			httpStatus:         0,
			expectedRecordsLen: 3,
			expectedPageSize:   4,
			decryptStringErr:   nil,
			decryptStringTimes: 1,
			balanceData:        []postgres.CryptoAccount{{}, {}, {}},
			balanceErr:         nil,
			balanceTimes:       1,
			encryptStringErr:   nil,
			encryptStingTimes:  0,
			expectErr:          require.NoError,
			expectNextPage:     require.False,
		}, {
			name:               "valid - has next page",
			pageSize:           "3",
			expectNextPageSize: "pageSize=3",
			expectErrMsg:       "",
			httpStatus:         0,
			expectedRecordsLen: 3,
			expectedPageSize:   3,
			decryptStringErr:   nil,
			decryptStringTimes: 1,
			balanceData:        fourRecords,
			balanceErr:         nil,
			balanceTimes:       1,
			encryptStringErr:   nil,
			encryptStingTimes:  1,
			expectErr:          require.NoError,
			expectNextPage:     require.True,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockAuth := mocks.NewMockAuth(mockCtrl)
			mockPostgres := mocks.NewMockPostgres(mockCtrl)

			gomock.InOrder(
				mockAuth.EXPECT().DecryptFromString(pageCursor).
					Return([]byte("decrypted-string"), test.decryptStringErr).
					Times(test.decryptStringTimes),

				mockPostgres.EXPECT().CryptoBalancePaginated(gomock.Any(), gomock.Any(), test.expectedPageSize+1).
					Return(test.balanceData, test.balanceErr).
					Times(test.balanceTimes),

				mockAuth.EXPECT().EncryptToString(gomock.Any()).
					Return("next-page-cursor", test.encryptStringErr).
					Times(test.encryptStingTimes),
			)

			actualDetails, status, errMsg, err := HTTPCryptoBalancePaginated(mockAuth, mockPostgres, zapLogger,
				uuid.UUID{}, pageCursor, test.pageSize)
			test.expectErr(t, err, "error expectation failed.")

			require.Equal(t, test.httpStatus, status, "http status code mismatched.")
			require.Contains(t, errMsg, test.expectErrMsg, "http error message mismatched.")

			if err != nil {
				return
			}

			require.Equal(t, 0, len(actualDetails.Links.PageCursor), "page cursor set.")
			require.Contains(t, actualDetails.Links.NextPage, test.expectNextPageSize, "expected page size mismatch.")
			test.expectNextPage(t, len(actualDetails.Links.NextPage) > 0, "next page link not set.")
			require.Equal(t, test.expectedRecordsLen, len(actualDetails.AccountBalances),
				"number of returned records mismatched")
		})
	}
}
