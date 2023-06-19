package graphql

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/mocks"
)

func TestGinContextFromContext(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		expectedMsg string
		expectErr   require.ErrorAssertionFunc
		ctx         context.Context //nolint:containedctx
	}{
		{
			name:        "no context",
			expectedMsg: "information not found",
			expectErr:   require.Error,
			ctx:         context.TODO(),
		}, {
			name:        "incorrect context",
			expectedMsg: "information malformed",
			expectErr:   require.Error,
			ctx:         context.WithValue(context.TODO(), GinContextKey{}, context.TODO()),
		}, {
			name:      "success",
			expectErr: require.NoError,
			ctx:       context.WithValue(context.TODO(), GinContextKey{}, &gin.Context{}),
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := GinContextFromContext(test.ctx, zapLogger)

			test.expectErr(t, err, "error expectation failed")
			if err != nil {
				require.Contains(t, err.Error(), test.expectedMsg, "incorrect error message returned")
			}
		})
	}
}

func TestAuthorizationCheck(t *testing.T) {
	t.Parallel()

	ginCtxNoAuth := &gin.Context{Request: &http.Request{Header: http.Header{}}}
	ginCtxNoAuth.Request.Header.Add(testAuthHeaderKey, "")

	ginCtxAuth := &gin.Context{Request: &http.Request{Header: http.Header{}}}
	ginCtxAuth.Request.Header.Add(testAuthHeaderKey, "test-token")

	testCases := []struct {
		name                 string
		expectedMsg          string
		expectErr            require.ErrorAssertionFunc
		ctx                  context.Context //nolint:containedctx
		authValidateJWTErr   error
		authValidateJWTTimes int
	}{
		{
			name:                 "no context",
			expectedMsg:          "information not found",
			expectErr:            require.Error,
			ctx:                  context.TODO(),
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 0,
		}, {
			name:                 "incorrect context",
			expectedMsg:          "information malformed",
			expectErr:            require.Error,
			ctx:                  context.WithValue(context.TODO(), GinContextKey{}, context.TODO()),
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 0,
		}, {
			name:                 "no token",
			expectedMsg:          "does not contain",
			expectErr:            require.Error,
			ctx:                  context.WithValue(context.TODO(), GinContextKey{}, ginCtxNoAuth),
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 0,
		}, {
			name:                 "no token",
			expectedMsg:          "failed to authenticate token",
			expectErr:            require.Error,
			ctx:                  context.WithValue(context.TODO(), GinContextKey{}, ginCtxAuth),
			authValidateJWTErr:   errors.New("failed to authenticate token"),
			authValidateJWTTimes: 1,
		}, {
			name:                 "success",
			expectErr:            require.NoError,
			ctx:                  context.WithValue(context.TODO(), GinContextKey{}, ginCtxAuth),
			authValidateJWTErr:   nil,
			authValidateJWTTimes: 1,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockAuth := mocks.NewMockAuth(mockCtrl)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).
				Return(uuid.UUID{}, int64(-1), test.authValidateJWTErr).
				Times(test.authValidateJWTTimes)

			_, _, err := AuthorizationCheck(test.ctx, mockAuth, zapLogger, testAuthHeaderKey)

			test.expectErr(t, err, "error expectation failed")
			if err != nil {
				require.Contains(t, err.Error(), test.expectedMsg, "incorrect error message returned")
			}
		})
	}
}
