
<p align="center">
  <img src="assets/bitcoin_64px.png" alt="FTeX"/>
</p>

# FTeX

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

## Quotes

Price quotes for Crypto and Fiat currencies are obtained through external third-party providers. The API endpoints used
in this project can be accessed with free accounts. Details can be found in the [`quotes`](pkg/quotes) package.

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

The HTTP endpoint details are located in the [`rest`](pkg/rest) package. The model used for REST API calls can
be found in the [`models`](pkg/models) package.

To review the REST API request and response formats please see the readme in the [`REST handlers`](pkg/rest/handlers)
package. The REST API server does also provide a Swagger UI to examine and test the API calls with details on request
formats.

The Swagger UI can be accessed using the provided default configurations through
[http://localhost:33723/swagger/index.html](http://localhost:33723/swagger/index.html).

### GraphQL

GraphQL has been exposed through an HTTP endpoint [`graphql`](pkg/graphql) package. The schema for the GraphQL queries
and mutations can be found in [`models`](pkg/models) package.

To review the GraphQL API request and response formats please see the readme in the [`graphql_resolvers`](pkg/graph/resolvers)
package. The GraphQL server does also provide a Playground to examine and test the API calls with details on request
formats.

The Playground can be accessed using the provided default configurations through
[http://localhost:47130/api/graphql/v1/playground](http://localhost:47130/api/graphql/v1/playground).

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

To spin-up the Postgres and Redis containers please use the commands below from the project root directory.

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

```bash
docker inspect --format='{{json .State.Health}}' redis
```

Get IP Addresses:

```bash
docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' postgres
```

```bash
docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' redis
```


**Postgres:**

- Username : `postgres`
- Password : `postgres`
- Port : `6432`
- Database : `ftex_db`

**Redis:**

- Username : `default`
- Password : `v7h0JZq8AZhVcF3NWz9u`

- Username : `ftex_service`
- Password : `ZoF1bncLLyYT1agKfWQY`

- Port : `7379`
- Database Fiat: `0`
- Database Crypto: `1`

<br/>

[Crypto icons created by Freepik - Flaticon](https://www.flaticon.com/free-icons/crypto)
