package graphql

import (
	"fmt"
	"time"

	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/configloader"
	"github.com/surahman/FTeX/pkg/constants"
)

// config is the configuration container for the HTTP REST endpoint.
//
//nolint:lll
type config struct {
	Server        serverConfig        `json:"server,omitempty" yaml:"server,omitempty" mapstructure:"server" validate:"required"`
	Authorization authorizationConfig `json:"authorization,omitempty" yaml:"authorization,omitempty" mapstructure:"authorization" validate:"required"`
}

// serverConfig contains the configurations for the HTTP REST server.
//
//nolint:lll
type serverConfig struct {
	BasePath          string        `json:"basePath,omitempty" yaml:"basePath,omitempty" mapstructure:"basePath" validate:"required"`
	PlaygroundPath    string        `json:"playgroundPath,omitempty" yaml:"playgroundPath,omitempty" mapstructure:"playgroundPath" validate:"required"`
	PortNumber        int           `json:"portNumber,omitempty" yaml:"portNumber,omitempty" mapstructure:"portNumber" validate:"required,min=1000"`
	ShutdownDelay     time.Duration `json:"shutdownDelay,omitempty" yaml:"shutdownDelay,omitempty" mapstructure:"shutdownDelay" validate:"required,min=0"`
	ReadTimeout       time.Duration `json:"readTimeout,omitempty" yaml:"readTimeout,omitempty" mapstructure:"readTimeout" validate:"required,min=1"`
	WriteTimeout      time.Duration `json:"writeTimeout,omitempty" yaml:"writeTimeout,omitempty" mapstructure:"writeTimeout" validate:"required,min=1"`
	ReadHeaderTimeout time.Duration `json:"readHeaderTimeout,omitempty" yaml:"readHeaderTimeout,omitempty" mapstructure:"readHeaderTimeout" validate:"required,min=1"`
}

// authorizationConfig contains the configurations for request authorization.
type authorizationConfig struct {
	HeaderKey string `json:"headerKey,omitempty" yaml:"headerKey,omitempty" mapstructure:"headerKey" validate:"required"`
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
		constants.GetHTTPGraphQLFileName(),
		constants.GetHTTPGraphQLPrefix(),
		"yaml"); err != nil {
		return fmt.Errorf("rest config loading failed: %w", err)
	}

	return nil
}
