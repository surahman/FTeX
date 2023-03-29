package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModels_CurrencyScan(t *testing.T) {
	t.Run("Byte Array", func(t *testing.T) {
		var curr Currency
		err := curr.Scan([]byte(CurrencyUSD))
		require.NoError(t, err, "valid byte array")
	})

	t.Run("String", func(t *testing.T) {
		var curr Currency
		err := curr.Scan("USD")
		require.NoError(t, err, "valid string")
	})

	t.Run("Invalid Type", func(t *testing.T) {
		var curr Currency
		err := curr.Scan(123)
		require.Error(t, err, "invalid type")
	})
}

func TestModels_CurrencyValid(t *testing.T) {
	testCases := []struct {
		name        string
		currency    Currency
		errExpected require.BoolAssertionFunc
	}{
		{
			name:        "Valid - USD",
			currency:    CurrencyUSD,
			errExpected: require.True,
		},
		{
			name:        "Valid - DEPOSIT",
			currency:    CurrencyDEPOSIT,
			errExpected: require.True,
		},
		{
			name:        "Valid - CRYPTO",
			currency:    CurrencyCRYPTO,
			errExpected: require.True,
		},
		{
			name:        "Invalid",
			currency:    "XYZ",
			errExpected: require.False,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.errExpected(t, testCase.currency.Valid(), "expected error condition failed.")
		})
	}
}

func TestModels_NullCurrencyScan(t *testing.T) {
	testCases := []struct {
		name         string
		nullCurr     any
		errExpected  require.ErrorAssertionFunc
		boolExpected require.BoolAssertionFunc
	}{
		{
			name:         "nil",
			nullCurr:     nil,
			errExpected:  require.NoError,
			boolExpected: require.False,
		},
		{
			name:         "Invalid",
			nullCurr:     123,
			errExpected:  require.Error,
			boolExpected: require.True,
		},
		{
			name:         "Valid - Byte Array",
			nullCurr:     []byte("USD"),
			errExpected:  require.NoError,
			boolExpected: require.True,
		},
		{
			name:         "Valid - String",
			nullCurr:     "USD",
			errExpected:  require.NoError,
			boolExpected: require.True,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var ns NullCurrency
			testCase.errExpected(t, ns.Scan(testCase.nullCurr), "expected error condition failed.")
			testCase.boolExpected(t, ns.Valid, "expected validity condition failed.")
		})
	}
}

func TestModels_NullCurrencyValue(t *testing.T) {
	testCases := []struct {
		name        string
		driverValue string
		nullCurr    NullCurrency
		errExpected require.ErrorAssertionFunc
		nilExpected require.ValueAssertionFunc
	}{
		{
			name:        "Invalid",
			driverValue: "",
			nullCurr: NullCurrency{
				Currency: "invalid",
				Valid:    false,
			},
			errExpected: require.NoError,
			nilExpected: require.Nil,
		},
		{
			name:        "Valid",
			driverValue: "DEPOSIT",
			nullCurr: NullCurrency{
				Currency: CurrencyDEPOSIT,
				Valid:    true,
			},
			errExpected: require.NoError,
			nilExpected: require.NotNil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			driver, err := testCase.nullCurr.Value()
			testCase.errExpected(t, err, "expected error condition failed.")
			testCase.nilExpected(t, driver, "nil driver value expectation failed.")
			if driver == nil {
				return
			}
			curr, ok := driver.(string)
			require.True(t, ok, "driver cast to string failed.")
			require.Equal(t, testCase.driverValue, curr, "incorrect driver value.")
		})
	}
}
