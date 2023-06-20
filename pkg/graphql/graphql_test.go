package graphql

import (
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/quotes"
)

func TestNewGraphQLServer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockAuth := mocks.NewMockAuth(mockCtrl)
	mockPostgres := mocks.NewMockPostgres(mockCtrl)
	mockRedis := mocks.NewMockRedis(mockCtrl)
	mockQuotes := quotes.NewMockQuotes(mockCtrl)

	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(constants.EtcDir(), 0644), "Failed to create in memory directory")
	require.NoError(t, afero.WriteFile(fs, constants.EtcDir()+constants.HTTPGraphQLFileName(),
		[]byte(graphQLConfigTestData["valid"]), 0644), "Failed to write in memory file")

	server, err := NewServer(&fs, mockAuth, mockPostgres, mockRedis, mockQuotes, zapLogger, &sync.WaitGroup{})
	require.NoError(t, err, "error whilst creating mock server")
	require.NotNil(t, server, "failed to create mock server")
}
