package configloader

import (
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/rs/xid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/validator"
	"gopkg.in/yaml.v3"
)

func TestConfigLoader(t *testing.T) {
	keyspaceLevel1 := testPrefix + "_PARENTLEVEL1."
	keyspaceLevel2 := testPrefix + "_PARENTLEVEL2."

	testCases := []struct {
		name      string
		input     string
		expectErr require.ErrorAssertionFunc
		expectLen int
	}{
		{
			name:      "empty - etc dir",
			input:     testConfigTestData["empty"],
			expectErr: require.Error,
			expectLen: 4,
		}, {
			name:      "valid - etc dir",
			input:     testConfigTestData["valid"],
			expectErr: require.NoError,
			expectLen: 0,
		}, {
			name:      "valid - required",
			input:     testConfigTestData["valid - required"],
			expectErr: require.NoError,
			expectLen: 0,
		}, {
			name:      "invalid - no lvl 1",
			input:     testConfigTestData["invalid - no lvl 1"],
			expectErr: require.Error,
			expectLen: 2,
		}, {
			name:      "invalid - no lvl1 child1",
			input:     testConfigTestData["invalid - no lvl1 child1"],
			expectErr: require.Error,
			expectLen: 1,
		}, {
			name:      "invalid - lvl1 child1 below threshold",
			input:     testConfigTestData["invalid - lvl1 child1 below threshold"],
			expectErr: require.Error,
			expectLen: 1,
		}, {
			name:      "invalid - lvl2 no child1",
			input:     testConfigTestData["invalid - lvl2 no child1"],
			expectErr: require.Error,
			expectLen: 1,
		}, {
			name:      "invalid - lvl2 child1 below threshold",
			input:     testConfigTestData["invalid - lvl2 child1 below threshold"],
			expectErr: require.Error,
			expectLen: 1,
		}, {
			name:      "invalid - lvl2 no child3",
			input:     testConfigTestData["invalid - lvl2 no child3"],
			expectErr: require.Error,
			expectLen: 1,
		}, {
			name:      "invalid - lvl2 child3 below threshold",
			input:     testConfigTestData["invalid - lvl2 child3 below threshold"],
			expectErr: require.Error,
			expectLen: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Configure mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(constants.EtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, constants.EtcDir()+testFilename, []byte(testCase.input), 0644),
				"Failed to write in memory file")

			// Load from mock filesystem.
			actual := &config{}
			err := Load(fs, actual, testFilename, testPrefix, "yaml")
			testCase.expectErr(t, err)

			validationError := &validator.ValidationError{}
			if errors.As(err, &validationError) {
				require.Lenf(t, validationError.Errors, testCase.expectLen, "Expected errors count is incorrect: %v", err)

				return
			}

			// Load expected struct.
			expected := &config{}
			require.NoError(t, yaml.Unmarshal([]byte(testCase.input), expected), "failed to unmarshal expected constants")
			require.True(t, reflect.DeepEqual(expected, actual))

			// Test configuring of environment variable.
			lvl1Child1 := xid.New().String()
			lvl1Child2 := 49
			lvl2Child1 := 29
			lvl2Child3 := xid.New().String()
			t.Setenv(keyspaceLevel1+"CHILDONE", lvl1Child1)
			t.Setenv(keyspaceLevel1+"CHILDTWO", strconv.Itoa(lvl1Child2))
			t.Setenv(keyspaceLevel1+"CHILDTHREE", strconv.FormatBool(false))
			t.Setenv(keyspaceLevel2+"CHILDONE", strconv.Itoa(lvl2Child1))
			t.Setenv(keyspaceLevel2+"CHILDTWO", strconv.FormatBool(false))
			t.Setenv(keyspaceLevel2+"CHILDTHREE", lvl2Child3)
			err = Load(fs, actual, testFilename, testPrefix, "yaml")
			require.NoErrorf(t, err, "failed to load constants file: %v", err)
			require.Equal(t, lvl1Child1, actual.ParentLevel1.ChildOne, "failed to load level 1 child 1")
			require.Equal(t, lvl1Child2, actual.ParentLevel1.ChildTwo, "failed to load level 1 child 2")
			require.False(t, actual.ParentLevel1.ChildThree, "failed to load level 1 child 2")
			require.Equal(t, lvl2Child1, actual.ParentLevel2.ChildOne, "failed to load level 2 child 1")
			require.False(t, actual.ParentLevel2.ChildTwo, "failed to load level 2 child 2")
			require.Equal(t, lvl2Child3, actual.ParentLevel2.ChildThree, "failed to load level 2 child 2")
		})
	}
}
