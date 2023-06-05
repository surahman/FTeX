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
	Authentication authenticationConfig `json:"authentication,omitempty" yaml:"authentication,omitempty" mapstructure:"authentication"`
	Connection     connectionConfig     `json:"connection,omitempty"     yaml:"connection,omitempty"     mapstructure:"connection"`
	Pool           poolConfig           `json:"pool,omitempty"           yaml:"pool,omitempty"           mapstructure:"pool"`
}

// authenticationConfig contains the Postgres session authentication information.
type authenticationConfig struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty" mapstructure:"username" validate:"required"`
	Password string `json:"password,omitempty" yaml:"password,omitempty" mapstructure:"password" validate:"required"`
}

// connectionConfig contains the Postgres session connection information.
//
//nolint:lll
type connectionConfig struct {
	Database        string `json:"database,omitempty"              yaml:"database,omitempty"              mapstructure:"database"              validate:"required"`
	Host            string `json:"host,omitempty"                  yaml:"host,omitempty"                  mapstructure:"host"                  validate:"required"`
	MaxConnAttempts int    `json:"maxConnectionAttempts,omitempty" yaml:"maxConnectionAttempts,omitempty" mapstructure:"maxConnectionAttempts" validate:"required,min=1"`
	Timeout         int    `json:"timeout,omitempty"               yaml:"timeout,omitempty"               mapstructure:"timeout"               validate:"required,min=5"`
	Port            uint16 `json:"port,omitempty"                  yaml:"port,omitempty"                  mapstructure:"port"                  validate:"required"`
}

// poolConfig contains the Postgres session connection pool specific information.
//
//nolint:lll
type poolConfig struct {
	HealthCheckPeriod time.Duration `json:"healthCheckPeriod,omitempty" yaml:"healthCheckPeriod,omitempty" mapstructure:"healthCheckPeriod" validate:"omitempty,min=5s"`
	MaxConns          int32         `json:"maxConns,omitempty"          yaml:"maxConns,omitempty"          mapstructure:"maxConns"          validate:"required,gte=4"`
	MinConns          int32         `json:"minConns,omitempty"          yaml:"minConns,omitempty"          mapstructure:"minConns"          validate:"required,gte=4"`
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
