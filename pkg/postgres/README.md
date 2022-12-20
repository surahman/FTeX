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
| `/etc/FTeXconf/`      | The `etc` directory is the canonical location for configurations.                                      |
| `$HOME/.FTeX/`        | Configurations can be located in the user's home directory.                                            |
| `./configs/`          | The config folder in the root directory where the application is located.                              |
| Environment variables | Finally, the configurations will be loaded from environment variables and override configuration files |

### Configuration File

The expected file name is `PostgresConfig.yaml`. All the configuration items below are _required_ unless specified otherwise.

| Name                      | Environment Variable Key     | Type          | Description                                                                               |
|---------------------------|------------------------------|---------------|-------------------------------------------------------------------------------------------|
| **_Authentication_**      | `POSTGRES_AUTHENTICATION`    |               | **_Parent key for authentication information._**                                          |
| ↳ username                | ↳ `.USERNAME`                | string        | Username for Postgres session login.                                                      |
| ↳ password                | ↳ `.PASSWORD`                | string        | Password for Postgres session login.                                                      |
| **_Connection_**          | `POSTGRES_CONNECTION`        |               | **_Parent key for connection information._**                                              |
| ↳ database                | ↳ `.DATABASE`                | string        | Database name.                                                                            |
| ↳ host                    | ↳ `.HOST`                    | string        | Hostname or IP address.                                                                   |
| ↳ max_connection_attempts | ↳ `.MAX_CONNECTION_ATTEMPTS` | int           | Number of times to attempt a connection to the database using Binary Exponential Backoff. |
| ↳ port                    | ↳ `.PORT`                    | uint16        | Host port.                                                                                |
| ↳ timeout                 | ↳ `.TIMEOUT`                 | uint16        | Connection timeout in seconds.                                                            |
| ↳ ssl_enabled             | ↳ `.SSL_ENABLED`             | bool          | Connection SSL enabled. _Optional_.                                                       |
| **_Pool_**                | `POSTGRES_POOL`              |               | **_Parent key for connection pool information._**                                         |
| ↳ health_check_period     | ↳ `.HEALTH_CHECK_PERIOD`     | time.Duration | Seconds (min=5) between health checks for each connection in the pool.                    |
| ↳ max_conns               | ↳ `.MAX_CONNS`               | int32         | Maximum connections (min=4) to retain in the connection pool.                             |
| ↳ min_conns               | ↳ `.MIN_CONNS`               | int32         | Minimum connections (min=4) to retain in the connection pool.                             |
| ↳ lazy_connect            | ↳ `.LAZY_CONNECT`            | bool          | Establish a connection only after an I/O request. _Optional_.                             |

 (min=5)
#### Example Configuration File

```yaml
authentication:
  username: postgres
  password: postgres
connection:
  database: ft-ex-db
  host: 127.0.0.1
  port: 6432
  timeout: 5
  ssl_enabled: false
pool:
  health_check_period: 30s
  max_conns: 8
  min_conns: 4
  lazy_connect: false
```

#### Example Environment Variables

```bash
export POSTGRES_AUTHENTICATION.USERNAME=admin
export POSTGRES_AUTHENTICATION.PASSWORD=root
```

### Design Documentation
The technology selection case study and table schema design documentation can be found [here](../model/postgres).
