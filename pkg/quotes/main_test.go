package quotes

import (
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
	"go.uber.org/zap"
)

// quotesConfigTestData is a map Quotes configuration test data.
var quotesConfigTestData = configTestData()

// testConfigs is the name of the Quotes configuration to use in the tests.
var testConfigs *config

// zapLogger is the Zap logger used strictly for the test suite in this package.
var zapLogger *logger.Logger

// quotes is used test wide to access third-party currency pricing services.
var quotes Quotes

func TestMain(m *testing.M) {
	// Parse commandline flags to check for short tests.
	flag.Parse()

	var err error
	// Configure logger.
	if zapLogger, err = logger.NewTestLogger(); err != nil {
		log.Printf("Test suite logger setup failed: %v\n", err)
		os.Exit(1)
	}

	// Setup test space.
	if err = setup(); err != nil {
		zapLogger.Error("Test suite setup failure", zap.Error(err))
		os.Exit(1)
	}

	// Configure Quotes.
	clientFiat, err := configFiatClient(testConfigs)
	if err != nil {
		zapLogger.Error("Failed to configure Fiat client", zap.Error(err))
		os.Exit(1)
	}

	quotes = &quotesImpl{
		clientFiat: clientFiat,
		conf:       testConfigs,
		logger:     zapLogger,
	}

	// Run test suite.
	exitCode := m.Run()

	// Cleanup test space.
	if err = tearDown(); err != nil {
		zapLogger.Error("Test suite teardown failure:", zap.Error(err))
		os.Exit(1)
	}

	os.Exit(exitCode)
}

// setup will configure the connection to the test database.
func setup() error {
	if testing.Short() {
		zapLogger.Warn("Short test: Skipping Quotes integration tests")

		return nil
	}

	var (
		rawConfigs string
		fs         afero.Fs
	)

	// If running on a GitHub Actions runner use the secret stored in the GitHub Actions Secrets.
	if _, ok := os.LookupEnv(constants.GetGithubCIKey()); ok {
		zapLogger.Info("Integration Test running on Github CI runner.")
		zapLogger.Warn("*** Please ensure that the Quotes configurations are upto date in GHA Secrets ***")

		if rawConfigs, ok = os.LookupEnv(constants.GetQuotesPrefix() + "_CI_CONFIGS"); !ok {
			msg := "failed to load Quotes configs from GitHub Actions Secrets"
			zapLogger.Error(msg)

			return fmt.Errorf(msg)
		}
	} else {
		fsDevLoader := afero.Afero{
			Fs: afero.NewOsFs(),
		}

		var bytes []byte

		bytes, err := fsDevLoader.ReadFile("../../configs/DevQuotesConfig.yaml")
		if err != nil || len(bytes) < 1 {
			zapLogger.Error("Please ensure there are API Keys and Endpoints configured in" +
				"\".configs/DevQuotesConfig.yaml\" for development.")

			return fmt.Errorf("failed to read dev confiq file, number of bytes: %d %w", len(bytes), err)
		}

		rawConfigs = string(bytes)
	}

	// Configure in memory file system to load configs.
	fs = afero.NewMemMapFs()

	if err := fs.MkdirAll(constants.GetEtcDir(), 0644); err != nil {
		return fmt.Errorf("failed to create in memory directory %w", err)
	}

	if err := afero.WriteFile(
		fs,
		constants.GetEtcDir()+constants.GetQuotesFileName(),
		[]byte(rawConfigs),
		0644); err != nil {
		return fmt.Errorf("failed to write in memory file %w", err)
	}

	// Load configs for tests.
	testConfigs = newConfig()
	if err := testConfigs.Load(fs); err != nil {
		return fmt.Errorf("failed to load test configs %w", err)
	}

	return nil
}

// tearDown will delete the test clusters keyspace.
func tearDown() error {
	return nil
}
