# HTTP REST API

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
- [Swagger UI](#swagger-ui)

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

The expected file name is `HTTPRESTConfig.yaml`. All the configuration items below are _required_.

| Name                | Environment Variable Key | Type   | Description                                                                                |
|---------------------|--------------------------|--------|--------------------------------------------------------------------------------------------|
| **_Server_**        | `REST_SERVER`            |        | **_Parent key for server configurations._**                                                |
| ↳ portNumber        | ↳ `.PORTNUMBER`          | int    | Service port for inbound and outbound connections.                                         |
| ↳ shutdownDelay     | ↳ `.SHUTDOWNDELAY`       | int    | The number of seconds to wait after a shutdown signal is received to terminate the server. |
| ↳ basePath          | ↳ `.BASEPATH`            | string | The service endpoints base path.                                                           |
| ↳ swaggerPath       | ↳ `.SWAGGERPATH`         | string | The path through which the Swagger UI will be accessible.                                  |
| **_Authorization_** | `REST_AUTHORIZATION`     |        | **_Parent key for authentication configurations._**                                        |
| ↳ headerKey         | ↳ `.HEADERKEY`           | string | The HTTP header key where the authorization token is stored.                               |


#### Example Configuration File

```yaml
server:
  portNumber: 33723
  shutdownDelay: 5
  basePath: api/rest/v1
  swaggerPath: /swagger/*any
authorization:
  headerKey: Authorization
```

#### Example Environment Variables

```bash
export REST_SERVER.PORTNUMBER=33723
export REST_SERVER.BASEPATH=api/rest/v1
```

### Swagger UI
The Swagger UI is accessible through the endpoint URL that is provided in the configurations to view the REST schemas as
well as issue test requests to the endpoints.
