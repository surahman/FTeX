package logger

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/configloader"
	"github.com/surahman/FTeX/pkg/constants"
)

// config contains the configurations loaded from the configuration file.
//
//nolint:lll
type config struct {
	BuiltinConfig        string         `json:"builtinConfig,omitempty"        mapstructure:"builtinConfig"        validate:"oneof='Production' 'production' 'Development' 'development'" yaml:"builtinConfig,omitempty"`
	BuiltinEncoderConfig string         `json:"builtinEncoderConfig,omitempty" mapstructure:"builtinEncoderConfig" validate:"oneof='Production' 'production' 'Development' 'development'" yaml:"builtinEncoderConfig,omitempty"`
	GeneralConfig        *generalConfig `json:"generalConfig,omitempty"        mapstructure:"generalConfig"        yaml:"generalConfig,omitempty"`
	EncoderConfig        *encoderConfig `json:"encoderConfig,omitempty"        mapstructure:"encoderConfig"        yaml:"encoderConfig,omitempty"`
}

// generalConfig contains all the general logger configurations.
//
//nolint:lll
type generalConfig struct {
	Development       bool     `json:"development"       mapstructure:"development"       validate:"required" yaml:"development"`
	DisableCaller     bool     `json:"disableCaller"     mapstructure:"disableCaller"     validate:"required" yaml:"disableCaller"`
	DisableStacktrace bool     `json:"disableStacktrace" mapstructure:"disableStacktrace" validate:"required" yaml:"disableStacktrace"`
	Encoding          string   `json:"encoding"          mapstructure:"encoding"          validate:"required" yaml:"encoding"`
	OutputPaths       []string `json:"outputPaths"       mapstructure:"outputPaths"       validate:"required" yaml:"outputPaths"`
	ErrorOutputPaths  []string `json:"errorOutputPaths"  mapstructure:"errorOutputPaths"  validate:"required" yaml:"errorOutputPaths"`
}

// encoderConfig contains all the log encoder configurations.
//
//nolint:lll
type encoderConfig struct {
	MessageKey       string `json:"messageKey"       mapstructure:"messageKey"       validate:"required" yaml:"messageKey"`
	LevelKey         string `json:"levelKey"         mapstructure:"levelKey"         validate:"required" yaml:"levelKey"`
	TimeKey          string `json:"timeKey"          mapstructure:"timeKey"          validate:"required" yaml:"timeKey"`
	NameKey          string `json:"nameKey"          mapstructure:"nameKey"          validate:"required" yaml:"nameKey"`
	CallerKey        string `json:"callerKey"        mapstructure:"callerKey"        validate:"required" yaml:"callerKey"`
	FunctionKey      string `json:"functionKey"      mapstructure:"functionKey"      validate:"required" yaml:"functionKey"`
	StacktraceKey    string `json:"stacktraceKey"    mapstructure:"stacktraceKey"    validate:"required" yaml:"stacktraceKey"`
	SkipLineEnding   bool   `json:"skipLineEnding"   mapstructure:"skipLineEnding"   validate:"required" yaml:"skipLineEnding"`
	LineEnding       string `json:"lineEnding"       mapstructure:"lineEnding"       validate:"required" yaml:"lineEnding"`
	ConsoleSeparator string `json:"consoleSeparator" mapstructure:"consoleSeparator" validate:"required" yaml:"consoleSeparator"`
}

// newConfig creates a blank configuration struct for the Zap Logger.
func newConfig() config {
	return config{}
}

// Load will attempt to load configurations from a file on a file system.
func (cfg *config) Load(fs afero.Fs) error {
	if err := configloader.Load(
		fs,
		cfg,
		constants.LoggerFileName(),
		constants.LoggerPrefix(),
		"yaml"); err != nil {
		return fmt.Errorf("zap logger config loading failed: %w", err)
	}

	return nil
}
