package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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

func TestHandlers_OpenCrypto(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		path                 string
		expectedStatus       int
		request              *models.HTTPOpenCurrencyAccountRequest
		authValidateJWTErr   error
		authValidateTimes    int
		cryptoCreateAccErr   error
		cryptoCreateAccTimes int
	}{
		{
			name:                 "valid",
			path:                 "/open/valid",
			expectedStatus:       http.StatusCreated,
			request:              &models.HTTPOpenCurrencyAccountRequest{Currency: "BTC"},
			authValidateJWTErr:   nil,
			authValidateTimes:    1,
			cryptoCreateAccErr:   nil,
			cryptoCreateAccTimes: 1,
		}, {
			name:                 "validation",
			path:                 "/open/validation",
			expectedStatus:       http.StatusBadRequest,
			request:              &models.HTTPOpenCurrencyAccountRequest{},
			authValidateJWTErr:   nil,
			authValidateTimes:    0,
			cryptoCreateAccErr:   nil,
			cryptoCreateAccTimes: 0,
		}, {
			name:                 "invalid jwt",
			path:                 "/open/invalid-jwt",
			expectedStatus:       http.StatusForbidden,
			request:              &models.HTTPOpenCurrencyAccountRequest{Currency: "BTC"},
			authValidateJWTErr:   errors.New("invalid jwt"),
			authValidateTimes:    1,
			cryptoCreateAccErr:   nil,
			cryptoCreateAccTimes: 0,
		}, {
			name:                 "db failure",
			path:                 "/open/db-failure",
			expectedStatus:       http.StatusConflict,
			request:              &models.HTTPOpenCurrencyAccountRequest{Currency: "ETH"},
			authValidateJWTErr:   nil,
			authValidateTimes:    1,
			cryptoCreateAccErr:   postgres.ErrCreateFiat,
			cryptoCreateAccTimes: 1,
		}, {
			name:                 "db failure unknown",
			path:                 "/open/db-failure-unknown",
			expectedStatus:       http.StatusInternalServerError,
			request:              &models.HTTPOpenCurrencyAccountRequest{Currency: "USDC"},
			authValidateJWTErr:   nil,
			authValidateTimes:    1,
			cryptoCreateAccErr:   errors.New("unknown server error"),
			cryptoCreateAccTimes: 1,
		},
	}

	//nolint:dupl
	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockAuth := mocks.NewMockAuth(mockCtrl)
			mockPostgres := mocks.NewMockPostgres(mockCtrl)

			openReqJSON, err := json.Marshal(&test.request)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

			gomock.InOrder(
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(uuid.UUID{}, int64(0), test.authValidateJWTErr).
					Times(test.authValidateTimes),

				mockPostgres.EXPECT().CryptoCreateAccount(gomock.Any(), gomock.Any()).
					Return(test.cryptoCreateAccErr).
					Times(test.cryptoCreateAccTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.POST(test.path, OpenCrypto(zapLogger, mockAuth, mockPostgres, "Authorization"))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path, bytes.NewBuffer(openReqJSON))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, w.Code, "expected status codes do not match")
		})
	}
}

