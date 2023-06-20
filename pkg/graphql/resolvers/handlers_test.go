package graphql

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/quotes"
)

func TestQueryHandler(t *testing.T) {
	// Mock configurations.
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockAuth := mocks.NewMockAuth(mockCtrl)
	mockPostgres := mocks.NewMockPostgres(mockCtrl)
	mockRedis := mocks.NewMockRedis(mockCtrl)
	mockQuotes := quotes.NewMockQuotes(mockCtrl)

	handler := QueryHandler("Authorization", mockAuth, mockRedis, mockPostgres, mockQuotes, zapLogger)

	require.NotNil(t, handler, "failed to create graphql endpoint handler")
}

func TestPlaygroundHandler(t *testing.T) {
	handler := PlaygroundHandler("/base-url", "/query-endpoint-url")
	require.NotNil(t, handler, "failed to create playground endpoint handler")
}

func TestGinContextToContextMiddleware(t *testing.T) {
	router := gin.Default()
	router.POST("/middleware-test", GinContextToContextMiddleware())

	req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, "/middleware-test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify responses
	require.Equal(t, http.StatusOK, w.Code, "expected status codes do not match")
}
