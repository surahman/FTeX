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
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/quotes"
)

func TestMutationResolver_RegisterUser(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		path              string
		user              string
		expectErr         bool
		authHashErr       error
		authHashTimes     int
		userRegisterErr   error
		userRegisterTimes int
		authGenJWTErr     error
		authGenJWTTimes   int
	}{
		{
			name:              "empty user",
			path:              "/register/empty-user",
			user:              fmt.Sprintf(testUserQuery["register"], "", "", "", "", ""),
			expectErr:         true,
			authHashErr:       nil,
			authHashTimes:     0,
			userRegisterErr:   nil,
			userRegisterTimes: 0,
			authGenJWTErr:     nil,
			authGenJWTTimes:   0,
		}, {
			name: "valid user",
			path: "/register/valid-user",
			user: fmt.Sprintf(testUserQuery["register"],
				"first name", "last name", "email@address.com", "username999", "password999"),
			expectErr:         false,
			authHashErr:       nil,
			authHashTimes:     1,
			userRegisterErr:   nil,
			userRegisterTimes: 1,
			authGenJWTErr:     nil,
			authGenJWTTimes:   1,
		}, {
			name: "password hash failure",
			path: "/register/pwd-hash-failure",
			user: fmt.Sprintf(testUserQuery["register"],
				"first name", "last name", "email@address.com", "username999", "password999"),
			expectErr:         true,
			authHashErr:       errors.New("password hash failure"),
			authHashTimes:     1,
			userRegisterErr:   nil,
			userRegisterTimes: 0,
			authGenJWTErr:     nil,
			authGenJWTTimes:   0,
		}, {
			name: "database failure",
			path: "/register/database-failure",
			user: fmt.Sprintf(testUserQuery["register"],
				"first name", "last name", "email@address.com", "username999", "password999"),
			expectErr:         true,
			authHashErr:       nil,
			authHashTimes:     1,
			userRegisterErr:   postgres.ErrRegisterUser,
			userRegisterTimes: 1,
			authGenJWTErr:     nil,
			authGenJWTTimes:   0,
		}, {
			name: "auth token failure",
			path: "/register/auth-token-failure",
			user: fmt.Sprintf(testUserQuery["register"],
				"first name", "last name", "email@address.com", "username999", "password999"),
			expectErr:         true,
			authHashErr:       nil,
			authHashTimes:     1,
			userRegisterErr:   nil,
			userRegisterTimes: 1,
			authGenJWTErr:     errors.New("auth token failure"),
			authGenJWTTimes:   1,
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
				mockAuth.EXPECT().HashPassword(gomock.Any()).
					Return("hashed-password", test.authHashErr).
					Times(test.authHashTimes),

				mockPostgres.EXPECT().UserRegister(gomock.Any()).
					Return(uuid.UUID{}, test.userRegisterErr).
					Times(test.userRegisterTimes),

				mockAuth.EXPECT().GenerateJWT(gomock.Any()).
					Return(&models.JWTAuthResponse{}, test.authGenJWTErr).
					Times(test.authGenJWTTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.POST(test.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockPostgres, mockQuotes, zapLogger))

			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path, bytes.NewBufferString(test.user))
			req.Header.Set("Content-Type", "application/json")
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
