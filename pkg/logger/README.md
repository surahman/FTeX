# Zap Structured Logging

Configuration loading is designed for containerization in mind. The container engine and orchestrator can mount volumes
(secret or regular) as well as set the environment variables as outlined below.

You may set configurations through both files and environment variables. Please note that environment variables will
override the settings in the configuration files. The configuration files are all expected to be in `YAML` format.

<br/>

## Table of contents

- [Zap Structured Logging](#zap-structured-logging)
- [File Location(s)](#file-locations)
- [Configuration File](#configuration-file)
    - [Example Configuration File](#example-configuration-file)
    - [Example Environment Variables](#example-environment-variables)

<br/>

### Zap Structured Logging

Structured logging is handled through [Zap](https://pkg.go.dev/go.uber.org/zap).
Please refer to Zap's documentation on how to set the fields below.

Only the `BuiltinConfig` and `BuiltinEncoder` settings can be provided through environment variables that override values
from the configuration files.

<br/>

### File Location(s)

The configuration loader will search for the configurations in the following order:

| Location              | Details                                                                                                |
|-----------------------|--------------------------------------------------------------------------------------------------------|
| `/etc/FTeX.conf/`     | The `etc` directory is the canonical location for configurations.                                      |
| `$HOME/.FTeX/`        | Configurations can be located in the user's home directory.                                            |
| `./configs/`          | The config folder in the root directory where the application is located.                              |
| Environment variables | Finally, the configurations will be loaded from environment variables and override configuration files |

### Configuration File

The expected file name is `LoggerConfig.yaml`. All the configuration items below are _required_.

**_If specifying settings for the `General Config` or `Encoder Config`, all values must be provided._**

| Name                       | Environment Variable Key      | Type                          | Description                                                                                                 |
|----------------------------|-------------------------------|-------------------------------|-------------------------------------------------------------------------------------------------------------|
| **_BuiltinConfig_**        | `LOGGER_BUILTINCONFIG`        | string                        | Must be one of `Development` or `Production`. Required.                                                     |
| **_BuiltinEncoderConfig_** | `LOGGER_BUILTINENCODERCONFIG` | string                        | Must be one of `Development` or `Production`. Required.                                                     |
| **_General Config_**       | _Not applicable._             | **_General Configurations._** | Please refer to [Zap user documentation](https://pkg.go.dev/go.uber.org/zap#Config).                        |
| ↳ development              |                               | boolean                       |                                                                                                             |
| ↳ disableCaller            |                               | boolean                       |                                                                                                             |
| ↳ disableStacktrace        |                               | boolean                       |                                                                                                             |
| ↳ encoding                 |                               | string                        |                                                                                                             |
| ↳ outputPaths              |                               | string array                  |                                                                                                             |
| ↳ errorOutputPaths         |                               | string array                  |                                                                                                             |
| **_Encoder Config_**       | _Not applicable._             | **_Encoder Configurations._** | Please refer to [Zap user documentation](https://pkg.go.dev/go.uber.org/zap@v1.23.0/zapcore#EncoderConfig). |
| ↳ messageKey               |                               | string                        |
| ↳ levelKey                 |                               | string                        |                                                                                                             |
| ↳ timeKey                  |                               | string                        |                                                                                                             |
| ↳ nameKey                  |                               | string                        |                                                                                                             |
| ↳ callerKey                |                               | string                        |                                                                                                             |
| ↳ functionKey              |                               | string                        |                                                                                                             |
| ↳ stacktraceKey            |                               | string                        |                                                                                                             |
| ↳ skipLineEnding           |                               | boolean                       |                                                                                                             |
| ↳ lineEnding               |                               | string                        |                                                                                                             |
| ↳ consoleSeparator         |                               | string                        |                                                                                                             |

#### Example Configuration File

```yaml
builtinConfig: Development | Production
builtinEncoderConfig: Development | Production
generalConfig:
  development: boolean
  disableCaller: boolean
  disableStacktrace: boolean
  encoding: string
  outputPaths: [string, array]
  errorOutputPaths: [string, array]
encoderConfig:
  messageKey: string
  levelKey: string
  timeKey: string
  nameKey: string
  callerKey: string
  functionKey: string
  stacktraceKey: string
  skipLineEnding: boolean
  lineEnding: string
  consoleSeparator: string
```

#### Example Environment Variables

```bash
export LOGGER_BUILTINCONFIG="Development"
export LOGGER_BUILTINENCODERCONFIG="Production"
```
