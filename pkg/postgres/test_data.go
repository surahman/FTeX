package postgres

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

// configTestData will return a map of test data containing valid and invalid Postgres configs.
func configTestData() map[string]string {
	return map[string]string{
		"empty": ``,

		"test_suite": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db_test
  host: 127.0.0.1
  maxConnectionAttempts: 1
  port: 6432
  timeout: 5
pool:
  healthCheckPeriod: 30s
  maxConns: 4
  minConns: 4`,

		"github-ci-runner": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db_test
  maxConnectionAttempts: 3
  host: 127.0.0.1
  port: 5432
  timeout: 5
pool:
  healthCheckPeriod: 30s
  maxConns: 8
  minConns: 4`,

		"valid": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db
  host: 127.0.0.1
  maxConnectionAttempts: 5
  port: 6432
  timeout: 5
pool:
  healthCheckPeriod: 30s
  maxConns: 8
  minConns: 4`,

		"bad_health_check": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db
  host: 127.0.0.1
  maxConnectionAttempts: 5
  port: 6432
  timeout: 5
pool:
  healthCheckPeriod: 3s
  maxConns: 8
  minConns: 4`,

		"invalid_conns": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db
  host: 127.0.0.1
  maxConnectionAttempts: 5
  port: 6432
  timeout: 5
pool:
  healthCheckPeriod: 30s
  maxConns: 2
  minConns: 2`,

		"invalid_max_conn_attempts": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db
  host: 127.0.0.1
  maxConnectionAttempts: 0
  port: 6432
  timeout: 5
pool:
  healthCheckPeriod: 30s
  maxConns: 8
  minConns: 4`,

		"invalid_timeout": `
authentication:
  username: postgres
  password: postgres
connection:
  database: ftex_db
  host: 127.0.0.1
  maxConnectionAttempts: 5
  port: 6432
  timeout: 2
pool:
  healthCheckPeriod: 30s
  maxConns: 8
  minConns: 4`,
	}
}

// getTestUsers will generate a number of test users for testing.
func getTestUsers() map[string]userCreateParams {
	users := make(map[string]userCreateParams)
	username := "username%d"
	password := "user-password-%d"
	firstname := "firstname-%d"
	lastname := "lastname-%d"
	email := "user%d@email-address.com"

	for idx := 1; idx < 5; idx++ {
		uname := fmt.Sprintf(username, idx)
		users[uname] = userCreateParams{
			Username:  fmt.Sprintf(username, idx),
			Password:  fmt.Sprintf(password, idx),
			FirstName: fmt.Sprintf(firstname, idx),
			LastName:  fmt.Sprintf(lastname, idx),
			Email:     fmt.Sprintf(email, idx),
		}
	}

	return users
}

// getTestFiatAccounts generates a number of test fiat accounts.
func getTestFiatAccounts(clientID1, clientID2 uuid.UUID) map[string][]FiatCreateAccountParams {
	return map[string][]FiatCreateAccountParams{
		"clientID1": {
			{
				ClientID: clientID1,
				Currency: CurrencyAED,
			}, {
				ClientID: clientID1,
				Currency: CurrencyUSD,
			}, {
				ClientID: clientID1,
				Currency: CurrencyCAD,
			},
		},
		"clientID2": {
			{
				ClientID: clientID2,
				Currency: CurrencyAED,
			}, {
				ClientID: clientID2,
				Currency: CurrencyUSD,
			}, {
				ClientID: clientID2,
				Currency: CurrencyCAD,
			},
		},
	}
}

// getTestJournalInternalFiatAccounts generates a number of test fiat internal transfer journal entries.
func getTestJournalInternalFiatAccounts(
	clientID1,
	clientID2 uuid.UUID) map[string]FiatInternalTransferJournalEntryParams {
	return map[string]FiatInternalTransferJournalEntryParams{
		"CAD-AED": {
			SourceAccount:       clientID1,
			SourceCurrency:      CurrencyCAD,
			DestinationAccount:  clientID2,
			DestinationCurrency: CurrencyAED,
			CreditAmount:        decimal.NewFromFloat(123.45),
			DebitAmount:         decimal.NewFromFloat(-123.45),
		},
		"CAD-USD": {
			SourceAccount:       clientID1,
			SourceCurrency:      CurrencyCAD,
			DestinationAccount:  clientID2,
			DestinationCurrency: CurrencyUSD,
			CreditAmount:        decimal.NewFromFloat(4567.89),
			DebitAmount:         decimal.NewFromFloat(-4567.89),
		},
		"USD-AED": {
			SourceAccount:       clientID1,
			SourceCurrency:      CurrencyUSD,
			DestinationAccount:  clientID2,
			DestinationCurrency: CurrencyAED,
			CreditAmount:        decimal.NewFromFloat(9192.24),
			DebitAmount:         decimal.NewFromFloat(-9192.24),
		},
	}
}

// getTestFiatJournal generates a number of test general ledger entry parameters.
func getTestFiatJournal(clientID1, clientID2 uuid.UUID) map[string]FiatExternalTransferJournalEntryParams {
	// Create balance amounts.
	var (
		amount1 = decimal.NewFromFloat(1024.55)
		amount2 = decimal.NewFromFloat(4096.89)
		amount3 = decimal.NewFromFloat(256.44)
	)

	return map[string]FiatExternalTransferJournalEntryParams{
		"Client ID 1 - USD": {
			ClientID: clientID1,
			Currency: CurrencyUSD,
			Amount:   amount1,
		},
		"Client ID 1 - AED": {
			ClientID: clientID1,
			Currency: CurrencyAED,
			Amount:   amount2,
		},
		"Client ID 1 - CAD": {
			ClientID: clientID1,
			Currency: CurrencyCAD,
			Amount:   amount3,
		},
		"Client ID 2 - USD": {
			ClientID: clientID2,
			Currency: CurrencyUSD,
			Amount:   amount2,
		},
		"Client ID 2 - AED": {
			ClientID: clientID2,
			Currency: CurrencyAED,
			Amount:   amount3,
		},
		"Client ID 2 - CAD": {
			ClientID: clientID2,
			Currency: CurrencyCAD,
			Amount:   amount1,
		},
	}
}
