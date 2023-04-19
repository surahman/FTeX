package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/models"
	modelsPostgres "github.com/surahman/FTeX/pkg/models/postgres"
	"github.com/surahman/FTeX/pkg/postgres"
)

func TestHandlers_UserRegister(t *testing.T) {
	t.Parallel()

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
		test := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockAuth := mocks.NewMockAuth(mockCtrl)
			mockPostgres := mocks.NewMockPostgres(mockCtrl)

			user := test.user
			userJSON, err := json.Marshal(&user)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

			gomock.InOrder(
				mockAuth.EXPECT().HashPassword(gomock.Any()).
					Return(test.authHashPass, test.authHashErr).
					Times(test.authHashTimes),

				mockPostgres.EXPECT().UserRegister(gomock.Any()).
					Return(uuid.UUID{}, test.createUserErr).
					Times(test.createUserTimes),

				mockAuth.EXPECT().GenerateJWT(gomock.Any()).
					Return(test.authGenJWTToken, test.authGenJWTErr).
					Times(test.authGenJWTTimes),
			)

			// Endpoint setup for test.
			router.POST(test.path, RegisterUser(zapLogger, mockAuth, mockPostgres))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path, bytes.NewBuffer(userJSON))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, w.Code, "expected status codes do not match")
		})
	}
}

