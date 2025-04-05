package auth

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
)

func TestNewAuth(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(constants.EtcDir(), 0644), "Failed to create in memory directory")
	require.NoError(t, afero.WriteFile(fs, constants.EtcDir()+constants.AuthFileName(),
		[]byte(authConfigTestData["valid"]), 0644), "Failed to write in memory file")

	testCases := []struct {
		name      string
		fs        *afero.Fs
		log       *logger.Logger
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
	}{
		{
			name:      "Invalid file system and logger",
			fs:        nil,
			log:       nil,
			expectErr: require.Error,
			expectNil: require.Nil,
		}, {
			name:      "Invalid file system",
			fs:        nil,
			log:       zapLogger,
			expectErr: require.Error,
			expectNil: require.Nil,
		}, {
			name:      "Invalid logger",
			fs:        &fs,
			log:       nil,
			expectErr: require.Error,
			expectNil: require.Nil,
		}, {
			name:      "Valid",
			fs:        &fs,
			log:       zapLogger,
			expectErr: require.NoError,
			expectNil: require.NotNil,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			auth, err := NewAuth(test.fs, test.log)
			test.expectErr(t, err, "error expectation failed.")
			test.expectNil(t, auth, "auth nil return expectation failed.")
		})
	}
}

func TestNewAuthImpl(t *testing.T) {
	testCases := []struct {
		name      string
		fileName  string
		input     string
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
	}{
		{
			name:      "File found",
			fileName:  constants.AuthFileName(),
			input:     authConfigTestData["valid"],
			expectErr: require.NoError,
			expectNil: require.NotNil,
		}, {
			name:      "File not found",
			fileName:  "wrong_file_name.yaml",
			input:     authConfigTestData["valid"],
			expectErr: require.Error,
			expectNil: require.Nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Configure mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(constants.EtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, constants.EtcDir()+testCase.fileName,
				[]byte(testCase.input), 0644), "Failed to write in memory file")

			auth, err := NewAuth(&fs, zapLogger)
			testCase.expectErr(t, err, "error expectation failed.")
			testCase.expectNil(t, auth, "auth nil return expectation failed.")
		})
	}
}

func TestAuthImpl_HashPassword(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     string
		expectErr require.ErrorAssertionFunc
	}{
		{
			name:      "Empty password",
			input:     "",
			expectErr: require.NoError,
		}, {
			name:      "Valid",
			input:     "ELy@FRrn7DW8Cj1QQj^zG&X%$9cjVU4R",
			expectErr: require.NoError,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := testAuth.HashPassword(test.input)
			test.expectErr(t, err, "error expectation failed.")
		})
	}
}

func TestAuthImpl_CheckPassword(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		plaintext string
		hashed    string
		expectErr require.ErrorAssertionFunc
	}{
		{
			name:      "Empty password's hash",
			plaintext: "",
			hashed:    "$2a$08$vZhD311uyi8FnkjyoT.1req7ixf0CXRARozPTaj4gnhr/F3m/q7NW",
			expectErr: require.NoError,
		}, {
			name:      "Valid password's hash",
			plaintext: "ELy@FRrn7DW8Cj1QQj^zG&X%$9cjVU4R",
			hashed:    "$2a$08$YXYc8lyxnS7VPy6f28Gmd.udRTVrxKewXX9E3ULs0/ynkTL6PY/0K",
			expectErr: require.NoError,
		}, {
			name:      "Invalid password's hash",
			plaintext: "$2a$08$YXYc8lyxnS7VPy6f28Gmd.udRTVrxKewXX9E3ULs0/ynkTL6PY/0K",
			hashed:    "$2a$08$YXYc8lyxnS7VPy6f28Gmd.udRTVrxKewXX9E3ULs0/ynkTL6PY/0K",
			expectErr: require.Error,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := testAuth.CheckPassword(test.hashed, test.plaintext)

			test.expectErr(t, err, "error expectation failed.")
		})
	}
}

func TestAuthImpl_GenerateJWT(t *testing.T) {
	t.Parallel()

	clientID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate clientID.")

	authResponse, err := testAuth.GenerateJWT(clientID)
	require.NoError(t, err, "JWT creation failed")
	require.Greater(t, authResponse.Expires, time.Now().Unix(), "JWT expires before current time.")
	require.Less(t, authResponse.Expires, time.Now().Add(time.Duration(expirationDuration+1)*time.Second).Unix(),
		"JWT expires after deadline.")

	// Check validate token and check for username in claim.
	actualUUID, expiresAt, err := testAuth.ValidateJWT(authResponse.Token)
	require.NoError(t, err, "failed to extract information from JWT.")
	require.Equal(t, actualUUID, clientID, "incorrect clientID retrieved from JWT.")
	require.Positive(t, expiresAt, "invalid expiration time")
}

