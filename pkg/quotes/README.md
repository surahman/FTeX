# Quotes

Configuration loading is designed for containerization in mind. The container engine and orchestrator can mount volumes
(secret or regular) as well as set the environment variables as outlined below.

You may set configurations through both files and environment variables. Please note that environment variables will
override the settings in the configuration files. The configuration files are all expected to be in `YAML` format.

<br/>

## Table of contents

- [Price Quote Providers](#price-quote-providers)
    - [File Location(s)](#file-locations)
    - [Configuration File](#configuration-file)
        - [Example Configuration File](#example-configuration-file)
        - [Example Environment Variables](#example-environment-variables)

<br/>

## Price Quote Providers

Configurations for the API Keys and endpoints for Fiat and Crypto-Currency prices will be set through this file. This
configuration file would be encrypted and stored in a secrets store for a production environment.

The development environment requires that a configuration file called `DevQuotesConfig.yaml` be located in the `configs` directory.
This file will contain the API credentials and endpoint information used in the local test/dev environment. The GitHub Actions CI
pipeline will need to have a secret with the configurations stored under the environment variable `QUOTES_CI_CONFIGS`.

Free API Keys for data can be obtained [here for fiat currencies](https://rapidapi.com/principalapis/api/currency-conversion-and-exchange-rates), and
[here for cryptocurrencies](https://www.coinapi.io/pricing?apikey).

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

The expected file name is `QuotesConfig.yaml`. Unless otherwise specified, all the configuration items below are _required_.

| Name                  | Environment Variable Key | Type          | Description                                                       |
|-----------------------|--------------------------|---------------|-------------------------------------------------------------------|
| **_Fiat Currency_**   | `QUOTES_FIATCURRENCY`    |               | **_Parent key for Fiat Exchange endpoint information._**          |
| ↳ APIKey              | ↳ `.APIKEY`              | string        | API Key for fiat currency quotes.                                 |
| ↳ HeaderKey           | ↳ `.HEADERKEY`           | string        | Header key under which the API Key must be stored.                |
| ↳ Endpoint            | ↳ `.ENDPOINT`            | string        | API endpoint for fiat currency quotes.                            |
| **_Crypto Currency_** | `QUOTES_CRYPTOCURRENCY`  |               | **_Parent key for Crypto Exchange endpoint information._**        |
| ↳ APIKey              | ↳ `.APIKEY`              | string        | API Key for crypto currency quotes.                               |
| ↳ HeaderKey           | ↳ `.HEADERKEY`           | string        | Header key under which the API Key must be stored.                |
| ↳ Endpoint            | ↳ `.ENDPOINT`            | string        | API endpoint for crypto currency quotes.                          |
| **_Connection_**      | `QUOTES_CONNECTION`      |               | **_Parent key for connection configuration._**                    |
| ↳ userAgent           | ↳ `.USERAGENT`           | string        | The user-agent to be used as the request client in http requests. |
| ↳ timeout             | ↳ `.TIMEOUT`             | time.Duration | The maximum duration to wait for a quote request.                 |

#### Example Configuration File

```yaml
fiatCurrency:
  apiKey: some-api-key-for-fiat-currencies
  headerKey: X-RapidAPI-Key
  endpoint: https://api.apilayer.com/exchangerates_data/convert?
cryptoCurrency:
  apiKey: some-api-key-for-crypto-currencies
  headerKey: X-CoinAPI-Key
  endpoint: url-to-data-source-for-crypto-currency-price-quotes
connection:
  userAgent: ftex_inc
  timeout: 1s

```

#### Example Environment Variables

```bash
export QUOTES_FIATCURRENCY.APIKEY=some-api-key
export QUOTES_FIATCURRENCY.ENDPOINT=https://url-to-fiat-currency-endpoint
```
