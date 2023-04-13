package quotes

// configTestData will return a map of test data containing valid and invalid http quotes client configs.
func configTestData() map[string]string {
	return map[string]string{
		"empty": ``,

		"valid": `
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
  timeout: 1s`,

		"no fiat api key": `
fiatCurrency:
  apiKey:
  headerKey: X-RapidAPI-Key
  endpoint: https://api.apilayer.com/exchangerates_data/convert?
cryptoCurrency:
  apiKey: some-api-key-for-crypto-currencies
  headerKey: X-CoinAPI-Key
  endpoint: url-to-data-source-for-crypto-currency-price-quotes
connection:
  userAgent: ftex_inc
  timeout: 1s`,

		"no fiat header key": `
fiatCurrency:
  apiKey: some-api-key-for-fiat-currencies
  headerKey:
  endpoint: https://api.apilayer.com/exchangerates_data/convert?
cryptoCurrency:
  apiKey: some-api-key-for-crypto-currencies
  headerKey: X-CoinAPI-Key
  endpoint: url-to-data-source-for-crypto-currency-price-quotes
connection:
  userAgent: ftex_inc
  timeout: 1s
`,

		"no fiat api endpoint": `
fiatCurrency:
  apiKey: some-api-key-for-fiat-currencies
  headerKey: X-RapidAPI-Key
  endpoint:
cryptoCurrency:
  apiKey: some-api-key-for-crypto-currencies
  headerKey: X-CoinAPI-Key
  endpoint: url-to-data-source-for-crypto-currency-price-quotes
connection:
  userAgent: ftex_inc
  timeout: 1s`,

		"no fiat": `
cryptoCurrency:
  apiKey: some-api-key-for-crypto-currencies
  headerKey: X-CoinAPI-Key
  endpoint: url-to-data-source-for-crypto-currency-price-quotes
connection:
  userAgent: ftex_inc
  timeout: 1s`,

		"no crypto api key": `
fiatCurrency:
  apiKey: some-api-key-for-fiat-currencies
  headerKey: X-RapidAPI-Key
  endpoint: https://api.apilayer.com/exchangerates_data/convert?
cryptoCurrency:
  apiKey:
  headerKey: X-CoinAPI-Key
  endpoint: url-to-data-source-for-crypto-currency-price-quotes
connection:
  userAgent: ftex_inc
  timeout: 1s`,

		"no crypto header key": `
fiatCurrency:
  apiKey: some-api-key-for-fiat-currencies
  headerKey: X-RapidAPI-Key
  endpoint: https://api.apilayer.com/exchangerates_data/convert?
cryptoCurrency:
  apiKey: some-api-key-for-crypto-currencies
  headerKey:
  endpoint: url-to-data-source-for-crypto-currency-price-quotes
connection:
  userAgent: ftex_inc
  timeout: 1s`,

		"no crypto api endpoint": `
fiatCurrency:
  apiKey: some-api-key-for-fiat-currencies
  headerKey: X-RapidAPI-Key
  endpoint: https://api.apilayer.com/exchangerates_data/convert?
cryptoCurrency:
  apiKey: some-api-key-for-crypto-currencies
  headerKey: X-CoinAPI-Key
  endpoint:
connection:
  userAgent: ftex_inc
  timeout: 1s`,

		"no crypto": `
fiatCurrency:
  apiKey: some-api-key-for-fiat-currencies
  headerKey: X-RapidAPI-Key
  endpoint: https://api.apilayer.com/exchangerates_data/convert?
connection:
  userAgent: ftex_inc
  timeout: 1s`,

		"no connection user-agent": `
fiatCurrency:
  apiKey: some-api-key-for-fiat-currencies
  headerKey: X-RapidAPI-Key
  endpoint: https://api.apilayer.com/exchangerates_data/convert?
cryptoCurrency:
  apiKey: some-api-key-for-crypto-currencies
  headerKey: X-CoinAPI-Key
  endpoint: url-to-data-source-for-crypto-currency-price-quotes
connection:
  userAgent:
  timeout: 1s`,

		"no connection timeout": `
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
  timeout:`,

		"no connection": `
fiatCurrency:
  apiKey: some-api-key-for-fiat-currencies
  headerKey: X-RapidAPI-Key
  endpoint: https://api.apilayer.com/exchangerates_data/convert?
cryptoCurrency:
  apiKey: some-api-key-for-crypto-currencies
  headerKey: X-CoinAPI-Key
  endpoint: url-to-data-source-for-crypto-currency-price-quotes`,
	}
}