func TestHandlers_UserCredentials(t *testing.T) {
	t.Parallel()

	router := getTestRouter()

	testCases := []struct {
		name               string
		path               string
		expectedStatus     int
		user               *modelsPostgres.UserLoginCredentials
		userCredsErr       error
		userCredsTimes     int
		authCheckPassErr   error
		authCheckPassTimes int
		authGenJWTErr      error
		authGenJWTTimes    int
	}{
		{
			name:               "empty user",
			path:               "/login/empty-user",
			expectedStatus:     http.StatusBadRequest,
			user:               &modelsPostgres.UserLoginCredentials{},
			userCredsErr:       nil,
			userCredsTimes:     0,
			authCheckPassErr:   nil,
			authCheckPassTimes: 0,
			authGenJWTErr:      nil,
			authGenJWTTimes:    0,
		}, {
			name:               "valid user",
			path:               "/login/valid-user",
			expectedStatus:     http.StatusOK,
			user:               &testUserData["username1"].UserLoginCredentials,
			userCredsErr:       nil,
			userCredsTimes:     1,
			authCheckPassErr:   nil,
			authCheckPassTimes: 1,
			authGenJWTErr:      nil,
			authGenJWTTimes:    1,
		}, {
			name:               "database failure",
			path:               "/login/database-failure",
			expectedStatus:     http.StatusForbidden,
			user:               &testUserData["username1"].UserLoginCredentials,
			userCredsErr:       errors.New("database failure"),
			userCredsTimes:     1,
			authCheckPassErr:   nil,
			authCheckPassTimes: 0,
			authGenJWTErr:      nil,
			authGenJWTTimes:    0,
		}, {
			name:               "password check failure",
			path:               "/login/pwd-check-failure",
			expectedStatus:     http.StatusForbidden,
			user:               &testUserData["username1"].UserLoginCredentials,
			userCredsErr:       nil,
			userCredsTimes:     1,
			authCheckPassErr:   errors.New("password hash failure"),
			authCheckPassTimes: 1,
			authGenJWTErr:      nil,
			authGenJWTTimes:    0,
		}, {
			name:               "auth token failure",
			path:               "/login/auth-token-failure",
			expectedStatus:     http.StatusInternalServerError,
			user:               &testUserData["username1"].UserLoginCredentials,
			authCheckPassErr:   nil,
			authCheckPassTimes: 1,
			userCredsErr:       nil,
			userCredsTimes:     1,
			authGenJWTErr:      errors.New("auth token failure"),
			authGenJWTTimes:    1,
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

			user := test.user
			userJSON, err := json.Marshal(&user)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

			gomock.InOrder(
				mockPostgres.EXPECT().UserCredentials(gomock.Any()).
					Return(uuid.UUID{}, "hashed password", test.userCredsErr).
					Times(test.userCredsTimes),

				mockAuth.EXPECT().CheckPassword(gomock.Any(), gomock.Any()).
					Return(test.authCheckPassErr).
					Times(test.authCheckPassTimes),

				mockAuth.EXPECT().GenerateJWT(gomock.Any()).
					Return(&models.JWTAuthResponse{}, test.authGenJWTErr).
					Times(test.authGenJWTTimes),
			)

			// Endpoint setup for test.
			router.POST(test.path, LoginUser(zapLogger, mockAuth, mockPostgres))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path, bytes.NewBuffer(userJSON))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, w.Code, "expected status codes do not match")
		})
	}
}
func TestLoginRefresh(t *testing.T) {
	t.Parallel()

	router := getTestRouter()

	testCases := []struct {
		name                 string
		path                 string
		expectedStatus       int
		authValidateJWTErr   error
		authValidateJWTExp   int64
		authValidateJWTTimes int
		userGetInfoAcc       modelsPostgres.User
		userGetInfoErr       error
		userGetInfoTimes     int
		authRefreshThresh    int64
		authRefreshTimes     int
		authGenJWTErr        error
		authGenJWTTimes      int
	}{
		{
			name:                 "empty token",
			path:                 "/refresh/empty-token",
			expectedStatus:       http.StatusForbidden,
			authValidateJWTErr:   errors.New("invalid token"),
			authValidateJWTExp:   time.Now().Unix(),
			authValidateJWTTimes: 1,
			userGetInfoAcc:       modelsPostgres.User{},
			userGetInfoErr:       nil,
			userGetInfoTimes:     0,
			authRefreshThresh:    60,
			authRefreshTimes:     0,
			authGenJWTErr:        nil,
			authGenJWTTimes:      0,
		}, {
			name:                 "valid token",
			path:                 "/refresh/valid-token",
			expectedStatus:       http.StatusOK,
			authValidateJWTErr:   nil,
			authValidateJWTExp:   time.Now().Add(time.Duration(30) * time.Second).Unix(),
			authValidateJWTTimes: 1,
			userGetInfoAcc:       modelsPostgres.User{IsDeleted: false},
			userGetInfoErr:       nil,
			userGetInfoTimes:     1,
			authRefreshThresh:    60,
			authRefreshTimes:     1,
			authGenJWTErr:        nil,
			authGenJWTTimes:      1,
		}, {
			name:                 "valid token not expiring",
			path:                 "/refresh/valid-token-not-expiring",
			expectedStatus:       http.StatusNotExtended,
			authValidateJWTErr:   nil,
			authValidateJWTExp:   time.Now().Add(time.Duration(90) * time.Second).Unix(),
			authValidateJWTTimes: 1,
			userGetInfoAcc:       modelsPostgres.User{IsDeleted: false},
			userGetInfoErr:       nil,
			userGetInfoTimes:     1,
			authRefreshThresh:    60,
			authRefreshTimes:     2, // Called once in error message.
			authGenJWTErr:        nil,
			authGenJWTTimes:      0,
		}, {
			name:                 "invalid token",
			path:                 "/refresh/invalid-token",
			expectedStatus:       http.StatusForbidden,
			authValidateJWTErr:   errors.New("validate JWT failure"),
			authValidateJWTExp:   time.Now().Unix(),
			authValidateJWTTimes: 1,
			userGetInfoAcc:       modelsPostgres.User{IsDeleted: false},
			userGetInfoErr:       nil,
			userGetInfoTimes:     0,
			authRefreshThresh:    60,
			authRefreshTimes:     0,
			authGenJWTErr:        nil,
			authGenJWTTimes:      0,
		}, {
			name:                 "db failure",
			path:                 "/refresh/db-failure",
			expectedStatus:       http.StatusInternalServerError,
			authValidateJWTErr:   nil,
			authValidateJWTExp:   time.Now().Add(time.Duration(30) * time.Second).Unix(),
			authValidateJWTTimes: 1,
			userGetInfoAcc: modelsPostgres.User{
				UserAccount: &modelsPostgres.UserAccount{
					UserLoginCredentials: modelsPostgres.UserLoginCredentials{Username: "some username"},
				},
			},
			userGetInfoErr:    errors.New("db failure"),
			userGetInfoTimes:  1,
			authRefreshThresh: 60,
			authRefreshTimes:  0,
			authGenJWTErr:     nil,
			authGenJWTTimes:   0,
		}, {
			name:                 "deleted user",
			path:                 "/refresh/deleted-user",
			expectedStatus:       http.StatusForbidden,
			authValidateJWTErr:   nil,
			authValidateJWTExp:   time.Now().Unix(),
			authValidateJWTTimes: 1,
			userGetInfoAcc: modelsPostgres.User{
				UserAccount: &modelsPostgres.UserAccount{
					UserLoginCredentials: modelsPostgres.UserLoginCredentials{Username: "some username"},
				},
				IsDeleted: true,
			},
			userGetInfoErr:    nil,
			userGetInfoTimes:  1,
			authRefreshThresh: 60,
			authRefreshTimes:  0,
			authGenJWTErr:     nil,
			authGenJWTTimes:   0,
		}, {
			name:                 "token generation failure",
			path:                 "/refresh/token-generation-failure",
			expectedStatus:       http.StatusInternalServerError,
			authValidateJWTErr:   nil,
			authValidateJWTExp:   time.Now().Add(time.Duration(30) * time.Second).Unix(),
			authValidateJWTTimes: 1,
			userGetInfoAcc:       modelsPostgres.User{IsDeleted: false},
			userGetInfoErr:       nil,
			userGetInfoTimes:     1,
			authRefreshThresh:    60,
			authRefreshTimes:     1,
			authGenJWTErr:        errors.New("failed to generate token"),
			authGenJWTTimes:      1,
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
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(uuid.UUID{}, test.authValidateJWTExp, test.authValidateJWTErr).
					Times(test.authValidateJWTTimes),

				mockPostgres.EXPECT().UserGetInfo(gomock.Any()).
					Return(test.userGetInfoAcc, test.userGetInfoErr).
					Times(test.userGetInfoTimes),

				mockAuth.EXPECT().RefreshThreshold().
					Return(test.authRefreshThresh).
					Times(test.authRefreshTimes),

				mockAuth.EXPECT().GenerateJWT(gomock.Any()).
					Return(&models.JWTAuthResponse{}, test.authGenJWTErr).
					Times(test.authGenJWTTimes),
			)

			// Endpoint setup for test.
			router.POST(test.path, LoginRefresh(zapLogger, mockAuth, mockPostgres, "Authorization"))
			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, test.expectedStatus, w.Code, "expected status codes do not match")
		})
	}
}
