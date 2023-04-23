package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
)

func TestHandlers_OpenFiat(t *testing.T) {
	t.Parallel()

	router := getTestRouter()

	testCases := []struct {
		name               string
		path               string
		expectedStatus     int
		request            *models.HTTPOpenCurrencyAccount
		authValidateJWTErr error
		authValidateTimes  int
		fiatCreateAccErr   error
		fiatCreateAccTimes int
	}{
		{
			name:               "valid",
			path:               "/open/valid",
			expectedStatus:     http.StatusCreated,
			request:            &models.HTTPOpenCurrencyAccount{Currency: "USD"},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			fiatCreateAccErr:   nil,
			fiatCreateAccTimes: 1,
		}, {
			name:               "validation",
			path:               "/open/validation",
			expectedStatus:     http.StatusBadRequest,
			request:            &models.HTTPOpenCurrencyAccount{},
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			fiatCreateAccErr:   nil,
			fiatCreateAccTimes: 0,
		}, {
			name:               "invalid currency",
			path:               "/open/invalid-currency",
			expectedStatus:     http.StatusBadRequest,
			request:            &models.HTTPOpenCurrencyAccount{Currency: "UVW"},
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			fiatCreateAccErr:   nil,
			fiatCreateAccTimes: 0,
		}, {
			name:               "invalid jwt",
			path:               "/open/invalid-jwt",
			expectedStatus:     http.StatusForbidden,
			request:            &models.HTTPOpenCurrencyAccount{Currency: "USD"},
			authValidateJWTErr: errors.New("invalid jwt"),
			authValidateTimes:  1,
			fiatCreateAccErr:   nil,
			fiatCreateAccTimes: 0,
		}, {
			name:               "db failure",
			path:               "/open/db-failure",
			expectedStatus:     http.StatusConflict,
			request:            &models.HTTPOpenCurrencyAccount{Currency: "USD"},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			fiatCreateAccErr:   postgres.ErrCreateFiat,
			fiatCreateAccTimes: 1,
		}, {
			name:               "db failure unknown",
			path:               "/open/db-failure-unknown",
			expectedStatus:     http.StatusInternalServerError,
			request:            &models.HTTPOpenCurrencyAccount{Currency: "USD"},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			fiatCreateAccErr:   errors.New("unknown server error"),
			fiatCreateAccTimes: 1,
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

			openReqJSON, err := json.Marshal(&test.request)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

			gomock.InOrder(
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(uuid.UUID{}, int64(0), test.authValidateJWTErr).
					Times(test.authValidateTimes),

				mockPostgres.EXPECT().FiatCreateAccount(gomock.Any(), gomock.Any()).
					Return(test.fiatCreateAccErr).
					Times(test.fiatCreateAccTimes),
			)

			// Endpoint setup for test.
			router.POST(test.path, OpenFiat(zapLogger, mockAuth, mockPostgres, "Authorization"))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path, bytes.NewBuffer(openReqJSON))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, w.Code, "expected status codes do not match")
		})
	}
}

func TestHandlers_DepositFiat(t *testing.T) {
	t.Parallel()

	router := getTestRouter()

	testCases := []struct {
		name               string
		expectedMsg        string
		path               string
		expectedStatus     int
		request            *models.HTTPDepositCurrency
		authValidateJWTErr error
		authValidateTimes  int
		extTransferErr     error
		extTransferTimes   int
	}{
		{
			name:               "empty request",
			expectedMsg:        "validation",
			path:               "/deposit/empty-request",
			expectedStatus:     http.StatusBadRequest,
			request:            &models.HTTPDepositCurrency{},
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			extTransferErr:     nil,
			extTransferTimes:   0,
		}, {
			name:               "invalid currency",
			expectedMsg:        "currency",
			path:               "/deposit/invalid-currency",
			expectedStatus:     http.StatusBadRequest,
			request:            &models.HTTPDepositCurrency{Currency: "INVALID", Amount: decimal.NewFromFloat(1)},
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			extTransferErr:     nil,
			extTransferTimes:   0,
		}, {
			name:               "too many decimal places",
			expectedMsg:        "amount",
			path:               "/deposit/too-many-decimal-places",
			expectedStatus:     http.StatusBadRequest,
			request:            &models.HTTPDepositCurrency{Currency: "USD", Amount: decimal.NewFromFloat(1.234)},
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			extTransferErr:     nil,
			extTransferTimes:   0,
		}, {
			name:               "negative",
			expectedMsg:        "amount",
			path:               "/deposit/negative",
			expectedStatus:     http.StatusBadRequest,
			request:            &models.HTTPDepositCurrency{Currency: "USD", Amount: decimal.NewFromFloat(-1)},
			authValidateJWTErr: nil,
			authValidateTimes:  0,
			extTransferErr:     nil,
			extTransferTimes:   0,
		}, {
			name:               "invalid jwt",
			expectedMsg:        "invalid jwt",
			path:               "/deposit/invalid-jwt",
			expectedStatus:     http.StatusForbidden,
			request:            &models.HTTPDepositCurrency{Currency: "USD", Amount: decimal.NewFromFloat(1337.89)},
			authValidateJWTErr: errors.New("invalid jwt"),
			authValidateTimes:  1,
			extTransferErr:     nil,
			extTransferTimes:   0,
		}, {
			name:               "unknown xfer error",
			expectedMsg:        "retry",
			path:               "/deposit/unknown-xfer-error",
			expectedStatus:     http.StatusInternalServerError,
			request:            &models.HTTPDepositCurrency{Currency: "USD", Amount: decimal.NewFromFloat(1337.89)},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			extTransferErr:     errors.New("unknown error"),
			extTransferTimes:   1,
		}, {
			name:               "xfer error",
			expectedMsg:        "could not complete",
			path:               "/deposit/xfer-error",
			expectedStatus:     http.StatusInternalServerError,
			request:            &models.HTTPDepositCurrency{Currency: "USD", Amount: decimal.NewFromFloat(1337.89)},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			extTransferErr:     postgres.ErrTransactFiat,
			extTransferTimes:   1,
		}, {
			name:               "valid",
			expectedMsg:        "successfully",
			path:               "/deposit/valid",
			expectedStatus:     http.StatusOK,
			request:            &models.HTTPDepositCurrency{Currency: "USD", Amount: decimal.NewFromFloat(1337.89)},
			authValidateJWTErr: nil,
			authValidateTimes:  1,
			extTransferErr:     nil,
			extTransferTimes:   1,
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

			depositReqJSON, err := json.Marshal(&test.request)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

			gomock.InOrder(
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(uuid.UUID{}, int64(0), test.authValidateJWTErr).
					Times(test.authValidateTimes),

				mockPostgres.EXPECT().FiatExternalTransfer(gomock.Any(), gomock.Any()).
					Return(&postgres.FiatAccountTransferResult{}, test.extTransferErr).
					Times(test.extTransferTimes),
			)

			// Endpoint setup for test.
			router.POST(test.path, DepositFiat(zapLogger, mockAuth, mockPostgres, "Authorization"))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path, bytes.NewBuffer(depositReqJSON))
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, recorder.Code, "expected status codes do not match")

			var resp map[string]interface{}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp), "failed to unpack success response.")

			errorMessage, ok := resp["message"].(string)
			require.True(t, ok, "failed to extract response message.")
			require.Contains(t, errorMessage, test.expectedMsg, "incorrect response message.")
		})
	}
}
