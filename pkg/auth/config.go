package auth

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/configloader"
	"github.com/surahman/FTeX/pkg/constants"
)

// Config contains all the configurations for authentication.
type config struct {
	JWTConfig jwtConfig     `json:"jwt,omitempty"     mapstructure:"jwt"     validate:"required" yaml:"jwt,omitempty"`
	General   generalConfig `json:"general,omitempty" mapstructure:"general" validate:"required" yaml:"general,omitempty"`
}

// jwtConfig contains the configurations for JWT creation and verification.
//
//nolint:lll
type jwtConfig struct {
	Key                string `json:"key,omitempty"                mapstructure:"key"                validate:"required,min=8,max=256"                     yaml:"key,omitempty"`
	Issuer             string `json:"issuer,omitempty"             mapstructure:"issuer"             validate:"required"                                   yaml:"issuer,omitempty"`
	ExpirationDuration int64  `json:"expirationDuration,omitempty" mapstructure:"expirationDuration" validate:"required,min=60,gtefield=RefreshThreshold"  yaml:"expirationDuration,omitempty"`
	RefreshThreshold   int64  `json:"refreshThreshold,omitempty"   mapstructure:"refreshThreshold"   validate:"required,min=1,ltefield=ExpirationDuration" yaml:"refreshThreshold,omitempty"`
}

// generalConfig contains the configurations for general encryption.
//
//nolint:lll
type generalConfig struct {
	BcryptCost   int    `json:"bcryptCost,omitempty"   mapstructure:"bcryptCost"   validate:"required,min=4,max=31" yaml:"bcryptCost,omitempty"`
	CryptoSecret string `json:"cryptoSecret,omitempty" mapstructure:"cryptoSecret" validate:"required,len=32"       yaml:"cryptoSecret,omitempty"`
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
