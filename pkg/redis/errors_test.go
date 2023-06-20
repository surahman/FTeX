package redis

import (
	"fmt"
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
			expectedCode: ErrorUnknown,
		}, {
			name:         "cache miss",
			err:          NewError("cache miss").errorCacheMiss(),
			expectedCode: ErrorCacheMiss,
		}, {
			name:         "cache set",
			err:          NewError("cache set").errorCacheSet(),
			expectedCode: ErrorCacheSet,
		}, {
			name:         "cache del",
			err:          NewError("cache del").errorCacheDel(),
			expectedCode: ErrorCacheDel,
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
			name:            "base vs cache miss",
			inputErr:        NewError("").errorCacheMiss(),
			baseErr:         baseError,
			boolExpectation: require.False,
		}, {
			name:            "base vs cache set",
			inputErr:        NewError("").errorCacheSet(),
			baseErr:         baseError,
			boolExpectation: require.False,
		}, {
			name:            "base vs cache del",
			inputErr:        NewError("").errorCacheDel(),
			baseErr:         baseError,
			boolExpectation: require.False,
		}, {
			name:            "cache miss",
			inputErr:        NewError("").errorCacheMiss(),
			baseErr:         NewError("").errorCacheMiss(),
			boolExpectation: require.True,
		}, {
			name:            "cache set",
			inputErr:        NewError("").errorCacheSet(),
			baseErr:         NewError("").errorCacheSet(),
			boolExpectation: require.True,
		}, {
			name:            "cache del",
			inputErr:        NewError("").errorCacheDel(),
			baseErr:         NewError("").errorCacheDel(),
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
