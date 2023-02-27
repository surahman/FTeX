package logger

import (
	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/configloader"
	"github.com/surahman/FTeX/pkg/constants"
)

// config contains the configurations loaded from the configuration file.
//
//nolint:lll
type config struct {
	BuiltinConfig        string         `json:"builtin_config,omitempty" yaml:"builtin_config,omitempty" mapstructure:"builtin_config" validate:"oneof='Production' 'production' 'Development' 'development'"`
	BuiltinEncoderConfig string         `json:"builtin_encoder_config,omitempty" yaml:"builtin_encoder_config,omitempty" mapstructure:"builtin_encoder_config" validate:"oneof='Production' 'production' 'Development' 'development'"`
	GeneralConfig        *generalConfig `json:"general_config,omitempty" yaml:"general_config,omitempty" mapstructure:"general_config"`
	EncoderConfig        *encoderConfig `json:"encoder_config,omitempty" yaml:"encoder_config,omitempty" mapstructure:"encoder_config"`
}

// generalConfig contains all the general logger configurations.
//
//nolint:lll
type generalConfig struct {
	Development       bool     `json:"development" yaml:"development" mapstructure:"development" validate:"required"`
	DisableCaller     bool     `json:"disableCaller" yaml:"disableCaller" mapstructure:"disableCaller" validate:"required"`
	DisableStacktrace bool     `json:"disableStacktrace" yaml:"disableStacktrace" mapstructure:"disableStacktrace" validate:"required"`
	Encoding          string   `json:"encoding" yaml:"encoding" mapstructure:"encoding" validate:"required"`
	OutputPaths       []string `json:"outputPaths" yaml:"outputPaths" mapstructure:"outputPaths" validate:"required"`
	ErrorOutputPaths  []string `json:"errorOutputPaths" yaml:"errorOutputPaths" mapstructure:"errorOutputPaths" validate:"required"`
}

// encoderConfig contains all the log encoder configurations.
//
//nolint:lll
type encoderConfig struct {
	MessageKey       string `json:"messageKey" yaml:"messageKey" mapstructure:"messageKey" validate:"required"`
	LevelKey         string `json:"levelKey" yaml:"levelKey" mapstructure:"levelKey" validate:"required"`
	TimeKey          string `json:"timeKey" yaml:"timeKey" mapstructure:"timeKey" validate:"required"`
	NameKey          string `json:"nameKey" yaml:"nameKey" mapstructure:"nameKey" validate:"required"`
	CallerKey        string `json:"callerKey" yaml:"callerKey" mapstructure:"callerKey" validate:"required"`
	FunctionKey      string `json:"functionKey" yaml:"functionKey" mapstructure:"functionKey" validate:"required"`
	StacktraceKey    string `json:"stacktraceKey" yaml:"stacktraceKey" mapstructure:"stacktraceKey" validate:"required"`
	SkipLineEnding   bool   `json:"skipLineEnding" yaml:"skipLineEnding" mapstructure:"skipLineEnding" validate:"required"`
	LineEnding       string `json:"lineEnding" yaml:"lineEnding" mapstructure:"lineEnding" validate:"required"`
	ConsoleSeparator string `json:"consoleSeparator" yaml:"consoleSeparator" mapstructure:"consoleSeparator" validate:"required"`
}

// newConfig creates a blank configuration struct for the Zap Logger.
func newConfig() config {
	return config{}
}

// Load will attempt to load configurations from a file on a file system.
func (cfg *config) Load(fs afero.Fs) (err error) {
	return configloader.Load(fs, cfg, constants.GetLoggerFileName(), constants.GetLoggerPrefix(), "yaml")
}
