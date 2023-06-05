package postgres

import (
	"fmt"
	"time"

	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/configloader"
	"github.com/surahman/FTeX/pkg/constants"
)

// config contains the configurations loaded from the configuration file.
//
//nolint:lll
type config struct {
	Authentication authenticationConfig `json:"authentication,omitempty" mapstructure:"authentication" yaml:"authentication,omitempty"`
	Connection     connectionConfig     `json:"connection,omitempty"     mapstructure:"connection"     yaml:"connection,omitempty"`
	Pool           poolConfig           `json:"pool,omitempty"           mapstructure:"pool"           yaml:"pool,omitempty"`
}

// authenticationConfig contains the Postgres session authentication information.
type authenticationConfig struct {
	Username string `json:"username,omitempty" mapstructure:"username" validate:"required" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" mapstructure:"password" validate:"required" yaml:"password,omitempty"`
}

// connectionConfig contains the Postgres session connection information.
//
//nolint:lll
type connectionConfig struct {
	Database        string `json:"database,omitempty"              mapstructure:"database"              validate:"required"       yaml:"database,omitempty"`
	Host            string `json:"host,omitempty"                  mapstructure:"host"                  validate:"required"       yaml:"host,omitempty"`
	MaxConnAttempts int    `json:"maxConnectionAttempts,omitempty" mapstructure:"maxConnectionAttempts" validate:"required,min=1" yaml:"maxConnectionAttempts,omitempty"`
	Timeout         int    `json:"timeout,omitempty"               mapstructure:"timeout"               validate:"required,min=5" yaml:"timeout,omitempty"`
	Port            uint16 `json:"port,omitempty"                  mapstructure:"port"                  validate:"required"       yaml:"port,omitempty"`
}

// poolConfig contains the Postgres session connection pool specific information.
//
//nolint:lll
type poolConfig struct {
	HealthCheckPeriod time.Duration `json:"healthCheckPeriod,omitempty" mapstructure:"healthCheckPeriod" validate:"omitempty,min=5s" yaml:"healthCheckPeriod,omitempty"`
	MaxConns          int32         `json:"maxConns,omitempty"          mapstructure:"maxConns"          validate:"required,gte=4"   yaml:"maxConns,omitempty"`
	MinConns          int32         `json:"minConns,omitempty"          mapstructure:"minConns"          validate:"required,gte=4"   yaml:"minConns,omitempty"`
}

// newConfig creates a blank configuration struct for Postgres.
func newConfig() config {
	return config{}
}

// Load will attempt to load configurations from a file on a file system.
func (cfg *config) Load(fs afero.Fs) error {
	if err := configloader.Load(
		fs,
		cfg,
		constants.GetPostgresFileName(),
		constants.GetPostgresPrefix(),
		"yaml"); err != nil {
		return fmt.Errorf("postgres config loading failed: %w", err)
	}

	return nil
}
