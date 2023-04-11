package quotes

// configTestData will return a map of test data containing valid and invalid http quotes client configs.
func configTestData() map[string]string {
	return map[string]string{
		"empty": ``,

		"valid": `
fiatCurrency:
  apiKey: some-api-key-for-fiat-currencies
  endpoint: url-to-data-source-for-fiat-currency-price-quotes
cryptoCurrency:
  apiKey: some-api-key-for-crypto-currencies
  endpoint: url-to-data-source-for-crypto-currency-price-quotes
connection:
  userAgent: ftex_inc
  timeout: 1s`,

		"no fiat api key": `
fiatCurrency:
  apiKey:
  endpoint: url-to-data-source-for-fiat-currency-price-quotes
cryptoCurrency:
  apiKey: some-api-key-for-crypto-currencies
  endpoint: url-to-data-source-for-crypto-currency-price-quotes
connection:
  userAgent: ftex_inc
  timeout: 1s`,

		"no fiat api endpoint": `
fiatCurrency:
  apiKey: some-api-key-for-fiat-currencies
  endpoint:
cryptoCurrency:
  apiKey: some-api-key-for-crypto-currencies
  endpoint: url-to-data-source-for-crypto-currency-price-quotes
connection:
  userAgent: ftex_inc
  timeout: 1s`,

		"no fiat": `
cryptoCurrency:
  apiKey: some-api-key-for-crypto-currencies
  endpoint: url-to-data-source-for-crypto-currency-price-quotes
connection:
  userAgent: ftex_inc
  timeout: 1s`,

		"no crypto api key": `
fiatCurrency:
  apiKey: some-api-key-for-fiat-currencies
  endpoint: url-to-data-source-for-fiat-currency-price-quotes
cryptoCurrency:
  apiKey:
  endpoint: url-to-data-source-for-crypto-currency-price-quotes
connection:
  userAgent: ftex_inc
  timeout: 1s`,

		"no crypto api endpoint": `
fiatCurrency:
  apiKey: some-api-key-for-fiat-currencies
  endpoint: url-to-data-source-for-fiat-currency-price-quotes
cryptoCurrency:
  apiKey: some-api-key-for-crypto-currencies
  endpoint:
connection:
  userAgent: ftex_inc
  timeout: 1s`,

		"no crypto": `
fiatCurrency:
  apiKey: some-api-key-for-fiat-currencies
  endpoint: url-to-data-source-for-fiat-currency-price-quotes
cryptoCurrency:
  apiKey: some-api-key-for-crypto-currencies
  endpoint: url-to-data-source-for-crypto-currency-price-quotes`,

		"no connection user-agent": `
fiatCurrency:
  apiKey: some-api-key-for-fiat-currencies
  endpoint: url-to-data-source-for-fiat-currency-price-quotes
cryptoCurrency:
  apiKey: some-api-key-for-crypto-currencies
  endpoint: url-to-data-source-for-crypto-currency-price-quotes
connection:
  userAgent:
  timeout: 1s`,

		"no connection timeout": `
fiatCurrency:
  apiKey: some-api-key-for-fiat-currencies
  endpoint: url-to-data-source-for-fiat-currency-price-quotes
cryptoCurrency:
  apiKey: some-api-key-for-crypto-currencies
  endpoint: url-to-data-source-for-crypto-currency-price-quotes
connection:
  userAgent: ftex_inc
  timeout:`,

		"no connection": `
fiatCurrency:
  apiKey: some-api-key-for-fiat-currencies
  endpoint: url-to-data-source-for-fiat-currency-price-quotes
cryptoCurrency:
  apiKey: some-api-key-for-crypto-currencies
  endpoint: url-to-data-source-for-crypto-currency-price-quotes`,
	}
}
