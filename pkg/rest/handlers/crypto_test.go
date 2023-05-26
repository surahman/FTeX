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
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
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
