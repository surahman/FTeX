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
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/models"
	modelsPostgres "github.com/surahman/FTeX/pkg/models/postgres"
	"github.com/surahman/FTeX/pkg/postgres"
)

func TestRegisterUser(t *testing.T) {
	router := getTestRouter()

	testCases := []struct {
		name            string
		path            string
		expectedStatus  int
		user            *modelsPostgres.UserAccount
		authHashPass    string
		authHashErr     error
		authHashTimes   int
		authGenJWTToken *models.JWTAuthResponse
		authGenJWTErr   error
		authGenJWTTimes int
		createUserErr   error
		createUserTimes int
	}{
		{
			name:            "empty user",
			path:            "/register/empty-user",
			expectedStatus:  http.StatusBadRequest,
			user:            &modelsPostgres.UserAccount{},
			authHashPass:    "",
			authHashErr:     nil,
			authHashTimes:   0,
			createUserErr:   nil,
			createUserTimes: 0,
			authGenJWTToken: nil,
			authGenJWTErr:   nil,
			authGenJWTTimes: 0,
		}, {
			name:            "valid user",
			path:            "/register/valid-user",
			expectedStatus:  http.StatusOK,
			user:            testUserData["username1"],
			authHashPass:    "hashed password",
			authHashErr:     nil,
			authHashTimes:   1,
			createUserErr:   nil,
			createUserTimes: 1,
			authGenJWTToken: nil,
			authGenJWTErr:   nil,
			authGenJWTTimes: 1,
		}, {
			name:            "password hash failure",
			path:            "/register/pwd-hash-failure",
			expectedStatus:  http.StatusInternalServerError,
			user:            testUserData["username1"],
			authHashPass:    "",
			authHashErr:     errors.New("password hash failure"),
			authHashTimes:   1,
			createUserErr:   nil,
			createUserTimes: 0,
			authGenJWTToken: nil,
			authGenJWTErr:   nil,
			authGenJWTTimes: 0,
		}, {
			name:            "database failure",
			path:            "/register/database-failure",
			expectedStatus:  http.StatusConflict,
			user:            testUserData["username1"],
			authHashPass:    "hashed password",
			authHashErr:     nil,
			authHashTimes:   1,
			createUserErr:   &postgres.Error{Code: http.StatusConflict},
			createUserTimes: 1,
			authGenJWTToken: nil,
			authGenJWTErr:   nil,
			authGenJWTTimes: 0,
		}, {
			name:            "auth token failure",
			path:            "/register/auth-token-failure",
			expectedStatus:  http.StatusInternalServerError,
			user:            testUserData["username1"],
			authHashPass:    "hashed password",
			authHashErr:     nil,
			authHashTimes:   1,
			createUserErr:   nil,
			createUserTimes: 1,
			authGenJWTToken: nil,
			authGenJWTErr:   errors.New("auth token failure"),
			authGenJWTTimes: 1,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockAuth := mocks.NewMockAuth(mockCtrl)
			mockPostgres := mocks.NewMockPostgres(mockCtrl)

			user := testCase.user
			userJSON, err := json.Marshal(&user)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

			gomock.InOrder(
				mockAuth.EXPECT().HashPassword(gomock.Any()).
					Return(testCase.authHashPass, testCase.authHashErr).
					Times(testCase.authHashTimes),

				mockPostgres.EXPECT().CreateUser(gomock.Any()).
					Return(uuid.UUID{}, testCase.createUserErr).
					Times(testCase.createUserTimes),

				mockAuth.EXPECT().GenerateJWT(gomock.Any()).
					Return(testCase.authGenJWTToken, testCase.authGenJWTErr).
					Times(testCase.authGenJWTTimes),
			)

			// Endpoint setup for test.
			router.POST(testCase.path, RegisterUser(zapLogger, mockAuth, mockPostgres))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, testCase.path, bytes.NewBuffer(userJSON))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")
		})
	}
}
