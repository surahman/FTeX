package postgres

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

// configTestData will return a map of test data containing valid and invalid logger configs.
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
func getTestUsers() map[string]createUserParams {
	users := make(map[string]createUserParams)
	username := "username%d"
	password := "user-password-%d"
	firstname := "firstname-%d"
	lastname := "lastname-%d"
	email := "user%d@email-address.com"

	for idx := 1; idx < 5; idx++ {
		uname := fmt.Sprintf(username, idx)
		users[uname] = createUserParams{
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
func getTestFiatAccounts(clientID1, clientID2 pgtype.UUID) map[string][]fiatCreateAccountParams {
	return map[string][]fiatCreateAccountParams{
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
func getTestJournalInternalFiatAccounts(clientID1, clientID2 pgtype.UUID) (
	map[string]fiatInternalTransferJournalEntryParams, error) {
	amount1Credit := pgtype.Numeric{}
	if err := amount1Credit.Scan("123.45"); err != nil {
		return nil, fmt.Errorf("failed to convert 123.45: %w", err)
	}

	amount1Debit := pgtype.Numeric{}
	if err := amount1Debit.Scan("-123.45"); err != nil {
		return nil, fmt.Errorf("failed to convert -123.45: %w", err)
	}

	amount2Credit := pgtype.Numeric{}
	if err := amount2Credit.Scan("4567.89"); err != nil {
		return nil, fmt.Errorf("failed to convert 4567.89 %w", err)
	}

	amount2Debit := pgtype.Numeric{}
	if err := amount2Debit.Scan("-4567.89"); err != nil {
		return nil, fmt.Errorf("failed to convert -4567.89 %w", err)
	}

	amount3Credit := pgtype.Numeric{}
	if err := amount3Credit.Scan("9192.24"); err != nil {
		return nil, fmt.Errorf("failed to convert 9192.24 %w", err)
	}

	amount3Debit := pgtype.Numeric{}
	if err := amount3Debit.Scan("-9192.24"); err != nil {
		return nil, fmt.Errorf("failed to convert -9192.24 %w", err)
	}

	return map[string]fiatInternalTransferJournalEntryParams{
			"CAD-AED": {
				SourceAccount:       clientID1,
				SourceCurrency:      CurrencyCAD,
				DestinationAccount:  clientID2,
				DestinationCurrency: CurrencyAED,
				CreditAmount:        amount1Credit,
				DebitAmount:         amount1Debit,
			},
			"CAD-USD": {
				SourceAccount:       clientID1,
				SourceCurrency:      CurrencyCAD,
				DestinationAccount:  clientID2,
				DestinationCurrency: CurrencyUSD,
				CreditAmount:        amount2Credit,
				DebitAmount:         amount2Debit,
			},
			"USD-AED": {
				SourceAccount:       clientID1,
				SourceCurrency:      CurrencyUSD,
				DestinationAccount:  clientID2,
				DestinationCurrency: CurrencyAED,
				CreditAmount:        amount3Credit,
				DebitAmount:         amount3Debit,
			},
		},
		nil
}

// getTestFiatJournal generates a number of test general ledger entry parameters.
func getTestFiatJournal(clientID1, clientID2 pgtype.UUID) (
	map[string]fiatExternalTransferJournalEntryParams, error) {
	// Create balance amounts.
	amount1 := pgtype.Numeric{}
	if err := amount1.Scan("1024.55"); err != nil {
		return nil, fmt.Errorf("failed to marshal 1024.55 to pgtype %w", err)
	}

	amount2 := pgtype.Numeric{}
	if err := amount2.Scan("4096.89"); err != nil {
		return nil, fmt.Errorf("failed to marshal 4096.89 to pgtype %w", err)
	}

	amount3 := pgtype.Numeric{}
	if err := amount3.Scan("256.44"); err != nil {
		return nil, fmt.Errorf("failed to marshal 256.44 to pgtype %w", err)
	}

	return map[string]fiatExternalTransferJournalEntryParams{
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
		},
		nil
}
