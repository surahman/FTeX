# syntax=docker/dockerfile:1

## README
# Please review the configuration files in the /configs directory and adjust configurations
# accordingly.


## Build
FROM golang:1.18-alpine AS build

RUN apk update && apk --no-cache --update add make

WORKDIR /build

COPY . ./
RUN make dep
RUN make build ARCH=linux

## Deploy
FROM alpine:latest

WORKDIR /

# Compiled binary.
COPY --from=build /build/FTeX-linux /FTeX-linux

# Service Configurations.
COPY --from=build /build/configs /etc/FTeX.conf/

# Port list:
# Please set these ports according to the configurations in the YAML files in /configs directory.
# 1) HTTP REST
# 2) HTTP GraphQL
EXPOSE 44243 44255

# Run Gin Web Framework in production mode.
ENV GIN_MODE=release

# Launch application.
ENTRYPOINT ["./FTeX-linux"]
