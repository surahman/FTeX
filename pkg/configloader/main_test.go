package configloader

// graphqlConfigTestData is a map of HTTP GraphQL configuration test data.
var testConfigTestData = configTestData()

// testFilename name of test config file.
var testFilename = "TestFileName"

// testPrefix is the testing prefix for environment variables.
var testPrefix = "TESTPREFIX"

// config is a test configuration container.
type config struct {
	ParentLevel1 `json:"parent_level_1" yaml:"parent_level_1" mapstructure:"parent_level_1" validate:"required"`
	ParentLevel2 `json:"parent_level_2" yaml:"parent_level_2" mapstructure:"parent_level_2" validate:"required"`
}

// ParentLevel1 is an embedded test struct.
// It is best not to use nested struct definitions for JSON and YAML un/marshalling.
type ParentLevel1 struct {
	ChildOne   string `json:"child_one" yaml:"child_one" mapstructure:"child_one" validate:"required"`
	ChildTwo   int    `json:"child_two" yaml:"child_two" mapstructure:"child_two" validate:"required,min=10"`
	ChildThree bool   `json:"child_three" yaml:"child_three" mapstructure:"child_three"`
}

// ParentLevel2 is an embedded test struct.
// It is best not to use nested struct definitions for JSON and YAML un/marshalling.
type ParentLevel2 struct {
	ChildOne   int    `json:"child_one" yaml:"child_one" mapstructure:"child_one" validate:"min=3"`
	ChildTwo   bool   `json:"child_two" yaml:"child_two" mapstructure:"child_two"`
	ChildThree string `json:"child_three" yaml:"child_three" mapstructure:"child_three" validate:"min=3"`
}
