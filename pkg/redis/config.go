package redis

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/configloader"
	"github.com/surahman/FTeX/pkg/constants"
)

// config is the configuration container for connecting to the Redis database.
//
//nolint:lll
type config struct {
	Authentication authenticationConfig `json:"authentication,omitempty" mapstructure:"authentication" yaml:"authentication,omitempty"`
	Connection     connectionConfig     `json:"connection,omitempty"     mapstructure:"connection"     yaml:"connection,omitempty"`
}

// authenticationConfig contains the Redis session authentication information.
type authenticationConfig struct {
	Username string `json:"username,omitempty" mapstructure:"username" validate:"required"       yaml:"username,omitempty"`
	Password string `json:"password,omitempty" mapstructure:"password" yaml:"password,omitempty"`
}

// connectionConfig contains the Redis session connection information.
//
//nolint:lll
type connectionConfig struct {
	Addr            string `json:"addr,omitempty"            mapstructure:"addr"            validate:"required"           yaml:"addr,omitempty"`
	MaxConnAttempts int    `json:"maxConnAttempts,omitempty" mapstructure:"maxConnAttempts" validate:"required,min=1"     yaml:"maxConnAttempts,omitempty"`
	MaxRetries      int    `json:"maxRetries,omitempty"      mapstructure:"maxRetries"      validate:"required,min=1"     yaml:"maxRetries,omitempty"`
	PoolSize        int    `json:"poolSize,omitempty"        mapstructure:"poolSize"        validate:"required,min=1"     yaml:"poolSize,omitempty"`
	MinIdleConns    int    `json:"minIdleConns,omitempty"    mapstructure:"minIdleConns"    validate:"required,min=1"     yaml:"minIdleConns,omitempty"`
	MaxIdleConns    int    `json:"maxIdleConns,omitempty"    mapstructure:"maxIdleConns"    yaml:"maxIdleConns,omitempty"`
}

// newConfig creates a blank configuration struct for Redis.
func newConfig() *config {
	return &config{}
}

// Load will attempt to load configurations from a file on a file system.
func (cfg *config) Load(fs afero.Fs) error {
	if err := configloader.Load(
		fs,
		cfg,
		constants.RedisFileName(),
		constants.RedisPrefix(),
		"yaml"); err != nil {
		return fmt.Errorf("redis config loading failed: %w", err)
	}

	return nil
}
