# Postgres

Configuration loading is designed for containerization in mind. The container engine and orchestrator can mount volumes
(secret or regular) as well as set the environment variables as outlined below.

You may set configurations through both files and environment variables. Please note that environment variables will
override the settings in the configuration files. The configuration files are all expected to be in `YAML` format.

<br/>

## Table of contents

- [File Location(s)](#file-locations)
- [Configuration File](#configuration-file)
    - [Example Configuration File](#example-configuration-file)
    - [Example Environment Variables](#example-environment-variables)
- [Design Documentation](#design-documentation)

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

The expected file name is `PostgresConfig.yaml`. All the configuration items below are _required_ unless specified otherwise.

| Name                    | Environment Variable Key   | Type          | Description                                                                               |
|-------------------------|----------------------------|---------------|-------------------------------------------------------------------------------------------|
| **_Authentication_**    | `POSTGRES_AUTHENTICATION`  |               | **_Parent key for authentication information._**                                          |
| ↳ username              | ↳ `.USERNAME`              | string        | Username for Postgres session login.                                                      |
| ↳ password              | ↳ `.PASSWORD`              | string        | Password for Postgres session login.                                                      |
| **_Connection_**        | `POSTGRES_CONNECTION`      |               | **_Parent key for connection information._**                                              |
| ↳ database              | ↳ `.DATABASE`              | string        | Database name.                                                                            |
| ↳ host                  | ↳ `.HOST`                  | string        | Hostname or IP address.                                                                   |
| ↳ maxConnectionAttempts | ↳ `.MAXCONNECTIONATTEMPTS` | int           | Number of times to attempt a connection to the database using Binary Exponential Backoff. |
| ↳ timeout               | ↳ `.TIMEOUT`               | int           | Connection timeout in seconds.                                                            |
| ↳ port                  | ↳ `.PORT`                  | uint16        | Host port.                                                                                |
| **_Pool_**              | `POSTGRES_POOL`            |               | **_Parent key for connection pool information._**                                         |
| ↳ healthCheckPeriod     | ↳ `.HEALTHCHECKPERIOD`     | time.Duration | Seconds (min=5) between health checks for each connection in the pool.                    |
| ↳ maxConns              | ↳ `.MAXCONNS`              | int32         | Maximum connections (min=4) to retain in the connection pool.                             |
| ↳ minConns              | ↳ `.MINCONNS`              | int32         | Minimum connections (min=4) to retain in the connection pool.                             |


#### Example Configuration File

```yaml
authentication:
  username: postgres
  password: postgres
connection:
  database: ft-ex-db
  host: 127.0.0.1
  maxConnectionAttempts: 5
  port: 6432
  timeout: 5
pool:
  healthCheckPeriod: 30s
  maxConns: 8
  minConns: 4
```

#### Example Environment Variables

```bash
export POSTGRES_AUTHENTICATION.USERNAME=admin
export POSTGRES_AUTHENTICATION.PASSWORD=root
```

### Design Documentation
The technology selection case study and table schema design documentation can be found [here](../../SQL).
