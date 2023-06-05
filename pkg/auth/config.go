package auth

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/configloader"
	"github.com/surahman/FTeX/pkg/constants"
)

// Config contains all the configurations for authentication.
type config struct {
	JWTConfig jwtConfig     `json:"jwt,omitempty"     yaml:"jwt,omitempty"     mapstructure:"jwt"     validate:"required"`
	General   generalConfig `json:"general,omitempty" yaml:"general,omitempty" mapstructure:"general" validate:"required"`
}

// jwtConfig contains the configurations for JWT creation and verification.
//
//nolint:lll
type jwtConfig struct {
	Key                string `json:"key,omitempty"                yaml:"key,omitempty"                mapstructure:"key"                validate:"required,min=8,max=256"`
	Issuer             string `json:"issuer,omitempty"             yaml:"issuer,omitempty"             mapstructure:"issuer"             validate:"required"`
	ExpirationDuration int64  `json:"expirationDuration,omitempty" yaml:"expirationDuration,omitempty" mapstructure:"expirationDuration" validate:"required,min=60,gtefield=RefreshThreshold"`
	RefreshThreshold   int64  `json:"refreshThreshold,omitempty"   yaml:"refreshThreshold,omitempty"   mapstructure:"refreshThreshold"   validate:"required,min=1,ltefield=ExpirationDuration"`
}

// generalConfig contains the configurations for general encryption.
//
//nolint:lll
type generalConfig struct {
	BcryptCost   int    `json:"bcryptCost,omitempty"   yaml:"bcryptCost,omitempty"   mapstructure:"bcryptCost"   validate:"required,min=4,max=31"`
	CryptoSecret string `json:"cryptoSecret,omitempty" yaml:"cryptoSecret,omitempty" mapstructure:"cryptoSecret" validate:"required,len=32"`
}

// newConfig creates a blank configuration struct for the authorization.
func newConfig() *config {
	return &config{}
}

// Load will attempt to load configurations from a file on a file system.
func (cfg *config) Load(fs afero.Fs) error {
	if err := configloader.Load(
		fs,
		cfg,
		constants.GetAuthFileName(),
		constants.GetAuthPrefix(),
		"yaml"); err != nil {
		return fmt.Errorf("authorization config loading failed: %w", err)
	}

	return nil
}