func TestAuthImpl_ValidateJWT(t *testing.T) {
	t.Parallel()

	t.Run("JWT claim parsing tests", func(t *testing.T) {
		t.Parallel()

		testAuthImpl := testConfigurationImpl(zapLogger, expirationDuration, refreshThreshold)

		_, _, err := testAuthImpl.ValidateJWT("")
		require.Error(t, err, "parsing an empty token should fail")

		_, _, err = testAuthImpl.ValidateJWT("bad#token#string")
		require.Error(t, err, "parsing and invalid token should fail")
	})

	clientID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate clientID.")

	testCases := []struct {
		name               string
		issuerName         string
		expectErrMsg       string
		expirationDuration int64
		expectErr          require.ErrorAssertionFunc
	}{
		{
			name:               "Valid token",
			issuerName:         "",
			expectErrMsg:       "",
			expirationDuration: 0,
			expectErr:          require.NoError,
		}, {
			name:               "Invalid issuer token",
			issuerName:         "some random name",
			expectErrMsg:       "issuer",
			expirationDuration: 0,
			expectErr:          require.Error,
		}, {
			name:               "Invalid expired token",
			issuerName:         "some random name",
			expectErrMsg:       "expired",
			expirationDuration: 1,
			expectErr:          require.Error,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Adjust the authorization issuer name and expiration deadline for tests, if needed.
			testAuthImpl := testConfigurationImpl(zapLogger, expirationDuration, refreshThreshold)
			if test.issuerName != "" {
				testAuthImpl.conf.JWTConfig.Issuer = test.issuerName
			}

			if test.expirationDuration != 0 {
				testAuthImpl.conf.JWTConfig.ExpirationDuration = test.expirationDuration
			}

			// Generate test token.
			testJWT, err := testAuthImpl.GenerateJWT(clientID)
			require.NoError(t, err, "failed to create test JWT")

			// Conditional sleep to expire token.
			if test.expirationDuration > 0 {
				time.Sleep(time.Duration(test.expirationDuration+1) * time.Second)
			}

			actualClientID, expiresAt, err := testAuth.ValidateJWT(testJWT.Token)
			test.expectErr(t, err, "validation of issued token error condition failed")

			if err != nil {
				require.Contains(t, err.Error(), test.expectErrMsg, "error message did not contain expected err")

				return
			}

			require.Equal(t, clientID, actualClientID, "clientId failed to match the expected")
			require.Greater(t, expiresAt, time.Now().Unix(), "invalid expiration time")
		})
	}
}

func TestAuthImpl_RefreshJWT(t *testing.T) {
	t.Parallel()

	clientID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate clientID.")

	testCases := []struct {
		name               string
		expirationDuration int64
		sleepTime          int
		expectErr          require.ErrorAssertionFunc
	}{
		{
			name:               "Valid token",
			expirationDuration: 4,
			sleepTime:          2,
			expectErr:          require.NoError,
		}, {
			name:               "Invalid token",
			expirationDuration: 1,
			sleepTime:          2,
			expectErr:          require.Error,
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Test authorization config and token generation.
			testAuthImpl := testConfigurationImpl(zapLogger, expirationDuration, refreshThreshold)

			testAuthImpl.conf.JWTConfig.ExpirationDuration = test.expirationDuration

			testJWT, err := testAuthImpl.GenerateJWT(clientID)
			require.NoError(t, err, "failed to create initial JWT")

			actualClientID, expiresAt, err := testAuthImpl.ValidateJWT(testJWT.Token)
			require.NoError(t, err, "failed to validate original test token")
			require.Equal(t, clientID, actualClientID, "failed to extract correct clientID from original JWT")
			require.Positive(t, expiresAt, "invalid expiration time of original token")

			time.Sleep(time.Duration(test.sleepTime) * time.Second)

			refreshedToken, err := testAuthImpl.RefreshJWT(testJWT.Token)

			test.expectErr(t, err, "error case when refreshing JWT failed")

			if err != nil {
				return
			}

			require.Greater(t,
				refreshedToken.Expires, time.Now().Add(
					time.Duration(testAuthImpl.conf.JWTConfig.ExpirationDuration-1)*time.Second).Unix(),
				"token expires before the required deadline")

			actualClientID, expiresAt, err = testAuthImpl.ValidateJWT(testJWT.Token)
			require.NoErrorf(t, err, "failed to validate refreshed JWT")
			require.Equal(t, clientID, actualClientID, "failed to extract correct clientID from refreshed JWT")
			require.Positive(t, expiresAt, "invalid expiration time of refreshed token")
		})
	}
}

