# syntax=docker/dockerfile:1

## README
# Please review the configuration files in the /configs directory and adjust configurations
# accordingly.


## Build
FROM golang:1.20-alpine AS build

RUN apk update && apk --no-cache --update add make

WORKDIR /build

COPY . ./
RUN make dep
RUN make build ARCH=linux

## Deploy
FROM alpine:latest

# Variables
ENV execName=FTeX-linux
ENV sopsDir=/etc/FTeX.conf/

# Install required packages.
RUN apk --no-cache --update add age rpm bash

# Manually install SOpS and perfom cleanup.
RUN wget https://github.com/getsops/sops/releases/download/v3.7.3/sops-3.7.3-1.x86_64.rpm
RUN rpm -i sops-3.7.3-1.x86_64.rpm && rm sops-3.7.3-1.x86_64.rpm

WORKDIR /

# Compiled binary.
COPY --from=build /build/$execName /$execName

# Service Configurations.
COPY --from=build /build/configs/*.sops $sopsDir

# Copy over decryption script.
COPY --from=build /build/docker/bootstrap.sh bootstrap.sh
RUN chmod +x bootstrap.sh

# Port list:
# Please set these ports according to the configurations in the YAML files in /configs directory.
# 1) HTTP REST
# 2) HTTP GraphQL
EXPOSE 33723 47130

# Run Gin Web Framework in production mode.
ENV GIN_MODE=release

# Launch application.
ENTRYPOINT ./bootstrap.sh" "$sopsDir" "$execName"
