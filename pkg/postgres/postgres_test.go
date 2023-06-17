package postgres

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
)

func TestNewPostgres(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(constants.EtcDir(), 0644), "Failed to create in memory directory")
	require.NoError(t, afero.WriteFile(fs, constants.EtcDir()+constants.PostgresFileName(),
		[]byte(postgresConfigTestData["test_suite"]), 0644), "Failed to write in memory file")

	testCases := []struct {
		name      string
		fs        *afero.Fs
		log       *logger.Logger
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
	}{
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
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			postgres, err := NewPostgres(test.fs, test.log)
			test.expectErr(t, err, "error expectation failed.")
			test.expectNil(t, postgres, "nil expectation for returned interface failed.")
		})
	}
}

func TestNewPostgresImpl(t *testing.T) {
	testCases := []struct {
		name      string
		fileName  string
		input     string
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
	}{
		{
			name:      "File found",
			fileName:  constants.PostgresFileName(),
			input:     postgresConfigTestData["test_suite"],
			expectErr: require.NoError,
			expectNil: require.NotNil,
		}, {
			name:      "File not found",
			fileName:  "wrong_file_name.yaml",
			input:     postgresConfigTestData["test_suite"],
			expectErr: require.Error,
			expectNil: require.Nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Configure mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(constants.EtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, constants.EtcDir()+testCase.fileName,
				[]byte(testCase.input), 0644), "Failed to write in memory file")

			c, err := newPostgresImpl(&fs, zapLogger)
			testCase.expectErr(t, err)
			testCase.expectNil(t, c)
		})
	}
}

func TestPostgres_verifySession(t *testing.T) {
	// Integration test check.
	if testing.Short() {
		t.Skip()
	}

	// Nil Session.
	postgres := &postgresImpl{}
	require.Error(t, postgres.verifySession(), "nil session should return error")

	// Setup mock filesystem for configs.
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(constants.EtcDir(), 0644), "Failed to create in memory directory")
	require.NoError(t, afero.WriteFile(fs, constants.EtcDir()+constants.PostgresFileName(),
		[]byte(postgresConfigTestData[configFileKey]), 0644), "Failed to write in memory file")

	// Not open session.
	postgres, err := newPostgresImpl(&fs, zapLogger)
	require.NoError(t, err, "failed to load configuration")
	require.Error(t, postgres.verifySession(), "failed to return error on closed connection")

	// *** Skip tests below if not running integration tests ***
	if testing.Short() {
		t.Skip()
	}

	// Open connection, verify, close, and verify.
	require.NoError(t, postgres.Open(), "failed to open connection to postgres")
	require.NoError(t, postgres.verifySession(), "failed to verify Postgres connection is open")
	require.NoError(t, postgres.Close(), "failed to close established Postgres connection")
	require.Error(t, postgres.verifySession(), "failed to return error on closed Postgres connection")
}

func TestPostgres_Open(t *testing.T) {
	// Integration test check.
	if testing.Short() {
		t.Skip()
	}

	// Setup mock filesystem for configs.
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(constants.EtcDir(), 0644), "Failed to create in memory directory")
	require.NoError(t, afero.WriteFile(fs, constants.EtcDir()+constants.PostgresFileName(),
		[]byte(postgresConfigTestData[configFileKey]), 0644), "Failed to write in memory file")

	// Open and close session.
	postgres, err := newPostgresImpl(&fs, zapLogger)
	require.NoError(t, err, "failed to load configuration")
	require.NoError(t, postgres.Open(), "failed to open connection")
	require.NoError(t, postgres.Close(), "failed to close connection")
	require.Error(t, postgres.Close(), "failed to return error whilst closing a closed connection")

	// Ping failure.
	postgres.conf.Connection.Host = "bad-host-name"
	require.Error(t, postgres.Open(), "failed to report ping failure on bad connection")
}

func TestPostgres_Close(t *testing.T) {
	// Integration test check.
	if testing.Short() {
		t.Skip()
	}

	// Setup mock filesystem for configs.
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(constants.EtcDir(), 0644), "Failed to create in memory directory")
	require.NoError(t, afero.WriteFile(fs, constants.EtcDir()+constants.PostgresFileName(),
		[]byte(postgresConfigTestData[configFileKey]), 0644), "Failed to write in memory file")

	// Open and close session.
	postgres, err := newPostgresImpl(&fs, zapLogger)
	require.NoError(t, err, "failed to load configuration")
	require.Error(t, postgres.Close(), "failed to return error on closing a connection not opened")
	require.NoError(t, postgres.Open(), "failed to open connection.")
	require.NoError(t, postgres.Close(), "failed to close connection.")
}

func TestPostgres_Healthcheck(t *testing.T) {
	// Integration test check.
	if testing.Short() {
		t.Skip()
	}

	// Setup mock filesystem for configs.
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(constants.EtcDir(), 0644), "Failed to create in memory directory")
	require.NoError(t, afero.WriteFile(fs, constants.EtcDir()+constants.PostgresFileName(),
		[]byte(postgresConfigTestData[configFileKey]), 0644), "Failed to write in memory file")

	// Open and close session.
	postgres, err := newPostgresImpl(&fs, zapLogger)
	require.NoError(t, err, "failed to load configuration")
	require.Error(t, postgres.Healthcheck(), "failed to return bad health on uninitialized connection")

	require.NoError(t, postgres.Open(), "failed to open connection.")
	require.NoError(t, postgres.Healthcheck(), "failed to return good health on open connection")

	require.NoError(t, postgres.Close(), "failed to close connection.")
	require.Error(t, postgres.Healthcheck(), "failed to return bad health on closed connection")
}