func TestHandlers_OfferCrypto(t *testing.T) { //nolint:maintidx
	t.Parallel()

	var (
		amountValid           = decimal.NewFromFloat(999)
		amountInvalidDecimal  = decimal.NewFromFloat(999.999)
		amountInvalidNegative = decimal.NewFromFloat(-999)
		isPurchase            = true
		isSale                = false
	)

	testCases := []struct {
		name               string
		expectedMsg        string
		path               string
		expectedStatus     int
		request            *models.HTTPCryptoOfferRequest
		isPurchase         bool
		authValidateJWTErr error
		authValidateTimes  int
		quotesErr          error
		quotesAmount       decimal.Decimal
		quotesTimes        int
		authEncryptErr     error
		authEncryptTimes   int
		redisErr           error
		redisTimes         int
	}{
		{
			name:               "empty request",
			expectedMsg:        "validation",
			path:               "/purchase-offer-crypto/empty-request",
			expectedStatus:     http.StatusBadRequest,
			request:            &models.HTTPCryptoOfferRequest{},
			isPurchase:         isPurchase,
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			quotesErr:          nil,
			quotesAmount:       amountValid,
			quotesTimes:        0,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:           "invalid source currency",
			expectedMsg:    "Fiat currency",
			path:           "/purchase-offer-crypto/invalid-fiat-currency",
			expectedStatus: http.StatusBadRequest,
			request: &models.HTTPCryptoOfferRequest{
				HTTPExchangeOfferRequest: models.HTTPExchangeOfferRequest{
					SourceCurrency:      "INVALID",
					DestinationCurrency: "BTC",
					SourceAmount:        amountValid,
				},
				IsPurchase: &isPurchase,
			},
			isPurchase:         isPurchase,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesAmount:       amountValid,
			quotesTimes:        0,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:           "no purchase or sale flag",
			expectedMsg:    "validation",
			path:           "/purchase-offer-crypto/invalid-fiat-currency",
			expectedStatus: http.StatusBadRequest,
			request: &models.HTTPCryptoOfferRequest{
				HTTPExchangeOfferRequest: models.HTTPExchangeOfferRequest{
					SourceCurrency:      "USD",
					DestinationCurrency: "BTC",
					SourceAmount:        amountValid,
				},
			},
			isPurchase:         isPurchase,
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			quotesErr:          nil,
			quotesAmount:       amountValid,
			quotesTimes:        0,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:           "too many decimal places",
			expectedMsg:    "amount",
			path:           "/purchase-offer-crypto/too-many-decimal-places",
			expectedStatus: http.StatusBadRequest,
			request: &models.HTTPCryptoOfferRequest{
				HTTPExchangeOfferRequest: models.HTTPExchangeOfferRequest{
					SourceCurrency:      "USD",
					DestinationCurrency: "BTC",
					SourceAmount:        amountInvalidDecimal,
				},
				IsPurchase: &isPurchase,
			},
			isPurchase:         isPurchase,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesAmount:       amountValid,
			quotesTimes:        0,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:           "negative",
			expectedMsg:    "amount",
			path:           "/purchase-offer-crypto/negative",
			expectedStatus: http.StatusBadRequest,
			request: &models.HTTPCryptoOfferRequest{
				HTTPExchangeOfferRequest: models.HTTPExchangeOfferRequest{
					SourceCurrency:      "USD",
					DestinationCurrency: "BTC",
					SourceAmount:        amountInvalidNegative,
				},
				IsPurchase: &isPurchase,
			},
			isPurchase:         isPurchase,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesAmount:       amountValid,
			quotesTimes:        0,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:           "invalid jwt",
			expectedMsg:    "invalid jwt",
			path:           "/purchase-offer-crypto/invalid-jwt",
			expectedStatus: http.StatusForbidden,
			request: &models.HTTPCryptoOfferRequest{
				HTTPExchangeOfferRequest: models.HTTPExchangeOfferRequest{
					SourceCurrency:      "USD",
					DestinationCurrency: "BTC",
					SourceAmount:        amountValid,
				},
				IsPurchase: &isPurchase,
			},
			isPurchase:         isPurchase,
			authValidateJWTErr: errors.New("invalid jwt"),
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesAmount:       amountValid,
			quotesTimes:        0,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:           "crypto conversion rate error",
			expectedMsg:    "retry",
			path:           "/purchase-offer-crypto/crypto-rate-error",
			expectedStatus: http.StatusInternalServerError,
			request: &models.HTTPCryptoOfferRequest{
				HTTPExchangeOfferRequest: models.HTTPExchangeOfferRequest{
					SourceCurrency:      "USD",
					DestinationCurrency: "BTC",
					SourceAmount:        amountValid,
				},
				IsPurchase: &isPurchase,
			},
			isPurchase:         isPurchase,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          errors.New(""),
			quotesAmount:       amountValid,
			quotesTimes:        1,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:           "crypto conversion amount too small",
			expectedMsg:    "purchase/sale amount",
			path:           "/purchase-offer-crypto/crypto-amount-too-small",
			expectedStatus: http.StatusBadRequest,
			request: &models.HTTPCryptoOfferRequest{
				HTTPExchangeOfferRequest: models.HTTPExchangeOfferRequest{
					SourceCurrency:      "USD",
					DestinationCurrency: "BTC",
					SourceAmount:        amountValid,
				},
				IsPurchase: &isPurchase,
			},
			isPurchase:         isPurchase,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesAmount:       decimal.NewFromFloat(0),
			quotesTimes:        1,
			authEncryptErr:     nil,
			authEncryptTimes:   0,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:           "encryption error",
			expectedMsg:    "retry",
			path:           "/purchase-offer-crypto/encryption-error",
			expectedStatus: http.StatusInternalServerError,
			request: &models.HTTPCryptoOfferRequest{
				HTTPExchangeOfferRequest: models.HTTPExchangeOfferRequest{
					SourceCurrency:      "USD",
					DestinationCurrency: "BTC",
					SourceAmount:        amountValid,
				},
				IsPurchase: &isPurchase,
			},
			isPurchase:         isPurchase,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesAmount:       amountValid,
			quotesTimes:        1,
			authEncryptErr:     errors.New(""),
			authEncryptTimes:   1,
			redisErr:           nil,
			redisTimes:         0,
		}, {
			name:           "redis error",
			expectedMsg:    "retry",
			path:           "/purchase-offer-crypto/redis-error",
			expectedStatus: http.StatusInternalServerError,
			request: &models.HTTPCryptoOfferRequest{
				HTTPExchangeOfferRequest: models.HTTPExchangeOfferRequest{
					SourceCurrency:      "USD",
					DestinationCurrency: "BTC",
					SourceAmount:        amountValid,
				},
				IsPurchase: &isPurchase,
			},
			isPurchase:         isPurchase,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesAmount:       amountValid,
			quotesTimes:        1,
			authEncryptErr:     nil,
			authEncryptTimes:   1,
			redisErr:           errors.New(""),
			redisTimes:         1,
		}, {
			name:           "valid - purchase",
			expectedMsg:    "",
			path:           "/purchase-offer-crypto/valid",
			expectedStatus: http.StatusOK,
			request: &models.HTTPCryptoOfferRequest{
				HTTPExchangeOfferRequest: models.HTTPExchangeOfferRequest{
					SourceCurrency:      "USD",
					DestinationCurrency: "BTC",
					SourceAmount:        amountValid,
				},
				IsPurchase: &isPurchase,
			},
			isPurchase:         isPurchase,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesAmount:       amountValid,
			quotesTimes:        1,
			authEncryptErr:     nil,
			authEncryptTimes:   1,
			redisErr:           nil,
			redisTimes:         1,
		}, {
			name:           "valid - sale",
			expectedMsg:    "",
			path:           "/sale-offer-crypto/valid",
			expectedStatus: http.StatusOK,
			request: &models.HTTPCryptoOfferRequest{
				HTTPExchangeOfferRequest: models.HTTPExchangeOfferRequest{
					SourceCurrency:      "BTC",
					DestinationCurrency: "USD",
					SourceAmount:        amountValid,
				},
				IsPurchase: &isSale,
			},
			isPurchase:         isSale,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			quotesErr:          nil,
			quotesAmount:       amountValid,
			quotesTimes:        1,
			authEncryptErr:     nil,
			authEncryptTimes:   1,
			redisErr:           nil,
			redisTimes:         1,
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

			offerReqJSON, err := json.Marshal(&test.request)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

			gomock.InOrder(
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(uuid.UUID{}, int64(0), test.authValidateJWTErr).
					Times(test.authValidateTimes),

				mockQuotes.EXPECT().CryptoConversion(gomock.Any(), gomock.Any(), gomock.Any(), test.isPurchase, nil).
					Return(amountValid, test.quotesAmount, test.quotesErr).
					Times(test.quotesTimes),

				mockAuth.EXPECT().EncryptToString(gomock.Any()).
					Return("OFFER-ID", test.authEncryptErr).
					Times(test.authEncryptTimes),

				mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(test.redisErr).
					Times(test.redisTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.POST(test.path, OfferCrypto(zapLogger, mockAuth, mockCache, mockQuotes, "Authorization"))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path, bytes.NewBuffer(offerReqJSON))
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, recorder.Code, "expected status codes do not match")

			var resp map[string]interface{}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp), "failed to unpack response.")

			errorMessage, ok := resp["message"].(string)
			require.True(t, ok, "failed to extract response message.")

			// Check for invalid currency codes and amount.
			if errorMessage == constants.GetInvalidRequest() {
				payload, ok := resp["payload"].(string)
				require.True(t, ok, "failed to extract payload from response.")
				require.Contains(t, payload, test.expectedMsg)
			} else {
				require.Contains(t, errorMessage, test.expectedMsg, "incorrect response message.")
			}
		})
	}
}

