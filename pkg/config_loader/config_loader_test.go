package config_loader

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
	keyspaceLevel1 := testPrefix + "_PARENT_LEVEL_1."
	keyspaceLevel2 := testPrefix + "_PARENT_LEVEL_2."

	testCases := []struct {
		name      string
		input     string
		expectErr require.ErrorAssertionFunc
		expectLen int
	}{
		// ----- test cases start ----- //
		{
			"empty - etc dir",
			testConfigTestData["empty"],
			require.Error,
			4,
		}, {
			"valid - etc dir",
			testConfigTestData["valid"],
			require.NoError,
			0,
		}, {
			"valid - required",
			testConfigTestData["valid - required"],
			require.NoError,
			0,
		}, {
			"invalid - no lvl 1",
			testConfigTestData["invalid - no lvl 1"],
			require.Error,
			2,
		}, {
			"invalid - no lvl1 child1",
			testConfigTestData["invalid - no lvl1 child1"],
			require.Error,
			1,
		}, {
			"invalid - lvl1 child1 below threshold",
			testConfigTestData["invalid - lvl1 child1 below threshold"],
			require.Error,
			1,
		}, {
			"invalid - lvl2 no child1",
			testConfigTestData["invalid - lvl2 no child1"],
			require.Error,
			1,
		}, {
			"invalid - lvl2 child1 below threshold",
			testConfigTestData["invalid - lvl2 child1 below threshold"],
			require.Error,
			1,
		}, {
			"invalid - lvl2 no child3",
			testConfigTestData["invalid - lvl2 no child3"],
			require.Error,
			1,
		}, {
			"invalid - lvl2 child3 below threshold",
			testConfigTestData["invalid - lvl2 child3 below threshold"],
			require.Error,
			1,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Configure mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+testFilename, []byte(testCase.input), 0644), "Failed to write in memory file")

			// Load from mock filesystem.
			actual := &config{}
			err := ConfigLoader(fs, actual, testFilename, testPrefix, "yaml")
			testCase.expectErr(t, err)

			validationError := &validator.ValidationError{}
			if errors.As(err, &validationError) {
				require.Equalf(t, testCase.expectLen, len(validationError.Errors), "Expected errors count is incorrect: %v", err)
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
			t.Setenv(keyspaceLevel1+"CHILD_ONE", lvl1Child1)
			t.Setenv(keyspaceLevel1+"CHILD_TWO", strconv.Itoa(lvl1Child2))
			t.Setenv(keyspaceLevel1+"CHILD_THREE", strconv.FormatBool(false))
			t.Setenv(keyspaceLevel2+"CHILD_ONE", strconv.Itoa(lvl2Child1))
			t.Setenv(keyspaceLevel2+"CHILD_TWO", strconv.FormatBool(false))
			t.Setenv(keyspaceLevel2+"CHILD_THREE", lvl2Child3)
			err = ConfigLoader(fs, actual, testFilename, testPrefix, "yaml")
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
