package postgres

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
)

func TestNewPostgres(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
	require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetPostgresFileName(),
		[]byte(postgresConfigTestData["test_suite"]), 0644), "Failed to write in memory file")

	testCases := []struct {
		name      string
		fs        *afero.Fs
		log       *logger.Logger
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			name:      "Invalid file system and logger",
			fs:        nil,
			log:       nil,
			expectErr: require.Error,
			expectNil: require.Nil,
		}, {
			name:      "Invalid file system",
			fs:        nil,
			log:       zapLogger,
			expectErr: require.Error,
			expectNil: require.Nil,
		}, {
			name:      "Invalid logger",
			fs:        &fs,
			log:       nil,
			expectErr: require.Error,
			expectNil: require.Nil,
		}, {
			name:      "Valid",
			fs:        &fs,
			log:       zapLogger,
			expectErr: require.NoError,
			expectNil: require.NotNil,
		},
		// ----- test cases end ----- //
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			postgres, err := NewPostgres(testCase.fs, testCase.log)
			testCase.expectErr(t, err)
			testCase.expectNil(t, postgres)
		})
	}
}
