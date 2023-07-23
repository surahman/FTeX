# Configuration Files

Configuration files for various components within the `FTeX` account transactions microservice are provided in this
directory and are written in the YAML format.

<br/>

## Table of contents


<br/>

### Security Note

Configuration files are provided in this directory in both their plaintext and `age` encrypted Mozilla Secret OPerationS
(`SOPS`) format.

The plaintext files with the `YAML` file extension are provided to facilitate a less complicated review of this
demonstration project. In an actual production system, this would pose a security risk. As such, the configurations have
additionally been included in this directory in the `SOPS` format using the `age` encryption tooling.

<br/>

### age Encryption

A description of [`age`](https://github.com/FiloSottile/age/tree/main), taken directly from their GitHub repository,
describes the project most adequately.

> age is a simple, modern and secure file encryption tool, format, and Go library.
>
> It features small explicit keys, no config options, and UNIX-style composability.

#### Installation

A number of packages for different Operating Systems from various package managers are
[provided](https://github.com/FiloSottile/age/tree/main#installation).

The Alpine `apk`manager is required and is used to install the package used to decrypt the `SOPS` in the Docker
container.
This is achieved using a bash script located in the [Docker](../docker/bootstrap.sh) directory.

#### Key File

:warning: The key file has been provided in this project to allow the decryption of configuration files by reviewers. The key file should not be submitted into a repository. :warning:

To generate a key file, run the following command:

```bash
age-keygen -o age/key.txt
```

The output file will contain the generation timestamp, the public key (used for encryption), and the private/secret key
that is used for decryption:

```text
# created: 2023-07-23T12:08:12-04:00
# public key: age17qltwhv4zxxc8n4rpku8jqpy3mzq37hd02dwtqyp889d23dl7sfskk342t
AGE-SECRET-KEY-1727NPT5T8X5VVTSHRP26U7SEKTV64YJ4CQX6VVQ8DN2R6LGDLYJQPHYJXA
```

<br/>

### Mozilla Secret OPerationS (SOPS)

A description of [`SOPS`](https://github.com/getsops/sops#encrypting-using-age), taken directly from their GitHub
repository, describes the project most adequately.

> sops is an editor of encrypted files that supports YAML, JSON, ENV, INI and BINARY formats and encrypts with AWS KMS,
> GCP KMS, Azure Key Vault, age, and PGP.

Please consult the `SOPS` documentation on working with the format during development.

#### Installation

The installers will need to be downloaded from the [releases](https://github.com/getsops/sops/releases) `SOPS`
GitHub project page. The Docker container will require the `rpm` package manager to be installed during the Alpine build
process.

#### Encrypting

The `age` encryption key must be supplied to the `SOPS` cli tool during the encryption process.

```bash
sops --encrypt --age age17qltwhv4zxxc8n4rpku8jqpy3mzq37hd02dwtqyp889d23dl7sfskk342t configs/AuthConfig.yaml > configs/AuthConfig.sops
```

#### Decrypting

Environment variables need to set that either supply the `age` private/secret key or the path to the actual key file.

```bash
export SOPS_AGE_KEY_FILE=path/to/keyfile/keys.txt
```

```bash
export SOPS_AGE_KEY=AGE-SECRET-KEY-1727NPT5T8X5VVTSHRP26U7SEKTV64YJ4CQX6VVQ8DN2R6LGDLYJQPHYJXA
```

Decrypting the `SOPS` files to plaintext can be achieved using the example command below after the environment variables
above have been configured:

```bash
sops -d --input-type yaml --output-type yaml configs/AuthConfig.sops > configs/AuthConfig.yaml
```

To test the Docker image locally, the `age` key information can be supplied using the `-e` cli flag.

```bash
docker run -d \
-p 33723:33723 \
-p 47130:47130 \
-e POSTGRES_CONNECTION.HOST=192.168.0.211 \
-e REDIS_CONNECTION.ADDR=192.168.0.211:7379 \
-e QUOTES_FIATCURRENCY.APIKEY='some-api-key' \
-e QUOTES_CRYPTOCURRENCY.APIKEY='some-api-key' \
-e SOPS_AGE_KEY=AGE-SECRET-KEY-1727NPT5T8X5VVTSHRP26U7SEKTV64YJ4CQX6VVQ8DN2R6LGDLYJQPHYJXA \
ftex
```
