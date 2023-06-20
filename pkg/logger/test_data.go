package logger

// configTestData will return a map of test data containing valid and invalid logger configs.
func configTestData() map[string]string {
	return map[string]string{
		"empty": ``,

		"valid_devel": `
builtinConfig: Development
builtinEncoderConfig: Development`,

		"valid_prod": `
builtinConfig: Production
builtinEncoderConfig: Production`,

		"invalid_builtin": `
builtinConfig: Invalid
builtinEncoderConfig: Invalid`,

		"valid_config": `
builtinConfig: Development
builtinEncoderConfig: Development
generalConfig:
  development: true
  disableCaller: true
  disableStacktrace: true
  encoding: json
  outputPaths: ["stdout", "stderr"]
  errorOutputPaths: ["stdout", "stderr"]
encoderConfig:
  messageKey: message key
  levelKey: level key
  timeKey: time key
  nameKey: name key
  callerKey: caller key
  functionKey: function key
  stacktraceKey: stacktrace key
  skipLineEnding: true
  lineEnding: line ending
  consoleSeparator: console separator`,
	}
}
