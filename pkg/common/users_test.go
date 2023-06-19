package common

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

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
		user            modelsPostgres.UserAccount
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
			user:            modelsPostgres.UserAccount{},
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
			user:            *testUserData["username1"],
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
			user:            *testUserData["username1"],
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
			user:            *testUserData["username1"],
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
			user:            *testUserData["username1"],
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

			response, httpMsg, httpCode, payload, err := HTTPRegisterUser(mockAuth, mockPostgres, zapLogger, &test.user)
			test.expectErr(t, err, "error expectation failed.")
			test.expectPayload(t, payload, "payload expectation failed.")
			test.expectResponse(t, response, "response expectation failed.")
			require.Equal(t, test.expectedStatus, httpCode, "http codes mismatched.")
			require.Contains(t, httpMsg, test.expectedMsg, "http message mismatched.")
		})
	}
}

func TestCommon_HTTPUserLogin(t *testing.T) {
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
			expectedMsg:        "invalid credentials",
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

func TestCommon_HTTPLoginRefresh(t *testing.T) {
	t.Parallel()

	notExpiringTime := time.Now().Add(2 * time.Minute).Unix()
	validTime := time.Now().Add(-30 * time.Second).Unix()

	testCases := []struct {
		name             string
		expectedMsg      string
		expectedStatus   int
		expiresAt        int64
		userGetInfoAcc   modelsPostgres.User
		userGetInfoErr   error
		userGetInfoTimes int
		authRefreshTimes int
		authGenJWTErr    error
		authGenJWTTimes  int
		expectErr        require.ErrorAssertionFunc
		expectToken      require.ValueAssertionFunc
	}{
		{
			name:             "valid token",
			expectedMsg:      "",
			expectedStatus:   0,
			expiresAt:        validTime,
			userGetInfoAcc:   modelsPostgres.User{IsDeleted: false},
			userGetInfoErr:   nil,
			userGetInfoTimes: 1,
			authRefreshTimes: 1,
			authGenJWTErr:    nil,
			authGenJWTTimes:  1,
			expectErr:        require.NoError,
			expectToken:      require.NotNil,
		}, {
			name:             "valid token not expiring",
			expectedMsg:      "still valid",
			expectedStatus:   http.StatusNotExtended,
			expiresAt:        notExpiringTime,
			userGetInfoAcc:   modelsPostgres.User{IsDeleted: false},
			userGetInfoErr:   nil,
			userGetInfoTimes: 1,
			authRefreshTimes: 2, // Called once in an error message.
			authGenJWTErr:    nil,
			authGenJWTTimes:  0,
			expectErr:        require.Error,
			expectToken:      require.Nil,
		}, {
			name:           "db failure",
			expectedMsg:    constants.RetryMessageString(),
			expectedStatus: http.StatusInternalServerError,
			userGetInfoAcc: modelsPostgres.User{
				UserAccount: &modelsPostgres.UserAccount{
					UserLoginCredentials: modelsPostgres.UserLoginCredentials{Username: "some username"},
				},
			},
			userGetInfoErr:   errors.New("db failure"),
			userGetInfoTimes: 1,
			authRefreshTimes: 0,
			authGenJWTErr:    nil,
			authGenJWTTimes:  0,
			expectErr:        require.Error,
			expectToken:      require.Nil,
		}, {
			name:           "deleted user",
			expectedMsg:    "invalid token",
			expectedStatus: http.StatusForbidden,
			expiresAt:      validTime,
			userGetInfoAcc: modelsPostgres.User{
				UserAccount: &modelsPostgres.UserAccount{
					UserLoginCredentials: modelsPostgres.UserLoginCredentials{Username: "some username"},
				},
				IsDeleted: true,
			},
			userGetInfoErr:   nil,
			userGetInfoTimes: 1,
			authRefreshTimes: 0,
			authGenJWTErr:    nil,
			authGenJWTTimes:  0,
			expectErr:        require.Error,
			expectToken:      require.Nil,
		}, {
			name:             "token generation failure",
			expectedMsg:      "failed to generate token",
			expectedStatus:   http.StatusInternalServerError,
			expiresAt:        validTime,
			userGetInfoAcc:   modelsPostgres.User{IsDeleted: false},
			userGetInfoErr:   nil,
			userGetInfoTimes: 1,
			authRefreshTimes: 1,
			authGenJWTErr:    errors.New("failed to generate token"),
			authGenJWTTimes:  1,
			expectErr:        require.Error,
			expectToken:      require.Nil,
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
				mockPostgres.EXPECT().UserGetInfo(gomock.Any()).
					Return(test.userGetInfoAcc, test.userGetInfoErr).
					Times(test.userGetInfoTimes),

				mockAuth.EXPECT().RefreshThreshold().
					Return(int64(60)).
					Times(test.authRefreshTimes),

				mockAuth.EXPECT().GenerateJWT(gomock.Any()).
					Return(&models.JWTAuthResponse{}, test.authGenJWTErr).
					Times(test.authGenJWTTimes),
			)

			token, httpMsg, httpCode, err := HTTPRefreshLogin(mockAuth, mockPostgres, zapLogger, uuid.UUID{}, test.expiresAt)
			test.expectErr(t, err, "error expectation failed.")
			test.expectToken(t, token, "token expectation failed.")
			require.Equal(t, test.expectedStatus, httpCode, "http codes mismatched.")
			require.Contains(t, httpMsg, test.expectedMsg, "http message mismatched.")
		})
	}
}

