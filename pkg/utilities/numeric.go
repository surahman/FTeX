package utilities

import (
	"fmt"
)

// Float64TwoDecimalPlaces will round down to two decimal places and return the float64 representation.
func Float64TwoDecimalPlaces(val float64) float64 {
	return float64(int(val*100)) / 100 //nolint:gomnd
}

// Float64TwoDecimalPlacesString will round down to two decimal places and return the string representation.
func Float64TwoDecimalPlacesString(val float64) string {
	return fmt.Sprintf("%.2f", Float64TwoDecimalPlaces(val))
}
