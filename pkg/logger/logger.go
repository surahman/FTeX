package logger

import (
	"errors"
	"log"
	"strings"

	"github.com/spf13/afero"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

// Logger is the Zap logger object.
type Logger struct {
	zapLogger *zap.Logger
}

// NewLogger will create a new uninitialized logger.
func NewLogger() *Logger {
	return &Logger{}
}

// Init will initialize the logger with configurations and start it.
func (l *Logger) Init(fs *afero.Fs) error {
	if l.zapLogger != nil {
		return errors.New("logger is already initialized")
	}

	var (
		err        error
		baseConfig zap.Config
		encConfig  zapcore.EncoderConfig
	)

	userConfig := newConfig()
	if err = userConfig.Load(*fs); err != nil {
		log.Printf("failed to load logger configuration file from disk: %v\n", err)

		return err
	}

	// Base logger configuration.
	switch strings.ToLower(userConfig.BuiltinConfig) {
	case "development":
		baseConfig = zap.NewDevelopmentConfig()
	case "production":
		baseConfig = zap.NewProductionConfig()
	default:
		msg := "could not select the base configuration type"
		log.Println(msg)

		return errors.New(msg)
	}

	// Encoder configuration.
	switch strings.ToLower(userConfig.BuiltinEncoderConfig) {
	case "development":
		encConfig = zap.NewDevelopmentEncoderConfig()
	case "production":
		encConfig = zap.NewProductionEncoderConfig()
	default:
		msg := "could not select the base encoder type"
		log.Println(msg)

		return errors.New(msg)
	}

	// Merge configurations.
	if err = mergeConfig[*zap.Config, *generalConfig](&baseConfig, userConfig.GeneralConfig); err != nil {
		log.Printf("failed to merge base configurations and user provided configurations for logger: %v\n", err)

		return err
	}

	if err = mergeConfig[*zapcore.EncoderConfig, *encoderConfig](&encConfig, userConfig.EncoderConfig); err != nil {
		log.Printf("failed to merge base and user provided encoder configurations for logger: %v\n", err)

		return err
	}

	// Init and create logger.
	baseConfig.EncoderConfig = encConfig
	if l.zapLogger, err = baseConfig.Build(zap.AddCallerSkip(1)); err != nil {
		log.Printf("failure configuring logger: %v\n", err)

		return err
	}

	return nil
}

// Info logs messages at the info level.
func (l *Logger) Info(message string, fields ...zap.Field) {
	l.zapLogger.Info(message, fields...)
}

// Debug logs messages at the debug level.
func (l *Logger) Debug(message string, fields ...zap.Field) {
	l.zapLogger.Debug(message, fields...)
}

// Warn logs messages at the warn level.
func (l *Logger) Warn(message string, fields ...zap.Field) {
	l.zapLogger.Warn(message, fields...)
}

// Error logs messages at the error level.
func (l *Logger) Error(message string, fields ...zap.Field) {
	l.zapLogger.Error(message, fields...)
}

// Panic logs messages at the panic level and then panics at the call site.
func (l *Logger) Panic(message string, fields ...zap.Field) {
	l.zapLogger.Panic(message, fields...)
}

// mergeConfig will merge the configuration files by marshalling and unmarshalling.
//
//nolint:lll
func mergeConfig[DST *zap.Config | *zapcore.EncoderConfig, SRC *generalConfig | *encoderConfig](dst DST, src SRC) (err error) {
	var yamlToConv []byte

	if yamlToConv, err = yaml.Marshal(src); err != nil {
		return
	}

	if err = yaml.Unmarshal(yamlToConv, dst); err != nil {
		return
	}

	return
}

// setTestLogger is a utility method that set a logger base for testing.
func (l *Logger) setTestLogger(testLogger *zap.Logger) {
	l.zapLogger = testLogger
}

// NewTestLogger will create a new development logger to be used in test suites.
func NewTestLogger() (logger *Logger, err error) {
	baseConfig := zap.NewDevelopmentConfig()
	baseConfig.EncoderConfig = zap.NewDevelopmentEncoderConfig()

	var zapLogger *zap.Logger

	if zapLogger, err = baseConfig.Build(zap.AddCallerSkip(1)); err != nil {
		log.Printf("failure configuring logger: %v\n", err)

		return nil, err
	}

	return &Logger{zapLogger: zapLogger}, err
}
