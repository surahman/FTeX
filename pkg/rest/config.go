package rest

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
	Server        serverConfig        `json:"server,omitempty"        mapstructure:"server"        validate:"required" yaml:"server,omitempty"`
	Authorization authorizationConfig `json:"authorization,omitempty" mapstructure:"authorization" validate:"required" yaml:"authorization,omitempty"`
}

// serverConfig contains the configurations for the HTTP REST server.
//
//nolint:lll
type serverConfig struct {
	BasePath          string        `json:"basePath,omitempty"          mapstructure:"basePath"          validate:"required"          yaml:"basePath,omitempty"`
	SwaggerPath       string        `json:"swaggerPath,omitempty"       mapstructure:"swaggerPath"       validate:"required"          yaml:"swaggerPath,omitempty"`
	PortNumber        int           `json:"portNumber,omitempty"        mapstructure:"portNumber"        validate:"required,min=1000" yaml:"portNumber,omitempty"`
	ShutdownDelay     time.Duration `json:"shutdownDelay,omitempty"     mapstructure:"shutdownDelay"     validate:"required,min=0"    yaml:"shutdownDelay,omitempty"`
	ReadTimeout       time.Duration `json:"readTimeout,omitempty"       mapstructure:"readTimeout"       validate:"required,min=1"    yaml:"readTimeout,omitempty"`
	WriteTimeout      time.Duration `json:"writeTimeout,omitempty"      mapstructure:"writeTimeout"      validate:"required,min=1"    yaml:"writeTimeout,omitempty"`
	ReadHeaderTimeout time.Duration `json:"readHeaderTimeout,omitempty" mapstructure:"readHeaderTimeout" validate:"required,min=1"    yaml:"readHeaderTimeout,omitempty"`
}

// authorizationConfig contains the configurations for request authorization.
type authorizationConfig struct {
	HeaderKey string `json:"headerKey,omitempty" mapstructure:"headerKey" validate:"required" yaml:"headerKey,omitempty"`
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
		constants.HTTPRESTFileName(),
		constants.HTTPRESTPrefix(),
		"yaml"); err != nil {
		return fmt.Errorf("rest config loading failed: %w", err)
	}

	return nil
}