func TestAuthImpl_RefreshThreshold(t *testing.T) {
	t.Parallel()

	require.Equal(t, refreshThreshold, testAuth.RefreshThreshold(),
		"token refresh threshold did not match expected threshold")
}

func TestAuthImpl_encryptAES256_and_decryptAES256(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		plainText string
		expectStr require.BoolAssertionFunc
	}{
		{
			name:      "to string",
			plainText: "this text should be encrypted",
			expectStr: require.True,
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Encrypt phase.
			cipherStr, cipherBytes, err := testAuth.encryptAES256([]byte(test.plainText))
			require.NoError(t, err, "error encrypting to AES256")
			require.NotNil(t, cipherBytes, "no cipher block returned as bytes")
			test.expectStr(t, len(cipherStr) > 0, "cipher string expectation failed")

			// Decrypt phase.
			plainTextBytes, err := testAuth.decryptAES256(cipherBytes)
			require.NoError(t, err, "error decrypting from AES256")
			require.NotNil(t, plainTextBytes, "no plaintext block returned as bytes")
			require.Equal(t, test.plainText, string(plainTextBytes), "decrypted cipher does not match input plaintext")
		})
	}
}

func TestAuthImpl_Encrypt_Decrypt_String(t *testing.T) {
	t.Parallel()

	toEncrypt := "this is a text string to be encrypted/decrypted"

	ciphertext, err := testAuth.EncryptToString([]byte(toEncrypt))
	require.NoError(t, err, "encrypt to string failed")
	require.NotEmpty(t, ciphertext, "encrypted string is empty")

	plaintext, err := testAuth.DecryptFromString(ciphertext)
	require.NoError(t, err, "encrypt from string failed")
	require.Equal(t, toEncrypt, string(plaintext), "decrypted string does not match original")
}

func TestAuth_AuthFromGinCtx(t *testing.T) {
	t.Parallel()

	validUUID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate valid clientID.")

	validExpiration := int64(99)

	valid := gin.Context{}
	valid.Set(constants.ClientIDCtxKey(), validUUID)
	valid.Set(constants.ExpiresAtCtxKey(), validExpiration)

	noClientID := gin.Context{}
	noClientID.Set(constants.ExpiresAtCtxKey(), validExpiration)

	badClientID := gin.Context{}
	badClientID.Set(constants.ClientIDCtxKey(), "random-string-here")
	badClientID.Set(constants.ExpiresAtCtxKey(), validExpiration)

	noExpiration := gin.Context{}
	noExpiration.Set(constants.ClientIDCtxKey(), validUUID)

	badExpiration := gin.Context{}
	badExpiration.Set(constants.ClientIDCtxKey(), validUUID)
	badExpiration.Set(constants.ExpiresAtCtxKey(), "random-string-here")

	testCases := []struct {
		name         string
		expectErrMsg string
		expectedUUID uuid.UUID
		expectedTime int64
		ctx          *gin.Context
		expectErr    require.ErrorAssertionFunc
	}{
		{
			name:         "bad expiration",
			expectErrMsg: "locate expiration",
			expectedUUID: validUUID,
			expectedTime: 0,
			ctx:          &badExpiration,
			expectErr:    require.Error,
		}, {
			name:         "no expiration",
			expectErrMsg: "locate expiration",
			expectedUUID: validUUID,
			expectedTime: 0,
			ctx:          &noExpiration,
			expectErr:    require.Error,
		}, {
			name:         "bad client id",
			expectErrMsg: "parse clientID",
			expectedUUID: uuid.UUID{},
			expectedTime: 0,
			ctx:          &badClientID,
			expectErr:    require.Error,
		}, {
			name:         "no client id",
			expectErrMsg: "locate clientID",
			expectedUUID: uuid.UUID{},
			expectedTime: 0,
			ctx:          &noClientID,
			expectErr:    require.Error,
		}, {
			name:         "valid",
			expectErrMsg: "",
			expectedUUID: validUUID,
			expectedTime: validExpiration,
			ctx:          &valid,
			expectErr:    require.NoError,
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			clientID, expiresAt, err := testAuth.TokenInfoFromGinCtx(test.ctx)
			test.expectErr(t, err, "error expectation failed.")
			require.Equal(t, test.expectedUUID, clientID, "clientID mismatched.")
			require.Equal(t, test.expectedTime, expiresAt, "expiration deadline mismatched.")

			if err != nil {
				require.Contains(t, err.Error(), test.expectErrMsg, "error message mismatch.")
			}
		})
	}
}
