package auth

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError_New(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		err          *Error
		expectedCode int
	}{
		{
			name:         "base error",
			err:          NewError("base error"),
			expectedCode: 0,
		}, {
			name:         "bad request",
			err:          NewError("bad request").SetStatus(http.StatusBadRequest),
			expectedCode: http.StatusBadRequest,
		}, {
			name:         "too many requests",
			err:          NewError("too many requests").SetStatus(http.StatusTooManyRequests),
			expectedCode: http.StatusTooManyRequests,
		}, {
			name:         "bad gateway",
			err:          NewError("bad gateway").SetStatus(http.StatusBadGateway),
			expectedCode: http.StatusBadGateway,
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.NotNil(t, test.err, "error should not be nil")
			require.Equal(t, test.expectedCode, test.err.Code, "expected error code did not match")
			require.Equal(t, test.name, test.err.Message, "error messages did not match")
		})
	}
}

func TestError_Is(t *testing.T) {
	t.Parallel()

	baseError := NewError("base error")

	testCases := []struct {
		name            string
		inputErr        error
		baseErr         *Error
		boolExpectation require.BoolAssertionFunc
	}{
		{
			name:            "input nil",
			inputErr:        nil,
			baseErr:         baseError,
			boolExpectation: require.False,
		}, {
			name:            "base nil",
			inputErr:        nil,
			baseErr:         nil,
			boolExpectation: require.False,
		}, {
			name:            "base different",
			inputErr:        fmt.Errorf("different error"),
			baseErr:         nil,
			boolExpectation: require.False,
		}, {
			name:            "base vs too many requests",
			inputErr:        &Error{Message: "input", Code: http.StatusTooManyRequests},
			baseErr:         baseError,
			boolExpectation: require.False,
		}, {
			name:            "too many requests",
			inputErr:        &Error{Message: "input", Code: http.StatusTooManyRequests},
			baseErr:         &Error{Message: "base", Code: http.StatusTooManyRequests},
			boolExpectation: require.True,
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := test.baseErr.Is(test.inputErr)
			test.boolExpectation(t, result, "error is value expectation failed.")
		})
	}
}
