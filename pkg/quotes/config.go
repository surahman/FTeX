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
	FiatCurrency   apiConfig        `json:"fiatCurrency,omitempty" yaml:"fiatCurrency,omitempty" mapstructure:"fiatCurrency" validate:"required"`
	CryptoCurrency apiConfig        `json:"cryptoCurrency,omitempty" yaml:"cryptoCurrency,omitempty" mapstructure:"cryptoCurrency" validate:"required"`
	Connection     connectionConfig `json:"connection,omitempty" yaml:"connection,omitempty" mapstructure:"connection"`
}

// apiConfig contains the API Key and URL information for a currency exchange endpoint.
type apiConfig struct {
	APIKey    string `json:"apiKey,omitempty" yaml:"apiKey,omitempty" mapstructure:"apiKey" validate:"required"`
	HeaderKey string `json:"headerKey,omitempty" yaml:"headerKey,omitempty" mapstructure:"headerKey" validate:"required"`
	Endpoint  string `json:"endpoint,omitempty" yaml:"endpoint,omitempty" mapstructure:"endpoint" validate:"required"`
}

// connectionConfig contains HTTP connection attempt information.
//
//nolint:lll
type connectionConfig struct {
	UserAgent string        `json:"userAgent,omitempty" yaml:"userAgent,omitempty" mapstructure:"userAgent" validate:"required"`
	Timeout   time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty" mapstructure:"timeout" validate:"required"`
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