func TestHandlers_ExchangeCrypto(t *testing.T) {
	t.Parallel()

	validClientID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate a valid uuid.")

	cryptoAmount := decimal.NewFromFloat(1234.56)
	fiatAmount := decimal.NewFromFloat(78910.11)

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
		name               string
		expectedMsg        string
		path               string
		expectedStatus     int
		request            *models.HTTPTransferRequest
		authValidateJWTErr error
		authValidateTimes  int
		authEncryptTimes   int
		authEncryptErr     error
		redisGetData       models.HTTPExchangeOfferResponse
		redisGetTimes      int
		redisDelTimes      int
		purchaseTimes      int
		sellTimes          int
		expectErr          require.ErrorAssertionFunc
	}{
		{
			name:               "invalid jwt",
			expectedMsg:        "invalid jwt",
			path:               "/exchange-crypto/invalid-jwt",
			expectedStatus:     http.StatusForbidden,
			request:            &models.HTTPTransferRequest{OfferID: "OFFER-ID"},
			authValidateTimes:  1,
			authValidateJWTErr: errors.New("invalid jwt"),
			authEncryptTimes:   0,
			authEncryptErr:     nil,
			redisGetData:       validPurchase,
			redisGetTimes:      0,
			redisDelTimes:      0,
			purchaseTimes:      0,
			sellTimes:          0,
			expectErr:          require.Error,
		}, {
			name:               "empty request",
			expectedMsg:        "validation",
			path:               "/exchange-crypto/empty-request",
			expectedStatus:     http.StatusBadRequest,
			request:            &models.HTTPTransferRequest{},
			authValidateTimes:  0,
			authValidateJWTErr: nil,
			authEncryptTimes:   0,
			authEncryptErr:     nil,
			redisGetData:       validPurchase,
			redisGetTimes:      0,
			redisDelTimes:      0,
			purchaseTimes:      0,
			sellTimes:          0,
			expectErr:          require.Error,
		}, {
			name:               "transaction failure",
			expectedMsg:        "retry",
			path:               "/exchange-crypto/transaction-failure",
			expectedStatus:     http.StatusInternalServerError,
			request:            &models.HTTPTransferRequest{OfferID: "OFFER-ID"},
			authValidateTimes:  1,
			authValidateJWTErr: nil,
			authEncryptTimes:   1,
			authEncryptErr:     errors.New("transaction failure"),
			redisGetData:       validPurchase,
			redisGetTimes:      0,
			redisDelTimes:      0,
			purchaseTimes:      0,
			sellTimes:          0,
			expectErr:          require.Error,
		}, {
			name:               "valid - purchase",
			expectedMsg:        "successful",
			path:               "/exchange-crypto/valid-purchase",
			expectedStatus:     http.StatusOK,
			request:            &models.HTTPTransferRequest{OfferID: "OFFER-ID"},
			authValidateTimes:  1,
			authValidateJWTErr: nil,
			authEncryptTimes:   1,
			authEncryptErr:     nil,
			redisGetData:       validPurchase,
			redisGetTimes:      1,
			redisDelTimes:      1,
			purchaseTimes:      1,
			sellTimes:          0,
			expectErr:          require.NoError,
		}, {
			name:               "valid - sale",
			expectedMsg:        "successful",
			path:               "/exchange-crypto/valid-sale",
			expectedStatus:     http.StatusOK,
			request:            &models.HTTPTransferRequest{OfferID: "OFFER-ID"},
			authValidateTimes:  1,
			authValidateJWTErr: nil,
			authEncryptTimes:   1,
			authEncryptErr:     nil,
			redisGetData:       validSale,
			redisGetTimes:      1,
			redisDelTimes:      1,
			purchaseTimes:      0,
			sellTimes:          1,
			expectErr:          require.NoError,
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
			mockDB := mocks.NewMockPostgres(mockCtrl)

			offerReqJSON, err := json.Marshal(&test.request)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

			gomock.InOrder(
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(validClientID, int64(0), test.authValidateJWTErr).
					Times(test.authValidateTimes),

				mockAuth.EXPECT().DecryptFromString(gomock.Any()).
					Return([]byte("OFFER-ID"), test.authEncryptErr).
					Times(test.authEncryptTimes),

				mockCache.EXPECT().Get(gomock.Any(), gomock.Any()).
					Return(nil).
					SetArg(1, test.redisGetData).
					Times(test.redisGetTimes),

				mockCache.EXPECT().Del(gomock.Any()).
					Return(nil).
					Times(test.redisDelTimes),

				mockDB.EXPECT().CryptoPurchase(
					gomock.Any(), postgres.CurrencyUSD, fiatAmount, "BTC", cryptoAmount).
					Return(&postgres.FiatJournal{}, &postgres.CryptoJournal{}, nil).
					Times(test.purchaseTimes),

				mockDB.EXPECT().CryptoSell(
					gomock.Any(), postgres.CurrencyUSD, fiatAmount, "BTC", cryptoAmount).
					Return(&postgres.FiatJournal{}, &postgres.CryptoJournal{}, nil).
					Times(test.sellTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.POST(test.path, ExchangeCrypto(zapLogger, mockAuth, mockCache, mockDB, "Authorization"))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path, bytes.NewBuffer(offerReqJSON))
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, recorder.Code, "expected status codes do not match")

			var resp map[string]interface{}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp), "failed to unpack response.")

			errorMessage, ok := resp["message"].(string)
			require.True(t, ok, "failed to extract response message.")

			// Check for invalid currency codes and amount.
			if errorMessage == constants.GetInvalidRequest() {
				payload, ok := resp["payload"].(string)
				require.True(t, ok, "failed to extract payload from response.")
				require.Contains(t, payload, test.expectedMsg)
			} else {
				require.Contains(t, errorMessage, test.expectedMsg, "incorrect response message.")
			}
		})
	}
}

