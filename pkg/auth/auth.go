package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Mock Auth interface stub generation.
//go:generate mockgen -destination=../mocks/mock_auth.go -package=mocks github.com/surahman/FTeX/pkg/auth Auth

// Auth is the interface through which the authorization operations can be accessed. Created to support mock testing.
type Auth interface {
	// HashPassword will take a plaintext string and generate a hashed representation of it.
	HashPassword(string) (string, error)

	// CheckPassword will take the plaintext and hashed passwords as input, in that order, and verify if they match.
	CheckPassword(string, string) error

	// GenerateJWT will create a valid JSON Web Token and return it in a JWT Authorization Response structure.
	GenerateJWT(string) (*models.JWTAuthResponse, error)

	// ValidateJWT will take the JSON Web Token and validate it. It will extract and return the username and expiration
	// time (Unix timestamp) or an error if validation fails.
	ValidateJWT(string) (string, int64, error)

	// RefreshJWT will take a valid JSON Web Token, and if valid and expiring soon, issue a fresh valid JWT with the time
	// extended in JWT Authorization Response structure.
	RefreshJWT(string) (*models.JWTAuthResponse, error)

	// RefreshThreshold returns the time before the end of the JSON Web Tokens validity interval that a JWT can be
	// refreshed in.
	RefreshThreshold() int64

	// EncryptToString will generate an encrypted base64 encoded character from the plaintext.
	EncryptToString([]byte) (string, error)

	// DecryptFromString will decrypt an encrypted base64 encoded character from the ciphertext.
	DecryptFromString(string) ([]byte, error)
}

// Check to ensure the Auth interface has been implemented.
var _ Auth = &authImpl{}

// authImpl implements the Auth interface and contains the logic for authorization functionality.
type authImpl struct {
	cryptoSecret []byte
	conf         *config
	logger       *logger.Logger
}

// NewAuth will create a new Authorization configuration by loading it.
func NewAuth(fs *afero.Fs, logger *logger.Logger) (Auth, error) {
	if fs == nil || logger == nil {
		return nil, errors.New("nil file system of logger supplied")
	}

	return newAuthImpl(fs, logger)
}

// newAuthImpl will create a new authImpl configuration and load it from disk.
func newAuthImpl(fs *afero.Fs, logger *logger.Logger) (a *authImpl, err error) {
	a = &authImpl{conf: newConfig(), logger: logger}
	if err = a.conf.Load(*fs); err != nil {
		a.logger.Error("failed to load Authorization configurations from disk", zap.Error(err))

		return nil, err
	}

	a.cryptoSecret = []byte(a.conf.General.CryptoSecret)

	return
}

// HashPassword hashes a password using the Bcrypt algorithm to avoid plaintext storage.
func (a *authImpl) HashPassword(plaintext string) (hashed string, err error) {
	var bytes []byte

	if bytes, err = bcrypt.GenerateFromPassword([]byte(plaintext), a.conf.General.BcryptCost); err != nil {
		return
	}

	hashed = string(bytes)

	return
}

// CheckPassword checks a hashed password against a plaintext password using the Bcrypt algorithm.
func (a *authImpl) CheckPassword(hashed, plaintext string) (err error) {
	if err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plaintext)); err != nil {
		return
	}

	return
}

