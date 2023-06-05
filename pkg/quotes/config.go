package quotes

import (
	"fmt"
	"time"

	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/configloader"
	"github.com/surahman/FTeX/pkg/constants"
)

// config is the configuration container for connecting to external Quotes services.
//
//nolint:lll
type config struct {
	FiatCurrency   apiConfig        `json:"fiatCurrency,omitempty"   mapstructure:"fiatCurrency"   validate:"required"         yaml:"fiatCurrency,omitempty"`
	CryptoCurrency apiConfig        `json:"cryptoCurrency,omitempty" mapstructure:"cryptoCurrency" validate:"required"         yaml:"cryptoCurrency,omitempty"`
	Connection     connectionConfig `json:"connection,omitempty"     mapstructure:"connection"     yaml:"connection,omitempty"`
}

// apiConfig contains the API Key and URL information for a currency exchange endpoint.
type apiConfig struct {
	APIKey    string `json:"apiKey,omitempty"    mapstructure:"apiKey"    validate:"required" yaml:"apiKey,omitempty"`
	HeaderKey string `json:"headerKey,omitempty" mapstructure:"headerKey" validate:"required" yaml:"headerKey,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"  mapstructure:"endpoint"  validate:"required" yaml:"endpoint,omitempty"`
}

// connectionConfig contains HTTP connection attempt information.
//
//nolint:lll
type connectionConfig struct {
	UserAgent string        `json:"userAgent,omitempty" mapstructure:"userAgent" validate:"required" yaml:"userAgent,omitempty"`
	Timeout   time.Duration `json:"timeout,omitempty"   mapstructure:"timeout"   validate:"required" yaml:"timeout,omitempty"`
}

// newConfig creates a blank configuration struct for Redis.
func newConfig() *config {
	return &config{}
}

// Load will attempt to load configurations from a file on a file system.
func (cfg *config) Load(fs afero.Fs) (err error) {
	if err := configloader.Load(
		fs,
		cfg,
		constants.GetQuotesFileName(),
		constants.GetQuotesPrefix(),
		"yaml"); err != nil {
		return fmt.Errorf("quotes config loading failed: %w", err)
	}

	return nil
}
