
<p align="center">
  <img src="assets/bitcoin_64px.png" alt="FT-eX"/>
</p>

# FT-eX

This is a demonstration project in `Golang` that provides an API for basic banking of Cryptocurrencies. There is an integration with a cyrpto exchange to get live cryptocurrency prices.

This project will be using an RDBMS (PostreSQL) because of the need for ACID transactions, rollbacks, and row-level locking across tables.

<br/>

:warning: **_Transport Layer Security_** :warning:

Encryption is vital to help safeguard against theft of login credentials and JSON Web Tokens.

In a production environment, `TLS` would be the only HTTP protocol over which the API endpoints would be exposed. Setting
up the `TLS`/`SSL` certificated for a Dockerized demonstration environment is unnecessary and complicates the tester's
experience.

Other methods like envelope encryption of payloads add an extra layer of security, but these add an excessive overhead for
the use case and workloads here.

<br/>

:warning: **_Protocols_** :warning:

This demonstration environment will launch both the `HTTP REST` as well as the `GraphQL` over `HTTP` endpoints. This is
unsuitable for a production environment.

Ideally, each of these protocol endpoints would be exposed in its own clusters with auto-scaling, load balancing, and
across availability zones.

<br/>

## Logging

Configuration information for the logger can be found in the [`logger`](pkg/logger) package.

<br/>

## Authentication

Information regarding authentication configurations can be found in the [`auth`](pkg/auth) package.

<br/>

## HTTP

Details on the HTTP endpoints can be found in their respective packages below.

### REST

The HTTP endpoint details are located in the [`http_rest`](pkg/http/rest) package. The model used for REST API calls can
be found in the [`model_http`](pkg/model/http).

To review the REST API request and response formats please see the readme in the [`http_handlers`](pkg/http/rest/handlers)
package. The REST API server does also provide a Swagger UI to examine and test the API calls with details on request
formats.

The Swagger UI can be accessed using the provided default configurations through
[http://localhost:44243/swagger/index.html](http://localhost:44243/swagger/index.html).

### GraphQL

GraphQL has been exposed through an HTTP endpoint [`graphql`](pkg/http/graph) package. The schema for the GraphQL queries
and mutations can be found in [`model_http`](pkg/model/http).

To review the GraphQL API request and response formats please see the readme in the [`graphql_resolvers`](pkg/http/graph/resolvers)
package. The GraphQL server does also provide a Playground to examine and test the API calls with details on request
formats.

The Playground can be accessed using the provided default configurations through
[http://localhost:44255/api/graphql/v1/playground](http://localhost:44255/api/graphql/v1/playground).

<br/>

# Make Executables

Please provide the `ARCH=` variable with `linux` or `darwin` as needed.

**_Build_**

```bash
make build ARCH=linux
```

**_Clean_**

```bash
make clean
```

<br/>

# Docker Containers

### Microservice Container

To build the container for deployment in a Kubernetes cluster please run the `docker build` command
with the required parameters. Please also review the configuration files in the [configs](configs)
folder and appropriately adjust the ports exposed in the container.

There are port configurations to expose the HTTP REST and GraphQL endpoints. They can be configured
from inside the `Dockerfile` and must match the config `.yaml` files. To expose them, please see the
[`-P`](https://docs.docker.com/engine/reference/commandline/run/#publish-or-expose-port--p---expose)
Docker flag.

When testing using `docker compose` on a local machine you may use the `ifconfig` to obtain your Host IP:

```bash
ifconfig | grep 'inet 192'
```

<br/>

### Data Tier Containers

To spin-up the Postgres container please use the commands below from the project root directory.

Create containers:

```bash
docker compose -f "docker/docker-compose.yaml" up -d
```

Destroy containers:

```bash
docker compose -f "docker/docker-compose.yaml" down
```

List Containers and Check Health:

```bash
docker ps
```

```bash
docker inspect --format='{{json .State.Health}}' postgres
```

Get IP Addresses:

```bash
docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' postgres
```


**Postgres:**
- 
- Username : `postgres`
- Password : `postgres`
- Database : `ft-ex-db`

<br/>

[Crypto icons created by Freepik - Flaticon](https://www.flaticon.com/free-icons/crypto)