// jwtClaim is used internally by the JWT generation and validation routines.
type jwtClaim struct {
	Username string `json:"username" yaml:"username"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a payload consisting of the JWT with the username as well as expiration time.
func (a *authImpl) GenerateJWT(username string) (*models.JWTAuthResponse, error) {
	claims := &jwtClaim{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: a.conf.JWTConfig.Issuer,
			ExpiresAt: jwt.NewNumericDate(
				time.
					Now().
					Add(time.Duration(a.conf.JWTConfig.ExpirationDuration) * time.Second).UTC()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(a.conf.JWTConfig.Key))
	if err != nil {
		msg := "failed to generate signed jwt"
		a.logger.Warn(msg, zap.Error(err))

		return nil, fmt.Errorf(msg+" %w", err)
	}

	authResponse := &models.JWTAuthResponse{
		Token:     tokenString,
		Expires:   claims.ExpiresAt.Unix(),
		Threshold: a.conf.JWTConfig.RefreshThreshold,
	}

	return authResponse, nil
}

// ValidateJWT will validate a signed JWT and extracts the username and unix expiration timestamp from it.
func (a *authImpl) ValidateJWT(signedToken string) (string, int64, error) {
	token, err := jwt.ParseWithClaims(signedToken, &jwtClaim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.conf.JWTConfig.Key), nil
	})
	if err != nil {
		msg := "failed to parse token"
		a.logger.Warn(msg, zap.Error(err))

		return "", -1, fmt.Errorf(msg+" %w", err)
	}

	// Cast token claim to JWT.
	claims, ok := token.Claims.(*jwtClaim)
	if !ok || !token.Valid {
		msg := "failed to extract jwt data"
		a.logger.Warn(msg, zap.Error(err))

		return "", -1, fmt.Errorf(msg+" %w", err)
	}

	// Check for errors and compare the expiration time in Unix format.
	expiration, err := claims.GetExpirationTime()
	if err != nil || expiration.Unix() < time.Now().Unix() {
		return "", -1, errors.New("token has expired")
	}

	// Check the issuer is correct.
	issuer, err := claims.GetIssuer()
	if err != nil || issuer != a.conf.JWTConfig.Issuer {
		return "", -1, errors.New("unauthorized issuer")
	}

	// Return the username and the unix expiration timestamp.
	return claims.Username, claims.ExpiresAt.Unix(), nil
}

// RefreshJWT will extend a valid JWTs lease by generating a fresh valid JWT.
func (a *authImpl) RefreshJWT(token string) (authResponse *models.JWTAuthResponse, err error) {
	var username string

	if username, _, err = a.ValidateJWT(token); err != nil {
		return
	}

	if authResponse, err = a.GenerateJWT(username); err != nil {
		return
	}

	return
}

// RefreshThreshold is the seconds before expiration that a JWT can be refreshed in.
func (a *authImpl) RefreshThreshold() int64 {
	return a.conf.JWTConfig.RefreshThreshold
}

// encryptAES256 employs Authenticated Encryption with Associated Data using Galois/Counter mode and returns the cipher
// as a Base64 encoded string to be used in URIs.
func (a *authImpl) encryptAES256(data []byte) (cipherStr string, cipherBytes []byte, err error) {
	var (
		cipherBlock cipher.Block
		gcm         cipher.AEAD
	)

	if cipherBlock, err = aes.NewCipher(a.cryptoSecret); err != nil {
		return
	}

	if gcm, err = cipher.NewGCM(cipherBlock); err != nil {
		return
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}

	// Encrypt to cipher text.
	cipherBytes = gcm.Seal(nonce, nonce, data, nil)

	// Convert to Base64 URL encoded string for use in URLs.
	cipherStr = base64.URLEncoding.EncodeToString(cipherBytes)

	return
}

// decryptAES256 employs Authenticated Encryption with Associated Data using Galois/Counter mode and returns the
// decrypted plaintext bytes.
func (a *authImpl) decryptAES256(data []byte) (cipherBytes []byte, err error) {
	var (
		cipherBlock cipher.Block
		gcm         cipher.AEAD
		nonceSize   int
	)

	if cipherBlock, err = aes.NewCipher(a.cryptoSecret); err != nil {
		return
	}

	if gcm, err = cipher.NewGCM(cipherBlock); err != nil {
		return
	}

	if nonceSize = gcm.NonceSize(); nonceSize < 0 {
		return nil, errors.New("bad nonce size")
	}

	// Extract the nonce and cipher blocks from the data.
	nonce, cipherText := data[:nonceSize], data[nonceSize:]

	// Decrypt cipher text.
	cipherBytes, err = gcm.Open(nil, nonce, cipherText, nil)

	return
}

// EncryptToString will generate an encrypted base64 encoded character from the plaintext.
func (a *authImpl) EncryptToString(plaintext []byte) (ciphertext string, err error) {
	ciphertext, _, err = a.encryptAES256(plaintext)

	return
}

// DecryptFromString will decrypt an encrypted base64 encoded character from the ciphertext.
func (a *authImpl) DecryptFromString(ciphertext string) (plaintext []byte, err error) {
	var bytes []byte

	if bytes, err = base64.URLEncoding.DecodeString(ciphertext); err != nil {
		return
	}

	return a.decryptAES256(bytes)
}
