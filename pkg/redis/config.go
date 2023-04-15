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
	Authentication authenticationConfig `json:"authentication,omitempty" yaml:"authentication,omitempty" mapstructure:"authentication"`
	Connection     connectionConfig     `json:"connection,omitempty" yaml:"connection,omitempty" mapstructure:"connection"`
	Data           dataConfig           `json:"data,omitempty" yaml:"data,omitempty" mapstructure:"data"`
}

// authenticationConfig contains the Redis session authentication information.
type authenticationConfig struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty" mapstructure:"username" validate:"required"`
	Password string `json:"password,omitempty" yaml:"password,omitempty" mapstructure:"password"`
}

// connectionConfig contains the Redis session connection information.
//
//nolint:lll
type connectionConfig struct {
	Addr            string `json:"addr,omitempty" yaml:"addr,omitempty" mapstructure:"addr" validate:"required"`
	MaxConnAttempts int    `json:"maxConnAttempts,omitempty" yaml:"maxConnAttempts,omitempty" mapstructure:"maxConnAttempts" validate:"required,min=1"`
	MaxRetries      int    `json:"maxRetries,omitempty" yaml:"maxRetries,omitempty" mapstructure:"maxRetries" validate:"required,min=1"`
	PoolSize        int    `json:"poolSize,omitempty" yaml:"poolSize,omitempty" mapstructure:"poolSize" validate:"required,min=1"`
	MinIdleConns    int    `json:"minIdleConns,omitempty" yaml:"minIdleConns,omitempty" mapstructure:"minIdleConns" validate:"required,min=1"`
	MaxIdleConns    int    `json:"maxIdleConns,omitempty" yaml:"maxIdleConns,omitempty" mapstructure:"maxIdleConns"`
}

// dataConfig contains the Redis data storage related information.
type dataConfig struct {
	TTL int64 `json:"ttl,omitempty" yaml:"ttl,omitempty" mapstructure:"ttl" validate:"omitempty,min=60"`
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
		constants.GetRedisFileName(),
		constants.GetRedisPrefix(),
		"yaml"); err != nil {
		return fmt.Errorf("redis config loading failed: %w", err)
	}

	return nil
}
