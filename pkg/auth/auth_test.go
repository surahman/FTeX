package auth

import (
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
)

func TestNewAuth(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
	require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetAuthFileName(),
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
			fileName:  constants.GetAuthFileName(),
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
			require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+testCase.fileName,
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

	userName := "test username"
	authResponse, err := testAuth.GenerateJWT(userName)
	require.NoError(t, err, "JWT creation failed")
	require.True(t, authResponse.Expires > time.Now().Unix(), "JWT expires before current time")
	require.True(t, authResponse.Expires < time.Now().Add(time.Duration(expirationDuration+1)*time.Second).Unix(),
		"JWT expires after deadline")

	// Check validate token and check for username in claim.
	actualUname, expiresAt, err := testAuth.ValidateJWT(authResponse.Token)
	require.NoError(t, err, "failed to extract username from JWT")
	require.Equalf(t, userName, actualUname, "incorrect username retrieved from JWT")
	require.True(t, expiresAt > 0, "invalid expiration time")
}

func TestAuthImpl_ValidateJWT(t *testing.T) {
	t.Parallel()

	t.Run("JWT claim parsing tests", func(t *testing.T) {
		t.Parallel()

		testAuthImpl := getTestConfiguration()

		_, _, err := testAuthImpl.ValidateJWT("")
		require.Error(t, err, "parsing an empty token should fail")

		_, _, err = testAuthImpl.ValidateJWT("bad#token#string")
		require.Error(t, err, "parsing and invalid token should fail")
	})

	const testUsername = "test username"

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
			testAuthImpl := getTestConfiguration()
			if test.issuerName != "" {
				testAuthImpl.conf.JWTConfig.Issuer = test.issuerName
			}
			if test.expirationDuration != 0 {
				testAuthImpl.conf.JWTConfig.ExpirationDuration = test.expirationDuration
			}

			// Generate test token.
			testJWT, err := testAuthImpl.GenerateJWT(testUsername)
			require.NoError(t, err, "failed to create test JWT")

			// Conditional sleep to expire token.
			if test.expirationDuration > 0 {
				time.Sleep(time.Duration(test.expirationDuration+1) * time.Second)
			}

			username, expiresAt, err := testAuth.ValidateJWT(testJWT.Token)
			test.expectErr(t, err, "validation of issued token error condition failed")

			if err != nil {
				require.Contains(t, err.Error(), test.expectErrMsg, "error message did not contain expected err")

				return
			}

			require.Equal(t, testUsername, username, "username failed to match the expected")
			require.True(t, expiresAt > time.Now().Unix(), "invalid expiration time")
		})
	}
}

func TestAuthImpl_RefreshJWT(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		testUsername       string
		expirationDuration int64
		sleepTime          int
		expectErr          require.ErrorAssertionFunc
	}{
		{
			name:               "Valid token",
			testUsername:       "test username",
			expirationDuration: 4,
			sleepTime:          2,
			expectErr:          require.NoError,
		}, {
			name:               "Invalid token",
			testUsername:       "test username",
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
			testAuthImpl := getTestConfiguration()

			testAuthImpl.conf.JWTConfig.ExpirationDuration = test.expirationDuration

			testJWT, err := testAuthImpl.GenerateJWT(test.testUsername)
			require.NoError(t, err, "failed to create initial JWT")

			actualUsername, expiresAt, err := testAuthImpl.ValidateJWT(testJWT.Token)
			require.NoError(t, err, "failed to validate original test token")
			require.Equal(t, test.testUsername, actualUsername, "failed to extract correct username from original JWT")
			require.True(t, expiresAt > 0, "invalid expiration time of original token")

			time.Sleep(time.Duration(test.sleepTime) * time.Second)
			refreshedToken, err := testAuthImpl.RefreshJWT(testJWT.Token)
			test.expectErr(t, err, "error case when refreshing JWT failed")

			if err != nil {
				return
			}

			require.True(t,
				refreshedToken.Expires > time.Now().Add(
					time.Duration(testAuthImpl.conf.JWTConfig.ExpirationDuration-1)*time.Second).Unix(),
				"token expires before the required deadline")

			actualUsername, expiresAt, err = testAuthImpl.ValidateJWT(testJWT.Token)
			require.NoErrorf(t, err, "failed to validate refreshed JWT")
			require.Equal(t, test.testUsername, actualUsername, "failed to extract correct username from refreshed JWT")
			require.True(t, expiresAt > 0, "invalid expiration time of refreshed token")
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
	require.True(t, len(ciphertext) > 0, "encrypted string is empty")

	plaintext, err := testAuth.DecryptFromString(ciphertext)
	require.NoError(t, err, "encrypt from string failed")
	require.Equal(t, toEncrypt, string(plaintext), "decrypted string does not match original")
}
