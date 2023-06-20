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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/models"
	modelsPostgres "github.com/surahman/FTeX/pkg/models/postgres"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/quotes"
)

func TestUserResolver_RegisterUser(t *testing.T) {
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

func TestUserResolver_DeleteUser(t *testing.T) {
	t.Parallel()

	validUserAccount := modelsPostgres.User{
		UserAccount: &modelsPostgres.UserAccount{
			UserLoginCredentials: modelsPostgres.UserLoginCredentials{
				Username: "username1",
				Password: "password",
			},
			FirstName: "",
			LastName:  "",
			Email:     "",
		},
		ClientID:  uuid.UUID{},
		IsDeleted: false,
	}

	deletedUserAccount := modelsPostgres.User{
		UserAccount: &modelsPostgres.UserAccount{
			UserLoginCredentials: modelsPostgres.UserLoginCredentials{
				Username: "username1",
				Password: "password",
			},
			FirstName: "",
			LastName:  "",
			Email:     "",
		},
		ClientID:  uuid.UUID{},
		IsDeleted: true,
	}

	testCases := []struct {
		name                 string
		path                 string
		query                string
		expectErr            bool
		authValidateJWTErr   error
		authValidateJWTTimes int
		readUserErr          error
		readUserTimes        int
		readUserData         modelsPostgres.User
		authCheckPassErr     error
		authCheckPassTimes   int
		deleteUserErr        error
		deleteUserTimes      int
	}{
		{
			name:                 "empty request",
			path:                 "/delete/empty-request",
			query:                fmt.Sprintf(testUserQuery["delete"], "", "", ""),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			readUserErr:          nil,
			readUserData:         modelsPostgres.User{},
			readUserTimes:        0,
			authCheckPassErr:     nil,
			authCheckPassTimes:   0,
			deleteUserErr:        nil,
			deleteUserTimes:      0,
		}, {
			name:                 "valid token",
			path:                 "/delete/valid-request",
			query:                fmt.Sprintf(testUserQuery["delete"], "username1", "password", "username1"),
			expectErr:            false,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			readUserErr:          nil,
			readUserData:         validUserAccount,
			readUserTimes:        1,
			authCheckPassErr:     nil,
			authCheckPassTimes:   1,
			deleteUserErr:        nil,
			deleteUserTimes:      1,
		}, {
			name:                 "invalid token",
			path:                 "/delete/invalid-token-request",
			query:                fmt.Sprintf(testUserQuery["delete"], "username1", "password", "username1"),
			expectErr:            true,
			authValidateJWTErr:   errors.New("JWT failed authorization check"),
			authValidateJWTTimes: 1,
			readUserErr:          nil,
			readUserData:         modelsPostgres.User{},
			readUserTimes:        0,
			authCheckPassErr:     nil,
			authCheckPassTimes:   0,
			deleteUserErr:        nil,
			deleteUserTimes:      0,
		}, {
			name:                 "db read failure",
			path:                 "/delete/db-read-failure",
			query:                fmt.Sprintf(testUserQuery["delete"], "username1", "password", "username1"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			readUserErr:          postgres.ErrNotFound,
			readUserData:         modelsPostgres.User{},
			readUserTimes:        1,
			authCheckPassErr:     nil,
			authCheckPassTimes:   0,
			deleteUserErr:        nil,
			deleteUserTimes:      0,
		}, {
			name:                 "token and request username mismatch",
			path:                 "/delete/token-and-request-username-mismatch",
			query:                fmt.Sprintf(testUserQuery["delete"], "invalid username", "password", "username1"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			readUserErr:          nil,
			readUserData:         validUserAccount,
			readUserTimes:        1,
			authCheckPassErr:     nil,
			authCheckPassTimes:   0,
			deleteUserErr:        nil,
			deleteUserTimes:      0,
		}, {
			name:                 "invalid password",
			path:                 "/delete/valid-password",
			query:                fmt.Sprintf(testUserQuery["delete"], "username1", "incorrect password", "username1"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			readUserErr:          nil,
			readUserData:         validUserAccount,
			readUserTimes:        1,
			authCheckPassErr:     errors.New("invalid password"),
			authCheckPassTimes:   1,
			deleteUserErr:        nil,
			deleteUserTimes:      0,
		}, {
			name: "bad deletion confirmation",
			path: "/delete/bad-deletion-confirmation",
			query: fmt.Sprintf(testUserQuery["delete"],
				"username1", "password", "incorrect and incomplete confirmation"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			readUserErr:          nil,
			readUserData:         validUserAccount,
			readUserTimes:        1,
			authCheckPassErr:     nil,
			authCheckPassTimes:   0,
			deleteUserErr:        nil,
			deleteUserTimes:      0,
		}, {
			name:                 "already deleted",
			path:                 "/delete/already-deleted",
			query:                fmt.Sprintf(testUserQuery["delete"], "username1", "password", "username1"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			readUserErr:          nil,
			readUserData:         deletedUserAccount,
			readUserTimes:        1,
			authCheckPassErr:     nil,
			authCheckPassTimes:   0,
			deleteUserErr:        nil,
			deleteUserTimes:      0,
		}, {
			name:                 "db delete failure",
			path:                 "/delete/db-delete-failure",
			query:                fmt.Sprintf(testUserQuery["delete"], "username1", "password", "username1"),
			expectErr:            true,
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
			readUserErr:          nil,
			readUserData:         validUserAccount,
			readUserTimes:        1,
			authCheckPassErr:     nil,
			authCheckPassTimes:   1,
			deleteUserErr:        postgres.ErrNotFound,
			deleteUserTimes:      1,
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

			authToken := xid.New().String()

			gomock.InOrder(
				mockAuth.EXPECT().ValidateJWT(authToken).
					Return(uuid.UUID{}, int64(-1), test.authValidateJWTErr).
					Times(test.authValidateJWTTimes),

				mockPostgres.EXPECT().UserGetInfo(gomock.Any()).
					Return(test.readUserData, test.readUserErr).
					Times(test.readUserTimes),

				mockAuth.EXPECT().CheckPassword(gomock.Any(), gomock.Any()).
					Return(test.authCheckPassErr).
					Times(test.authCheckPassTimes),

				mockPostgres.EXPECT().UserDelete(gomock.Any()).
					Return(test.deleteUserErr).
					Times(test.deleteUserTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.Use(GinContextToContextMiddleware())
			router.POST(test.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockPostgres, mockQuotes, zapLogger))

			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path, bytes.NewBufferString(test.query))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", authToken)
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
				require.True(t, ok, "data key expected but not set.")
				confirmation, ok := data.(map[string]any)["deleteUser"].(string)
				require.True(t, ok, "failed to extract confirmation.")
				require.Equal(t, "account successfully deleted", confirmation,
					"confirmation message does not match expected.")
			}
		})
	}
}

func TestUserResolver_LoginUser(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                   string
		path                   string
		user                   string
		expectErr              bool
		userCredentialsReadErr error
		userCredentialsTimes   int
		authCheckPassErr       error
		authCheckPassTimes     int
		authGenJWTErr          error
		authGenJWTTimes        int
	}{
		{
			name:                   "empty user",
			path:                   "/login/empty-user",
			user:                   fmt.Sprintf(testUserQuery["login"], "", ""),
			expectErr:              true,
			userCredentialsReadErr: nil,
			userCredentialsTimes:   0,
			authCheckPassErr:       nil,
			authCheckPassTimes:     0,
			authGenJWTErr:          nil,
			authGenJWTTimes:        0,
		}, {
			name:                   "valid user",
			path:                   "/login/valid-user",
			user:                   fmt.Sprintf(testUserQuery["login"], "username999", "password999"),
			expectErr:              false,
			userCredentialsReadErr: nil,
			userCredentialsTimes:   1,
			authCheckPassErr:       nil,
			authCheckPassTimes:     1,
			authGenJWTErr:          nil,
			authGenJWTTimes:        1,
		},
		{
			name:                   "database failure",
			path:                   "/login/database-failure",
			user:                   fmt.Sprintf(testUserQuery["login"], "username999", "password999"),
			expectErr:              true,
			userCredentialsReadErr: postgres.ErrNotFound,
			userCredentialsTimes:   1,
			authCheckPassErr:       nil,
			authCheckPassTimes:     0,
			authGenJWTErr:          nil,
			authGenJWTTimes:        0,
		}, {
			name:                   "password check failure",
			path:                   "/login/pwd-check-failure",
			user:                   fmt.Sprintf(testUserQuery["login"], "username999", "password999"),
			expectErr:              true,
			userCredentialsReadErr: nil,
			userCredentialsTimes:   1,
			authCheckPassErr:       errors.New("password hash error"),
			authCheckPassTimes:     1,
			authGenJWTErr:          nil,
			authGenJWTTimes:        0,
		}, {
			name:                   "auth token failure",
			path:                   "/login/auth-token-failure",
			user:                   fmt.Sprintf(testUserQuery["login"], "username999", "password999"),
			expectErr:              true,
			userCredentialsReadErr: nil,
			userCredentialsTimes:   1,
			authCheckPassErr:       nil,
			authCheckPassTimes:     1,
			authGenJWTErr:          errors.New("auth token failure"),
			authGenJWTTimes:        1,
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
				mockPostgres.EXPECT().UserCredentials(gomock.Any()).
					Return(uuid.UUID{}, "hashed-password", test.userCredentialsReadErr).
					Times(test.userCredentialsTimes),

				mockAuth.EXPECT().CheckPassword(gomock.Any(), gomock.Any()).
					Return(test.authCheckPassErr).
					Times(test.authCheckPassTimes),

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

func TestUserResolver_RefreshToken(t *testing.T) {
	t.Parallel()

	validUserAccount := modelsPostgres.User{
		UserAccount: &modelsPostgres.UserAccount{
			UserLoginCredentials: modelsPostgres.UserLoginCredentials{
				Username: "username1",
				Password: "password",
			},
			FirstName: "",
			LastName:  "",
			Email:     "",
		},
		ClientID:  uuid.UUID{},
		IsDeleted: false,
	}

	deletedUserAccount := modelsPostgres.User{
		UserAccount: &modelsPostgres.UserAccount{
			UserLoginCredentials: modelsPostgres.UserLoginCredentials{
				Username: "username1",
				Password: "password",
			},
			FirstName: "",
			LastName:  "",
			Email:     "",
		},
		ClientID:  uuid.UUID{},
		IsDeleted: true,
	}

	testCases := []struct {
		name                  string
		path                  string
		expectErr             bool
		authValidateJWTErr    error
		authValidateJWTExp    int64
		authValidateJWTTimes  int
		userGetInfoErr        error
		userGetInfoData       modelsPostgres.User
		userGetInfoTimes      int
		authRefThresholdTimes int
		authRefThreshold      int64
		authGenJWTErr         error
		authGenJWTTimes       int
	}{
		{
			name:                  "empty token",
			path:                  "/refresh/empty-token",
			expectErr:             true,
			authValidateJWTErr:    errors.New("invalid token"),
			authValidateJWTExp:    0,
			authValidateJWTTimes:  1,
			userGetInfoErr:        nil,
			userGetInfoData:       modelsPostgres.User{},
			userGetInfoTimes:      0,
			authRefThresholdTimes: 0,
			authRefThreshold:      60,
			authGenJWTErr:         nil,
			authGenJWTTimes:       0,
		}, {
			name:                  "valid token",
			path:                  "/refresh/valid-token",
			expectErr:             false,
			authValidateJWTErr:    nil,
			authValidateJWTExp:    time.Now().Add(-time.Duration(30) * time.Second).Unix(),
			authValidateJWTTimes:  1,
			userGetInfoErr:        nil,
			userGetInfoData:       validUserAccount,
			userGetInfoTimes:      1,
			authRefThresholdTimes: 1,
			authRefThreshold:      60,
			authGenJWTErr:         nil,
			authGenJWTTimes:       1,
		}, {
			name:                  "valid token not expiring",
			path:                  "/refresh/valid-token-not-expiring",
			expectErr:             true,
			authValidateJWTErr:    nil,
			authValidateJWTExp:    time.Now().Add(time.Duration(3) * time.Minute).Unix(),
			authValidateJWTTimes:  1,
			userGetInfoErr:        nil,
			userGetInfoData:       validUserAccount,
			userGetInfoTimes:      1,
			authRefThresholdTimes: 2,
			authRefThreshold:      60,
			authGenJWTErr:         nil,
			authGenJWTTimes:       0,
		}, {
			name:                  "invalid token",
			path:                  "/refresh/invalid-token",
			expectErr:             true,
			authValidateJWTErr:    errors.New("validate JWT failure"),
			authValidateJWTExp:    time.Now().Add(time.Duration(30) * time.Second).Unix(),
			authValidateJWTTimes:  1,
			userGetInfoErr:        nil,
			userGetInfoData:       validUserAccount,
			userGetInfoTimes:      0,
			authRefThresholdTimes: 0,
			authRefThreshold:      60,
			authGenJWTErr:         nil,
			authGenJWTTimes:       0,
		}, {
			name:                  "db failure",
			path:                  "/refresh/db-failure",
			expectErr:             true,
			authValidateJWTErr:    nil,
			authValidateJWTExp:    time.Now().Add(time.Duration(30) * time.Second).Unix(),
			authValidateJWTTimes:  1,
			userGetInfoErr:        postgres.ErrNotFound,
			userGetInfoData:       modelsPostgres.User{},
			userGetInfoTimes:      1,
			authRefThresholdTimes: 0,
			authRefThreshold:      60,
			authGenJWTErr:         nil,
			authGenJWTTimes:       0,
		}, {
			name:                  "deleted user",
			path:                  "/refresh/deleted-user",
			expectErr:             true,
			authValidateJWTErr:    nil,
			authValidateJWTExp:    time.Now().Add(time.Duration(30) * time.Second).Unix(),
			authValidateJWTTimes:  1,
			userGetInfoErr:        nil,
			userGetInfoData:       deletedUserAccount,
			userGetInfoTimes:      1,
			authRefThresholdTimes: 0,
			authRefThreshold:      60,
			authGenJWTErr:         nil,
			authGenJWTTimes:       0,
		}, {
			name:                  "token generation failure",
			path:                  "/refresh/token-generation-failure",
			expectErr:             true,
			authValidateJWTErr:    nil,
			authValidateJWTExp:    time.Now().Add(time.Duration(30) * time.Second).Unix(),
			authValidateJWTTimes:  1,
			userGetInfoErr:        nil,
			userGetInfoData:       validUserAccount,
			userGetInfoTimes:      1,
			authRefThresholdTimes: 1,
			authRefThreshold:      60,
			authGenJWTErr:         errors.New("failed to generate token"),
			authGenJWTTimes:       1,
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
				// JWT check.
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).
					Return(uuid.UUID{}, test.authValidateJWTExp, test.authValidateJWTErr).
					Times(test.authValidateJWTTimes),

				// Database call for user record.
				mockPostgres.EXPECT().UserGetInfo(gomock.Any()).
					Return(test.userGetInfoData, test.userGetInfoErr).
					Times(test.userGetInfoTimes),

				// Refresh threshold request.
				mockAuth.EXPECT().RefreshThreshold().
					Return(test.authRefThreshold).
					Times(test.authRefThresholdTimes),

				// Generate fresh JWT.
				mockAuth.EXPECT().GenerateJWT(gomock.Any()).
					Return(&models.JWTAuthResponse{}, test.authGenJWTErr).
					Times(test.authGenJWTTimes),
			)

			// Endpoint setup for test.
			router := gin.Default()
			router.Use(GinContextToContextMiddleware())
			router.POST(test.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockPostgres, mockQuotes, zapLogger))

			req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, test.path,
				bytes.NewBufferString(testUserQuery["refresh"]))
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
