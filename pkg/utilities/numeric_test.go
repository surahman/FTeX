package utilities

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUtilities_Float64TwoDecimalPlaces(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		value    float64
		expected float64
	}{
		{
			name:     "1999.999999",
			value:    1999.999999,
			expected: 1999.99,
		}, {
			name:     "99999999.999999",
			value:    99999999.999999,
			expected: 99999999.99,
		}, {
			name:     "1.00001",
			value:    1.00001,
			expected: 1.0,
		}, {
			name:     "1.1111111",
			value:    1.1111111,
			expected: 1.11,
		}, {
			name:     "pi",
			value:    3.14159265359,
			expected: 3.14,
		}, {
			name:     "-1999.999999",
			value:    -1999.999999,
			expected: -1999.99,
		}, {
			name:     "-99999999.999999",
			value:    -99999999.999999,
			expected: -99999999.99,
		}, {
			name:     "-1.00001",
			value:    -1.00001,
			expected: -1.0,
		}, {
			name:     "1.1111111",
			value:    -1.1111111,
			expected: -1.11,
		}, {
			name:     "pi",
			value:    -3.14159265359,
			expected: -3.14,
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(fmt.Sprintf("Truncating %s", test.name), func(t *testing.T) {
			t.Parallel()

			actual := Float64TwoDecimalPlaces(test.value)
			require.InDeltaf(t, test.expected, actual, 0.001, "decimal truncation not within delta %f.", actual)
		})
	}
}

func TestUtilities_Float64TwoDecimalPlacesString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		expected string
		value    float64
	}{
		{
			name:     "1999.999999",
			value:    1999.999999,
			expected: "1999.99",
		}, {
			name:     "99999999.999999",
			value:    99999999.999999,
			expected: "99999999.99",
		}, {
			name:     "1.00001",
			value:    1.00001,
			expected: "1.00",
		}, {
			name:     "1.1111111",
			value:    1.1111111,
			expected: "1.11",
		}, {
			name:     "pi",
			value:    3.14159265359,
			expected: "3.14",
		}, {
			name:     "-1999.999999",
			value:    -1999.999999,
			expected: "-1999.99",
		}, {
			name:     "-99999999.999999",
			value:    -99999999.999999,
			expected: "-99999999.99",
		}, {
			name:     "-1.00001",
			value:    -1.00001,
			expected: "-1.00",
		}, {
			name:     "1.1111111",
			value:    -1.1111111,
			expected: "-1.11",
		}, {
			name:     "pi",
			value:    -3.14159265359,
			expected: "-3.14",
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(fmt.Sprintf("Truncating %s", test.name), func(t *testing.T) {
			t.Parallel()

			actual := Float64TwoDecimalPlacesString(test.value)
			require.Equal(t, test.expected, actual)
		})
	}
}
