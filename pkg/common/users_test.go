package common

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/models"
	modelsPostgres "github.com/surahman/FTeX/pkg/models/postgres"
	"github.com/surahman/FTeX/pkg/postgres"
)

func TestCommon_HTTPUserRegister(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		expectedMsg     string
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
		expectErr       require.ErrorAssertionFunc
		expectPayload   require.ValueAssertionFunc
		expectResponse  require.ValueAssertionFunc
	}{
		{
			name:            "empty user",
			expectedMsg:     "validation",
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
			expectErr:       require.Error,
			expectPayload:   require.NotNil,
			expectResponse:  require.Nil,
		}, {
			name:            "valid user",
			expectedMsg:     "",
			expectedStatus:  0,
			user:            testUserData["username1"],
			authHashPass:    "hashed password",
			authHashErr:     nil,
			authHashTimes:   1,
			createUserErr:   nil,
			createUserTimes: 1,
			authGenJWTToken: &models.JWTAuthResponse{},
			authGenJWTErr:   nil,
			authGenJWTTimes: 1,
			expectErr:       require.NoError,
			expectPayload:   require.Nil,
			expectResponse:  require.NotNil,
		}, {
			name:            "password hash failure",
			expectedMsg:     "retry",
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
			expectErr:       require.Error,
			expectPayload:   require.Nil,
			expectResponse:  require.Nil,
		}, {
			name:            "database failure",
			expectedMsg:     "",
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
			expectErr:       require.Error,
			expectPayload:   require.Nil,
			expectResponse:  require.Nil,
		}, {
			name:            "auth token failure",
			expectedMsg:     "auth token failure",
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
			expectErr:       require.Error,
			expectPayload:   require.Nil,
			expectResponse:  require.Nil,
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

			response, httpMsg, httpCode, payload, err := HTTPRegisterUser(mockAuth, mockPostgres, zapLogger, test.user)
			test.expectErr(t, err, "error expectation failed.")
			test.expectPayload(t, payload, "payload expectation failed.")
			test.expectResponse(t, response, "response expectation failed.")
			require.Equal(t, test.expectedStatus, httpCode, "http codes mismatched.")
			require.Contains(t, httpMsg, test.expectedMsg, "http message mismatched.")
		})
	}
}

func TestHandlers_UserLogin(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		expectedMsg        string
		expectedStatus     int
		user               *modelsPostgres.UserLoginCredentials
		userCredsErr       error
		userCredsTimes     int
		authCheckPassErr   error
		authCheckPassTimes int
		authGenJWTErr      error
		authGenJWTTimes    int
		expectErr          require.ErrorAssertionFunc
		expectPayload      require.ValueAssertionFunc
		expectToken        require.ValueAssertionFunc
	}{
		{
			name:               "empty user",
			expectedMsg:        constants.ValidationString(),
			expectedStatus:     http.StatusBadRequest,
			user:               &modelsPostgres.UserLoginCredentials{},
			userCredsErr:       nil,
			userCredsTimes:     0,
			authCheckPassErr:   nil,
			authCheckPassTimes: 0,
			authGenJWTErr:      nil,
			authGenJWTTimes:    0,
			expectErr:          require.Error,
			expectPayload:      require.NotNil,
			expectToken:        require.Nil,
		}, {
			name:               "valid user",
			expectedMsg:        "",
			expectedStatus:     0,
			user:               &testUserData["username1"].UserLoginCredentials,
			userCredsErr:       nil,
			userCredsTimes:     1,
			authCheckPassErr:   nil,
			authCheckPassTimes: 1,
			authGenJWTErr:      nil,
			authGenJWTTimes:    1,
			expectErr:          require.NoError,
			expectPayload:      require.Nil,
			expectToken:        require.NotNil,
		}, {
			name:               "database failure",
			expectedMsg:        "invalid username or password",
			expectedStatus:     http.StatusForbidden,
			user:               &testUserData["username1"].UserLoginCredentials,
			userCredsErr:       errors.New("database failure"),
			userCredsTimes:     1,
			authCheckPassErr:   nil,
			authCheckPassTimes: 0,
			authGenJWTErr:      nil,
			authGenJWTTimes:    0,
			expectErr:          require.Error,
			expectPayload:      require.Nil,
			expectToken:        require.Nil,
		}, {
			name:               "password check failure",
			expectedMsg:        "invalid username or password",
			expectedStatus:     http.StatusForbidden,
			user:               &testUserData["username1"].UserLoginCredentials,
			userCredsErr:       nil,
			userCredsTimes:     1,
			authCheckPassErr:   errors.New("password hash failure"),
			authCheckPassTimes: 1,
			authGenJWTErr:      nil,
			authGenJWTTimes:    0,
			expectErr:          require.Error,
			expectPayload:      require.Nil,
			expectToken:        require.Nil,
		}, {
			name:               "auth token failure",
			expectedMsg:        "auth token failure",
			expectedStatus:     http.StatusInternalServerError,
			user:               &testUserData["username1"].UserLoginCredentials,
			authCheckPassErr:   nil,
			authCheckPassTimes: 1,
			userCredsErr:       nil,
			userCredsTimes:     1,
			authGenJWTErr:      errors.New("auth token failure"),
			authGenJWTTimes:    1,
			expectErr:          require.Error,
			expectPayload:      require.Nil,
			expectToken:        require.Nil,
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

			token, httpMsg, httpCode, payload, err := HTTPLoginUser(mockAuth, mockPostgres, zapLogger, test.user)
			test.expectErr(t, err, "error expectation failed.")
			test.expectPayload(t, payload, "payload expectation failed.")
			test.expectToken(t, token, "token expectation failed.")
			require.Equal(t, test.expectedStatus, httpCode, "http codes mismatched.")
			require.Contains(t, httpMsg, test.expectedMsg, "http message mismatched.")
		})
	}
}
