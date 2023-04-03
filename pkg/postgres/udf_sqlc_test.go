package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestUDF_RoundHalfEven(t *testing.T) {
	t.Parallel()

	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	testCases := []struct {
		name       string
		parameters testRoundHalfEvenParams
		expected   decimal.Decimal
	}{
		{
			name: "Num: 101.313723140517182, Scale: 4",
			parameters: testRoundHalfEvenParams{
				Num:   decimal.NewFromFloat(101.313723140517182),
				Scale: 4,
			},
			expected: decimal.NewFromFloat(101.3137),
		}, {
			name: "Num: 101.313723140517182, Scale: 3",
			parameters: testRoundHalfEvenParams{
				Num:   decimal.NewFromFloat(101.313723140517182),
				Scale: 3,
			},
			expected: decimal.NewFromFloat(101.314),
		}, {
			name: "Num: 89.774758184848455, Scale: 4",
			parameters: testRoundHalfEvenParams{
				Num:   decimal.NewFromFloat(89.774758184848455),
				Scale: 4,
			},
			expected: decimal.NewFromFloat(89.7748),
		}, {
			name: "Num: 89.7745, Scale: 3",
			parameters: testRoundHalfEvenParams{
				Num:   decimal.NewFromFloat(89.7745),
				Scale: 3,
			},
			expected: decimal.NewFromFloat(89.774),
		}, {
			name: "Num: 7.396500000, Scale: 3",
			parameters: testRoundHalfEvenParams{
				Num:   decimal.NewFromFloat(7.396500000),
				Scale: 3,
			},
			expected: decimal.NewFromFloat(7.396),
		}, {
			name: "Num: 7.39649999, Scale: 3",
			parameters: testRoundHalfEvenParams{
				Num:   decimal.NewFromFloat(7.39649999),
				Scale: 3,
			},
			expected: decimal.NewFromFloat(7.396),
		}, {
			name: "Num: 7.39649999, Scale: 4",
			parameters: testRoundHalfEvenParams{
				Num:   decimal.NewFromFloat(7.39649999),
				Scale: 4,
			},
			expected: decimal.NewFromFloat(7.3965),
		}, {
			name: "Num: 7.89649999, Scale: 1",
			parameters: testRoundHalfEvenParams{
				Num:   decimal.NewFromFloat(7.89649999),
				Scale: 1,
			},
			expected: decimal.NewFromFloat(7.9),
		}, {
			name: "Num: 7.89649999, Scale: 0",
			parameters: testRoundHalfEvenParams{
				Num:   decimal.NewFromFloat(7.89649999),
				Scale: 0,
			},
			expected: decimal.NewFromFloat(8.0),
		}, {
			name: "Num: 7.5, Scale: 0",
			parameters: testRoundHalfEvenParams{
				Num:   decimal.NewFromFloat(7.5),
				Scale: 0,
			},
			expected: decimal.NewFromFloat(8.0),
		}, {
			name: "Num: 8.5, Scale: 0",
			parameters: testRoundHalfEvenParams{
				Num:   decimal.NewFromFloat(8.5),
				Scale: 0,
			},
			expected: decimal.NewFromFloat(8.0),
		},
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)

	t.Cleanup(func() {
		cancel()
	})

	for _, testCase := range testCases {
		test := testCase

		t.Run(fmt.Sprintf("Round Half-to-Even: %s", test.name), func(t *testing.T) {
			t.Parallel()

			result, err := connection.Query.testRoundHalfEven(ctx, &test.parameters)
			require.NoError(t, err, "failed to get Half-to-Even rounded value.")
			require.Equal(t, test.expected, result, "expected result mismatched.")
		})
	}
}
