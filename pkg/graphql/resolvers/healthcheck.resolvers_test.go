package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/quotes"
	"github.com/surahman/FTeX/pkg/redis"
)

func TestQueryResolver_Healthcheck(t *testing.T) {
	t.Parallel()

	query := getHealthcheckQuery()

	testCases := []struct {
		name                string
		path                string
		expectedMsg         string
		expectErr           bool
		postgresHealthErr   error
		postgresHealthTimes int
		redisHealthErr      error
		redisHealthTimes    int
	}{
		{
			name:                "postgres failure",
			path:                "/healthcheck/postgres-failure",
			expectedMsg:         "postgres",
			expectErr:           true,
			postgresHealthErr:   postgres.ErrUnhealthy,
			postgresHealthTimes: 1,
			redisHealthErr:      nil,
			redisHealthTimes:    0,
		}, {
			name:                "redis failure",
			path:                "/healthcheck/redis-failure",
			expectedMsg:         "Redis",
			expectErr:           true,
			postgresHealthErr:   nil,
			postgresHealthTimes: 1,
			redisHealthErr:      redis.ErrUnhealthy,
			redisHealthTimes:    1,
		}, {
			name:                "success",
			path:                "/healthcheck/success",
			expectedMsg:         "healthy",
			expectErr:           false,
			postgresHealthErr:   nil,
			postgresHealthTimes: 1,
			redisHealthErr:      nil,
			redisHealthTimes:    1,
		},
	}
	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockAuth := mocks.NewMockAuth(mockCtrl) // Not called.
			mockPostgres := mocks.NewMockPostgres(mockCtrl)
			mockRedis := mocks.NewMockRedis(mockCtrl)
			mockQuotes := quotes.NewMockQuotes(mockCtrl) // Not called.

			gomock.InOrder(
				mockPostgres.EXPECT().Healthcheck().
					Return(test.postgresHealthErr).
					Times(test.postgresHealthTimes),

				mockRedis.EXPECT().Healthcheck().
					Return(test.redisHealthErr).
					Times(test.redisHealthTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.POST(test.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockPostgres, mockQuotes, zapLogger))

			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path, bytes.NewBufferString(query))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// Verify responses
			require.Equal(t, http.StatusOK, recorder.Code, "expected status codes do not match.")

			response := map[string]any{}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response), "failed to unmarshal response body.")

			// Error is expected check to ensure one is set.
			if test.expectErr {
				verifyErrorReturned(t, response)
			} else {
				// Auth token is expected.
				data, ok := response["data"]
				require.True(t, ok, "data key expected but not set")

				responseMessage, ok := data.(map[string]any)["healthcheck"].(string)
				require.True(t, ok, "response message not found.")
				require.Equal(t, "OK", responseMessage, "healthcheck did not return OK status.")
			}
		})
	}
}
