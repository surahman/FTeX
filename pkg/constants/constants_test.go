package constants

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetGithubCIKey(t *testing.T) {
	require.Equal(t, githubCIKey, GetGithubCIKey(), "Incorrect Github CI environment key")
}

func TestGetEtcDir(t *testing.T) {
	require.Equal(t, configEtcDir, GetEtcDir(), "Incorrect etc directory")
}

func TestGetHomeDir(t *testing.T) {
	require.Equal(t, configHomeDir, GetHomeDir(), "Incorrect home directory")
}

func TestGetBaseDir(t *testing.T) {
	require.Equal(t, configBaseDir, GetBaseDir(), "Incorrect base directory")
}
