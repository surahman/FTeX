package rest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/mocks"
)

func TestAuthMiddleware(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockAuth := mocks.NewMockAuth(mockCtrl)
	mockDB := mocks.NewMockPostgres(mockCtrl)

	handler := AuthMiddleware(mockAuth, mockDB, zapLogger, "Authorization")
	require.NotNil(t, handler)
}

func TestAuthMiddleware_Handler(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		path              string
		token             string
		authJWTUUID       uuid.UUID
		authJWTError      error
		authJWTExpiration int64
		authJWTTimes      int
		isDeletedError    error
		isDeletedTimes    int
		isDeletedValue    bool
		expectedStatus    int
	}{
		{
			name:              "no token",
			path:              "/auth-middleware/no-token",
			token:             "",
			expectedStatus:    http.StatusUnauthorized,
			authJWTUUID:       uuid.UUID{},
			authJWTExpiration: -1,
			authJWTError:      nil,
			authJWTTimes:      0,
			isDeletedError:    nil,
			isDeletedTimes:    0,
			isDeletedValue:    false,
		}, {
			name:              "invalid token",
			path:              "/auth-middleware/invalid-token",
			token:             "invalid-token",
			expectedStatus:    http.StatusForbidden,
			authJWTUUID:       uuid.UUID{},
			authJWTExpiration: -1,
			authJWTError:      errors.New("JWT validation failure"),
			authJWTTimes:      1,
			isDeletedError:    nil,
			isDeletedTimes:    0,
			isDeletedValue:    false,
		}, {
			name:              "db failure",
			path:              "/auth-middleware/db-failure",
			token:             "valid-token",
			expectedStatus:    http.StatusInternalServerError,
			authJWTUUID:       uuid.UUID{},
			authJWTExpiration: -1,
			authJWTError:      nil,
			authJWTTimes:      1,
			isDeletedError:    errors.New("db failure"),
			isDeletedTimes:    1,
			isDeletedValue:    false,
		}, {
			name:              "deleted account",
			path:              "/auth-middleware/deleted-account",
			token:             "valid-token",
			expectedStatus:    http.StatusForbidden,
			authJWTUUID:       uuid.UUID{},
			authJWTExpiration: -1,
			authJWTError:      nil,
			authJWTTimes:      1,
			isDeletedError:    nil,
			isDeletedTimes:    1,
			isDeletedValue:    true,
		}, {
			name:              "valid token",
			path:              "/auth-middleware/valid-token",
			token:             "valid-token",
			expectedStatus:    http.StatusOK,
			authJWTUUID:       uuid.UUID{},
			authJWTExpiration: -1,
			authJWTError:      nil,
			authJWTTimes:      1,
			isDeletedError:    nil,
			isDeletedTimes:    1,
			isDeletedValue:    false,
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
					Return(
						test.authJWTUUID,
						test.authJWTExpiration,
						test.authJWTError,
					).Times(test.authJWTTimes),

				mockDB.EXPECT().UserIsDeleted(gomock.Any()).
					Return(test.isDeletedValue, test.isDeletedError).
					Times(test.isDeletedTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.POST(test.path, AuthMiddleware(mockAuth, mockDB, zapLogger, "Authorization"))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path, nil)
			req.Header.Set("Authorization", test.token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, w.Code, "expected status codes do not match")
		})
	}
}
