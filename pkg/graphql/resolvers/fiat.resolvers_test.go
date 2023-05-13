package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/quotes"
)

func TestFiatResolver_OpenFiat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		path                 string
		query                string
		expectErr            bool
		authValidateJWTErr   error
		authValidateJWTTimes int
		fiatCreateAccErr     error
		fiatCreateAccTimes   int
	}{
		{
			name:                 "empty token",
			path:                 "/open-fiat/empty-token",
			query:                fmt.Sprintf(testFiatQuery["openFiat"], ""),
			expectErr:            true,
			authValidateJWTErr:   errors.New("invalid token"),
			authValidateJWTTimes: 1,
			fiatCreateAccErr:     nil,
			fiatCreateAccTimes:   0,
		}, {
			name:                 "invalid currency",
			path:                 "/open-fiat/invalid-currency",
			query:                fmt.Sprintf(testFiatQuery["openFiat"], "INVALID"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatCreateAccErr:     nil,
			fiatCreateAccTimes:   0,
		}, {
			name:                 "db failure",
			path:                 "/open-fiat/db-failure",
			query:                fmt.Sprintf(testFiatQuery["openFiat"], "CAD"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatCreateAccErr:     postgres.ErrNotFound,
			fiatCreateAccTimes:   1,
		}, {
			name:                 "valid",
			path:                 "/open-fiat/valid",
			query:                fmt.Sprintf(testFiatQuery["openFiat"], "USD"),
			expectErr:            false,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			fiatCreateAccErr:     nil,
			fiatCreateAccTimes:   1,
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
			mockRedis := mocks.NewMockRedis(mockCtrl)    // Not called.
			mockQuotes := quotes.NewMockQuotes(mockCtrl) // Not called.

			gomock.InOrder(
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(uuid.UUID{}, int64(-1), test.authValidateJWTErr).
					Times(test.authValidateJWTTimes),

				mockPostgres.EXPECT().FiatCreateAccount(gomock.Any(), gomock.Any()).
					Return(test.fiatCreateAccErr).
					Times(test.fiatCreateAccTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.Use(GinContextToContextMiddleware())
			router.POST(test.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockPostgres, mockQuotes, zapLogger))

			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path,
				bytes.NewBufferString(test.query))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "some valid auth token goes here")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// Verify responses
			require.Equal(t, http.StatusOK, recorder.Code, "expected status codes do not match")

			response := map[string]any{}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response), "failed to unmarshal response body")

			// Error is expected check to ensure one is set.
			if test.expectErr {
				verifyErrorReturned(t, response)
			}
		})
	}
}
