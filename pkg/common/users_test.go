package common

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
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
