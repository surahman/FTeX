package postgres

import (
	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/config_loader"
	"github.com/surahman/FTeX/pkg/constants"
)

// config contains the configurations loaded from the configuration file.
type config struct {
	Authentication *authenticationConfig `json:"authentication,omitempty" yaml:"authentication,omitempty" mapstructure:"authentication"`
	Connection     *connectionConfig     `json:"connection,omitempty" yaml:"connection,omitempty" mapstructure:"connection"`
}

// authenticationConfig contains the Postgres session authentication information.
type authenticationConfig struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty" mapstructure:"username" validate:"required"`
	Password string `json:"password,omitempty" yaml:"password,omitempty" mapstructure:"password" validate:"required"`
}

// connectionConfig contains the Postgres session connection information.
type connectionConfig struct {
	Hostname   string `json:"hostname,omitempty" yaml:"hostname,omitempty" mapstructure:"hostname" validate:"required"`
	Port       uint16 `json:"port,omitempty" yaml:"port,omitempty" mapstructure:"port" validate:"required"`
	Timeout    uint16 `json:"timeout,omitempty" yaml:"timeout,omitempty" mapstructure:"timeout" validate:"required"`
	SslEnabled bool   `json:"ssl_enabled,omitempty" yaml:"ssl_enabled,omitempty" mapstructure:"ssl_enabled"`
}

// newConfig creates a blank configuration struct for the Zap Logger.
func newConfig() *config {
	return &config{}
}

// Load will attempt to load configurations from a file on a file system.
func (cfg *config) Load(fs afero.Fs) (err error) {
	return config_loader.ConfigLoader(fs, cfg, constants.GetPostgresFileName(), constants.GetPostgresPrefix(), "yaml")
}