func TestHandler_BalanceCurrencyCrypto(t *testing.T) { //nolint:dupl
	t.Parallel()

	const basePath = "/crypto/balance/currency/"

	testCases := []struct {
		name               string
		currency           string
		expectedMsg        string
		expectedStatus     int
		authValidateJWTErr error
		authValidateTimes  int
		cryptoBalanceErr   error
		cryptoBalanceTimes int
	}{
		{
			name:               "invalid currency",
			currency:           "INVALID",
			expectedMsg:        "invalid currency",
			expectedStatus:     http.StatusBadRequest,
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			cryptoBalanceErr:   nil,
			cryptoBalanceTimes: 0,
		}, {
			name:               "invalid JWT",
			currency:           "USDT",
			expectedMsg:        "invalid JWT",
			expectedStatus:     http.StatusForbidden,
			authValidateJWTErr: errors.New("invalid JWT"),
			authValidateTimes:  1,
			cryptoBalanceErr:   nil,
			cryptoBalanceTimes: 0,
		}, {
			name:               "unknown db error",
			currency:           "USDC",
			expectedMsg:        "retry",
			expectedStatus:     http.StatusInternalServerError,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			cryptoBalanceErr:   errors.New("unknown error"),
			cryptoBalanceTimes: 1,
		}, {
			name:               "known db error",
			currency:           "BTC",
			expectedMsg:        "records not found",
			expectedStatus:     http.StatusNotFound,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			cryptoBalanceErr:   postgres.ErrNotFound,
			cryptoBalanceTimes: 1,
		}, {
			name:               "valid",
			currency:           "BTC",
			expectedMsg:        "account balance",
			expectedStatus:     http.StatusOK,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			cryptoBalanceErr:   nil,
			cryptoBalanceTimes: 1,
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
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(uuid.UUID{}, int64(0), test.authValidateJWTErr).
					Times(test.authValidateTimes),

				mockDB.EXPECT().CryptoBalanceCurrency(gomock.Any(), gomock.Any()).
					Return(postgres.CryptoAccount{}, test.cryptoBalanceErr).
					Times(test.cryptoBalanceTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.GET(basePath+":ticker", BalanceCurrencyCrypto(zapLogger, mockAuth, mockDB, "Authorization"))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodGet, basePath+test.currency, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, recorder.Code, "expected status codes do not match")

			var resp map[string]interface{}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp), "failed to unpack response.")

			actualMessage, ok := resp["message"].(string)
			require.True(t, ok, "failed to extract response message.")
			require.Contains(t, actualMessage, test.expectedMsg, "response message mismatch.")
		})
	}
}

