package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/redis"
)

func TestHealthcheck(t *testing.T) {
	t.Parallel()

	router := getTestRouter()

	testCases := []struct {
		name                string
		path                string
		expectedMsg         string
		expectedStatus      int
		postgresHealthError error
		postgresHealthTimes int
		redisHealthError    error
		redisHealthTimes    int
	}{
		{
			name:                "postgres failure",
			path:                "/healthcheck/postgres-failure",
			expectedMsg:         "Postgres",
			expectedStatus:      http.StatusServiceUnavailable,
			postgresHealthError: postgres.NewError(""),
			postgresHealthTimes: 1,
			redisHealthError:    nil,
			redisHealthTimes:    0,
		}, {
			name:                "redis failure",
			path:                "/healthcheck/redis-failure",
			expectedMsg:         "Redis",
			expectedStatus:      http.StatusServiceUnavailable,
			postgresHealthError: nil,
			postgresHealthTimes: 1,
			redisHealthError:    redis.NewError(""),
			redisHealthTimes:    1,
		}, {
			name:                "success",
			path:                "/healthcheck/success",
			expectedMsg:         "healthy",
			expectedStatus:      http.StatusOK,
			postgresHealthError: nil,
			postgresHealthTimes: 1,
			redisHealthError:    nil,
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

			mockPostgres := mocks.NewMockPostgres(mockCtrl)
			mockRedis := mocks.NewMockRedis(mockCtrl)

			// Configure mock expectations.
			gomock.InOrder(
				mockPostgres.EXPECT().Healthcheck().
					Return(test.postgresHealthError).
					Times(test.postgresHealthTimes),

				mockRedis.EXPECT().Healthcheck().
					Return(test.redisHealthError).
					Times(test.redisHealthTimes),
			)

			// Endpoint setup for test.
			router.GET(test.path, Healthcheck(zapLogger, mockPostgres, mockRedis))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodGet, test.path, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, recorder.Code, "expected status codes do not match")

			response := models.HTTPSuccess{}
			require.NoError(t, json.NewDecoder(recorder.Body).Decode(&response), "failed to unmarshall response body.")
			require.Containsf(t, response.Message, test.expectedMsg, "got incorrect message %s", response.Message)
		})
	}
}
