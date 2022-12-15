package logger

// configTestData will return a map of test data containing valid and invalid logger configs.
func configTestData() map[string]string {
	return map[string]string{
		"empty": ``,

		"valid_devel": `
builtin_config: Development
builtin_encoder_config: Development`,

		"valid_prod": `
builtin_config: Production
builtin_encoder_config: Production`,

		"invalid_builtin": `
builtin_config: Invalid
builtin_encoder_config: Invalid`,

		"valid_config": `
builtin_config: Development
builtin_encoder_config: Development
general_config:
  development: true
  disableCaller: true
  disableStacktrace: true
  encoding: json
  outputPaths: ["stdout", "stderr"]
  errorOutputPaths: ["stdout", "stderr"]
encoder_config:
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
