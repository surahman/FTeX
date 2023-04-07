# Redis

Configuration loading is designed for containerization in mind. The container engine and orchestrator can mount volumes
(secret or regular) as well as set the environment variables as outlined below.

You may set configurations through both files and environment variables. Please note that environment variables will
override the settings in the configuration files. The configuration files are all expected to be in `YAML` format.

<br/>

## Table of contents

- [Case Study and Justification](#case-study-and-justification)
    - [File Location(s)](#file-locations)
    - [Configuration File](#configuration-file)
        - [Example Configuration File](#example-configuration-file)
        - [Example Environment Variables](#example-environment-variables)

<br/>

## Case Study and Justification

To reduce latency and API calls to a third-party data provider a cache must be used. It is assumed for this demonstration
application that price quotes update at evenly spaced intervals over an hour (i.e. every fifteen minutes or four times an
hour).

There is a risk of a cache stampede where all keys get evicted at the same time. Fiat currencies are finite in number (~197)
and it is extremely likely that there are just one or two dozen prevailing ones. There are many Cryptographic currencies
but only a few are assumed to be popular. Both of these situations mean that API calls for price quotes will likely be limited.
Any evictions from the cache will be easily refilled without overloading the third-party API.

Redis is ideal for the use case here:
* Highly performant cache.
* Built-in data replication for high-availability across a cache cluster.
* On-disk persistence protects against cold cache scenarios which may see the backend database get overwhelmed with
  requests in the event of a cache failure.
* Automatic failover.
* Keys are evicted using an LRU policy.
* Keys can have an expiration time set via a time-to-live.

Cache Policy:
* Quotes will be evicted from the cache at the following intervals, plus a random delta of a few seconds above.
  * hour + 00 minutes
  * hour + 15 minutes
  * hour + 30 minutes
  * hour + 45 minutes

<br/>

### File Location(s)

The configuration loader will search for the configurations in the following order:

| Location                 | Details                                                                                                |
|--------------------------|--------------------------------------------------------------------------------------------------------|
| `/etc/MCQPlatform.conf/` | The `etc` directory is the canonical location for configurations.                                      |
| `$HOME/.MCQPlatform/`    | Configurations can be located in the user's home directory.                                            |
| `./configs/`             | The config folder in the root directory where the application is located.                              |
| Environment variables    | Finally, the configurations will be loaded from environment variables and override configuration files |

### Configuration File

The expected file name is `RedisConfig.yaml`. Unless otherwise specified, all the configuration items below are _required_.

| Name                 | Environment Variable Key | Type   | Description                                                                                       |
|----------------------|--------------------------|--------|---------------------------------------------------------------------------------------------------|
| **_Authentication_** | `REDIS_AUTHENTICATION`   |        | **_Parent key for authentication information._**                                                  |
| ↳ username           | ↳ `.USERNAME`            | string | Username for Redis session login.                                                                 |
| ↳ password           | ↳ `.PASSWORD`            | string | Password for Redis session login.                                                                 |
| **_Connection_**     | `REDIS_CONNECTION`       |        | **_Parent key for connection configuration._**                                                    |
| ↳ addr               | ↳ `.ADDR`                | string | The cluster IPs to bootstrap the connection. Must contain the port number.                        |
| ↳ maxConnAttempts    | ↳ `.MAXCONNATTEMPTS`     | int    | The maximum number of times to try to establish a connection.                                     |
| ↳ maxRetries         | ↳ `.MAXRETRIES`          | int    | The maximum number of times to try an operation.                                                  |
| ↳ poolSize           | ↳ `.POOLSIZE`            | int    | The connection pool size on a per cluster basis.                                                  |
| ↳ minIdleConns       | ↳ `.MINIDLECONNS`        | int    | The number of minimum idle connections per client.                                                |
| ↳ maxIdleConns       | ↳ `.MAXIDLECONNS`        | int    | The maximum number idle connections per client.                                                   |
| **_Data_**           | `REDIS_DATA`             |        | **_Parent key for data configuration._**                                                          |
| ↳ ttl                | ↳ `.TTL`                 | int    | The maximum time in seconds tha an item can remain in the cache before it is evicted. _Optional._ |

#### Example Configuration File

```yaml
authentication:
  username: ftex_service
  password: ZoF1bncLLyYT1agKfWQY
connection:
  addr: 127.0.0.1:7379
  maxConnAttempts: 5
  maxRetries: 3
  poolSize: 4
  minIdleConns: 1
  maxIdleConns: 20
data:
  ttl: 900
```

#### Example Environment Variables

```bash
export REDIS_AUTHENTICATION.PASSWORD=root
export REDIS_CONNECTION.TTL=28800
```
