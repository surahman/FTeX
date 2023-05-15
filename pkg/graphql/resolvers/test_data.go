package graphql

// getHealthcheckQuery is the health check query.
func getHealthcheckQuery() string {
	return `{
		"query": "query { healthcheck() }"
	}`
}

// getUsersQuery is a map of test user queries.
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

// getFiatQuery is a map of test Fiat queries.
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
		"query": "mutation { exchangeOfferFiat(input: { sourceCurrency:\"%s\" destinationCurrency: \"%s\" sourceAmount: %f }) { priceQuote{ ClientID, SourceAcc, DestinationAcc, Rate, Amount }, debitAmount, offerID, expires } }"
		}`,

		"exchangeTransferFiat": `{
		"query": "mutation { exchangeTransferFiat(offerID: \"%s\") { sourceReceipt { txId, clientId, txTimestamp, balance, lastTx, currency }, destinationReceipt { txId, clientId, txTimestamp, balance, lastTx, currency } } }"
		}`,
	}
}
