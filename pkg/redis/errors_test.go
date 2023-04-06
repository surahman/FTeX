package redis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		err          *Error
		expectedCode int
		expectedType any
	}{
		// ----- test cases start ----- //
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
		// ----- test cases end ----- //
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
