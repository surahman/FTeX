package graphql

// getHealthcheckQuery is the health check query.
func getHealthcheckQuery() string {
	return `{
		"query": "query { healthcheck() }"
	}`
}

// getUsersQuery is a map of test user mutations and queries.
//
//nolint:lll
func getUsersQuery() map[string]string {
	return map[string]string{
		"register": `{
		"query": "mutation { registerUser(input: { firstname: \"%s\", lastname:\"%s\", email: \"%s\", userLoginCredentials: { username:\"%s\", password: \"%s\" } }) { token, expires, threshold }}"
		}`,

		"login": `{
		"query": "mutation { loginUser(input: { username:\"%s\", password: \"%s\" }) { token, expires, threshold }}"
		}`,

		"refresh": `{
		"query": "mutation { refreshToken() { token expires threshold }}"
		}`,

		"delete": `{
	    "query": "mutation { deleteUser(input: { username: \"%s\" password: \"%s\" confirmation:\"I understand the consequences, delete my user account %s\" })}"
		}`,
	}
}

// getFiatQuery is a map of test Fiat mutations and queries.
//
//nolint:lll
func getFiatQuery() map[string]string {
	return map[string]string{
		"openFiat": `{
		"query": "mutation { openFiat(currency: \"%s\") { clientID, currency }}"
		}`,

		"depositFiat": `{
		"query": "mutation { depositFiat(input: { amount:%f, currency: \"%s\" }) { txId, clientId, txTimestamp, balance, lastTx, currency } }"
		}`,

		"exchangeOfferFiat": `{
		"query": "mutation { exchangeOfferFiat(input: { sourceCurrency:\"%s\" destinationCurrency: \"%s\" sourceAmount: %f }) { priceQuote{ clientID, sourceAcc, destinationAcc, rate, amount }, debitAmount, offerID, expires } }"
		}`,

		"exchangeTransferFiat": `{
		"query": "mutation { exchangeTransferFiat(offerID: \"%s\") { sourceReceipt { txId, clientId, txTimestamp, balance, lastTx, currency }, destinationReceipt { txId, clientId, txTimestamp, balance, lastTx, currency } } }"
		}`,

		"balanceFiat": `{
		"query": "query { balanceFiat(currencyCode: \"%s\") { currency, balance, lastTx, lastTxTs, createdAt, clientID } }"
		}`,

		"balanceAllFiat": `{
		"query": "query { balanceAllFiat( pageCursor: \"%s\", pageSize: %d ) { accountBalances { currency, balance, lastTx, lastTxTs, createdAt, clientID }, links { pageCursor } } }"
		}`,

		"balanceAllFiatNoParams": `{
		"query": "query { balanceAllFiat { accountBalances { currency, balance, lastTx, lastTxTs, createdAt, clientID }, links { pageCursor } } }"
		}`,

		"transactionDetailsFiat": `{
		"query": "query { transactionDetailsFiat( transactionID: \"%s\") }"
		}`,

		"transactionDetailsAllFiatInit": `{
		"query": "query { transactionDetailsAllFiat(input: { currency: \"%s\", pageSize:\"%d\", timezone:\"%s\", month: \"%d\", year:\"%d\" }) { transactions { currency, amount, transactedAt, clientID, txID }, links { pageCursor } } }"
		}`,

		"transactionDetailsAllFiatSubsequent": `{
		"query": "query { transactionDetailsAllFiat(input: { currency: \"%s\", pageSize:\"%d\", pageCursor:\"%s\" }) { transactions { currency, amount, transactedAt, clientID, txID }, links { pageCursor } } }"
		}`,
	}
}

// getCryptoQuery is a map of test Crypto mutations and queries.
//
//nolint:lll
func getCryptoQuery() map[string]string {
	return map[string]string{
		"openCrypto": `{
		"query": "mutation { openCrypto(ticker: \"%s\") { clientID, ticker } }"
		}`,

		"offerCrypto": `{
		"query": "mutation { offerCrypto(input: { sourceAmount: %f, sourceCurrency:\"%s\", destinationCurrency:\"%s\", isPurchase: %t, }) { priceQuote { clientID, sourceAcc, destinationAcc, rate, amount }, debitAmount, offerID, expires } }"
		}`,

		"exchangeCrypto": `{
		"query": "mutation { exchangeCrypto(offerID: \"%s\") { fiatTxReceipt{ currency, amount, transactedAt, clientID, txID, }, cryptoTxReceipt{ ticker, amount, transactedAt, clientID, txID, }, } }"
		}`,

		"balanceCrypto": `{
		"query": "query { balanceCrypto(ticker: \"%s\") { ticker, balance, lastTx, lastTxTs, createdAt, clientID } }"
		}`,

		"balanceAllCrypto": `{
		"query": "query { balanceAllCrypto( pageCursor: \"%s\", pageSize: %d ) { accountBalances { ticker, balance, lastTx, lastTxTs, createdAt, clientID }, links { pageCursor } } }"
		}`,

		"balanceAllCryptoNoParams": `{
		"query": "query { balanceAllCrypto { accountBalances { ticker, balance, lastTx, lastTxTs, createdAt, clientID }, links { pageCursor } } }"
		}`,

		"transactionDetailsCrypto": `{
		"query": "query { transactionDetailsCrypto(transactionID: \"%s\") }"
		}`,
	}
}