func TestCommon_DeleteUser(t *testing.T) {
	t.Parallel()

	userAccount := &modelsPostgres.UserAccount{
		UserLoginCredentials: modelsPostgres.UserLoginCredentials{
			Username: "username1",
			Password: "user-password-1",
		},
	}

	userValid := &modelsPostgres.User{
		UserAccount: userAccount,
		ClientID:    uuid.UUID{},
		IsDeleted:   false,
	}

	userDeleted := &modelsPostgres.User{
		UserAccount: userAccount,
		ClientID:    uuid.UUID{},
		IsDeleted:   true,
	}

	testCases := []struct {
		name              string
		expectedMsg       string
		expectedStatus    int
		deleteRequest     *models.HTTPDeleteUserRequest
		userGetInfoAcc    modelsPostgres.User
		userGetInfoErr    error
		userGetInfoTimes  int
		authCheckPwdErr   error
		authCheckPwdTimes int
		userDeleteErr     error
		userDeleteTimes   int
		expectErr         require.ErrorAssertionFunc
		expectPayload     require.ValueAssertionFunc
	}{
		{
			name:              "empty request",
			expectedMsg:       constants.ValidationString(),
			expectedStatus:    http.StatusBadRequest,
			deleteRequest:     &models.HTTPDeleteUserRequest{},
			userGetInfoAcc:    modelsPostgres.User{},
			userGetInfoErr:    nil,
			userGetInfoTimes:  0,
			authCheckPwdErr:   nil,
			authCheckPwdTimes: 0,
			userDeleteErr:     nil,
			userDeleteTimes:   0,
			expectErr:         require.Error,
			expectPayload:     require.NotNil,
		}, {
			name:           "valid",
			expectedMsg:    "",
			expectedStatus: 0,
			deleteRequest: &models.HTTPDeleteUserRequest{
				UserLoginCredentials: modelsPostgres.UserLoginCredentials{
					Username: "username1",
					Password: "password",
				},
				Confirmation: fmt.Sprintf(constants.DeleteUserAccountConfirmation(), "username1"),
			},
			userGetInfoAcc:    *userValid,
			userGetInfoErr:    nil,
			userGetInfoTimes:  1,
			authCheckPwdErr:   nil,
			authCheckPwdTimes: 1,
			userDeleteErr:     nil,
			userDeleteTimes:   1,
			expectErr:         require.NoError,
			expectPayload:     require.Nil,
		}, {
			name:           "token and request username mismatch",
			expectedMsg:    "invalid deletion request",
			expectedStatus: http.StatusForbidden,
			deleteRequest: &models.HTTPDeleteUserRequest{
				UserLoginCredentials: modelsPostgres.UserLoginCredentials{
					Username: "different username",
					Password: "password",
				},
				Confirmation: fmt.Sprintf(constants.DeleteUserAccountConfirmation(), "username1"),
			},
			userGetInfoAcc:    *userValid,
			userGetInfoErr:    nil,
			userGetInfoTimes:  1,
			authCheckPwdErr:   nil,
			authCheckPwdTimes: 0,
			userDeleteErr:     nil,
			userDeleteTimes:   0,
			expectErr:         require.Error,
			expectPayload:     require.Nil,
		}, {
			name:           "db read failure",
			expectedMsg:    constants.RetryMessageString(),
			expectedStatus: http.StatusInternalServerError,
			deleteRequest: &models.HTTPDeleteUserRequest{
				UserLoginCredentials: modelsPostgres.UserLoginCredentials{
					Username: "username1",
					Password: "password",
				},
				Confirmation: fmt.Sprintf(constants.DeleteUserAccountConfirmation(), "username1"),
			},
			userGetInfoAcc:    modelsPostgres.User{},
			userGetInfoErr:    errors.New("db read failure"),
			userGetInfoTimes:  1,
			authCheckPwdErr:   nil,
			authCheckPwdTimes: 0,
			userDeleteErr:     nil,
			userDeleteTimes:   0,
			expectErr:         require.Error,
			expectPayload:     require.Nil,
		}, {
			name:           "already deleted",
			expectedMsg:    "already deleted",
			expectedStatus: http.StatusForbidden,
			deleteRequest: &models.HTTPDeleteUserRequest{
				UserLoginCredentials: modelsPostgres.UserLoginCredentials{
					Username: "username1",
					Password: "password",
				},
				Confirmation: fmt.Sprintf(constants.DeleteUserAccountConfirmation(), "username1"),
			},
			userGetInfoAcc:    *userDeleted,
			userGetInfoErr:    nil,
			userGetInfoTimes:  1,
			authCheckPwdErr:   nil,
			authCheckPwdTimes: 0,
			userDeleteErr:     nil,
			userDeleteTimes:   0,
			expectErr:         require.Error,
			expectPayload:     require.Nil,
		}, {
			name:           "db delete failure",
			expectedMsg:    constants.RetryMessageString(),
			expectedStatus: http.StatusInternalServerError,
			deleteRequest: &models.HTTPDeleteUserRequest{
				UserLoginCredentials: modelsPostgres.UserLoginCredentials{
					Username: "username1",
					Password: "password",
				},
				Confirmation: fmt.Sprintf(constants.DeleteUserAccountConfirmation(), "username1"),
			},
			userGetInfoAcc:    *userValid,
			userGetInfoErr:    nil,
			userGetInfoTimes:  1,
			authCheckPwdErr:   nil,
			authCheckPwdTimes: 1,
			userDeleteErr:     errors.New("db delete failure"),
			userDeleteTimes:   1,
			expectErr:         require.Error,
			expectPayload:     require.Nil,
		}, {
			name:           "bad deletion confirmation",
			expectedMsg:    "incorrect or incomplete",
			expectedStatus: http.StatusBadRequest,
			deleteRequest: &models.HTTPDeleteUserRequest{
				UserLoginCredentials: modelsPostgres.UserLoginCredentials{
					Username: "username1",
					Password: "password",
				},
				Confirmation: fmt.Sprintf(constants.DeleteUserAccountConfirmation(), "incorrect and incomplete confirmation"),
			},
			userGetInfoAcc:    *userValid,
			userGetInfoErr:    nil,
			userGetInfoTimes:  1,
			authCheckPwdErr:   nil,
			authCheckPwdTimes: 0,
			userDeleteErr:     nil,
			userDeleteTimes:   0,
			expectErr:         require.Error,
			expectPayload:     require.Nil,
		}, {
			name:           "invalid password",
			expectedMsg:    "invalid user credentials",
			expectedStatus: http.StatusForbidden,
			deleteRequest: &models.HTTPDeleteUserRequest{
				UserLoginCredentials: modelsPostgres.UserLoginCredentials{
					Username: "username1",
					Password: "password",
				},
				Confirmation: fmt.Sprintf(constants.DeleteUserAccountConfirmation(), "username1"),
			},
			userGetInfoAcc:    *userValid,
			userGetInfoErr:    nil,
			userGetInfoTimes:  1,
			authCheckPwdErr:   errors.New("password check failed"),
			authCheckPwdTimes: 1,
			userDeleteErr:     nil,
			userDeleteTimes:   0,
			expectErr:         require.Error,
			expectPayload:     require.Nil,
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
				mockPostgres.EXPECT().UserGetInfo(gomock.Any()).
					Return(test.userGetInfoAcc, test.userGetInfoErr).
					Times(test.userGetInfoTimes),

				mockAuth.EXPECT().CheckPassword(gomock.Any(), gomock.Any()).
					Return(test.authCheckPwdErr).
					Times(test.authCheckPwdTimes),

				mockPostgres.EXPECT().UserDelete(gomock.Any()).
					Return(test.userDeleteErr).
					Times(test.userDeleteTimes),
			)

			httpMsg, httpCode, payload, err := HTTPDeleteUser(mockAuth, mockPostgres, zapLogger, uuid.UUID{}, test.deleteRequest)
			test.expectErr(t, err, "error expectation failed.")
			test.expectPayload(t, payload, "payload expectation failed.")
			require.Equal(t, test.expectedStatus, httpCode, "http codes mismatched.")
			require.Contains(t, httpMsg, test.expectedMsg, "http message mismatched.")
		})
	}
}