func TestHandler_TxDetailsCrypto(t *testing.T) {
	t.Parallel()

	const basePath = "/crypto/transaction/currency/"

	validTxID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate uuid.")

	testCases := []struct {
		name               string
		expectedMsg        string
		expectedStatus     int
		authValidateJWTErr error
		authValidateTimes  int
		getTxErr           error
		getTxTimes         int
	}{
		{
			name:               "invalid jwt",
			expectedMsg:        "invalid jwt",
			expectedStatus:     http.StatusForbidden,
			authValidateJWTErr: errors.New("invalid jwt"),
			authValidateTimes:  1,
			getTxErr:           nil,
			getTxTimes:         0,
		}, {
			name:               "db error",
			expectedMsg:        "could not retrieve transaction details",
			expectedStatus:     http.StatusInternalServerError,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			getTxErr:           postgres.ErrTransactCryptoDetails,
			getTxTimes:         1,
		}, {
			name:               "valid",
			expectedMsg:        "transaction details",
			expectedStatus:     http.StatusOK,
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			getTxErr:           nil,
			getTxTimes:         1,
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
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(uuid.UUID{}, int64(0), test.authValidateJWTErr).
					Times(test.authValidateTimes),

				mockDB.EXPECT().CryptoTxDetailsCurrency(gomock.Any(), gomock.Any()).
					Return([]postgres.CryptoJournal{{}}, test.getTxErr).
					Times(test.getTxTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.GET(basePath+":transactionID", TxDetailsCrypto(zapLogger, mockAuth, mockDB, "Authorization"))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodGet, basePath+validTxID.String(), nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, recorder.Code, "expected status codes do not match")

			var resp map[string]interface{}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp), "failed to unpack response.")

			actualMessage, ok := resp["message"].(string)
			require.True(t, ok, "failed to extract response message.")
			require.Contains(t, actualMessage, test.expectedMsg, "response message mismatch.")
		})
	}
}
