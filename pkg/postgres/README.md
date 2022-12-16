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

The expected file name is `PostgresConfig.yaml`. All the configuration items below are _required_.

| Name                      | Environment Variable Key  | Type         | Description                                      |
|---------------------------|---------------------------|--------------|--------------------------------------------------|
| **_Authentication_**      | `POSTGRES_AUTHENTICATION` |              | **_Parent key for authentication information._** |
| ↳ username                | ↳ `.USERNAME`             | string       | Username for Postgres session login.             |
| ↳ password                | ↳ `.PASSWORD`             | string       | Password for Postgres session login.             |


#### Example Configuration File

```yaml
authentication:
  username: admin
  password: root
keyspace:
  name: mcq_platform
  replication_class: SimpleStrategy
  replication_factor: 3
connection:
  consistency: quorum
  cluster_ip: [127.0.0.1]
  proto_version: 4
  timeout: 10
  max_connection_attempts: 5
```

#### Example Environment Variables

```bash
export POSTGRES_AUTHENTICATION.USERNAME=admin
export POSTGRES_AUTHENTICATION.PASSWORD=root
```

### Design Documentation
The technology selection case study and table schema design documentation can be found [here](../model/postgres).
