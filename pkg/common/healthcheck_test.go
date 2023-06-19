package common

import (
	"errors"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/mocks"
)

func TestCommon_HTTPHealthcheck(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		expectMsg     string
		expectCode    int
		databaseErr   error
		databaseTimes int
		cacheErr      error
		cacheTimes    int
		expectErr     require.ErrorAssertionFunc
	}{
		{
			name:          "postgres failure",
			databaseErr:   errors.New("db failure"),
			databaseTimes: 1,
			cacheErr:      nil,
			cacheTimes:    0,
			expectCode:    http.StatusServiceUnavailable,
			expectMsg:     "Postgres",
			expectErr:     require.Error,
		}, {
			name:          "redis failure",
			databaseErr:   nil,
			databaseTimes: 1,
			cacheErr:      errors.New("cache failure"),
			cacheTimes:    1,
			expectCode:    http.StatusServiceUnavailable,
			expectMsg:     "Redis",
			expectErr:     require.Error,
		}, {
			name:          "healthy",
			databaseErr:   nil,
			databaseTimes: 1,
			cacheErr:      nil,
			cacheTimes:    1,
			expectCode:    http.StatusOK,
			expectMsg:     "healthy",
			expectErr:     require.NoError,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockDB := mocks.NewMockPostgres(mockCtrl)
			mockCache := mocks.NewMockRedis(mockCtrl)

			gomock.InOrder(
				mockDB.EXPECT().
					Healthcheck().
					Return(test.databaseErr).
					Times(test.databaseTimes),

				mockCache.EXPECT().
					Healthcheck().
					Return(test.cacheErr).
					Times(test.cacheTimes),
			)

			actualCode, actualMsg, err := HTTPHealthcheck(mockDB, mockCache, zapLogger)
			test.expectErr(t, err, "error expectation failed.")
			require.Equal(t, test.expectCode, actualCode, "http codes mismatched.")
			require.Contains(t, actualMsg, test.expectMsg, "http message mismatched.")
		})
	}
}
