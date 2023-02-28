package configloader

// graphqlConfigTestData is a map of HTTP GraphQL configuration test data.
var testConfigTestData = configTestData()

// testFilename name of test config file.
var testFilename = "TestFileName"

// testPrefix is the testing prefix for environment variables.
var testPrefix = "TESTPREFIX"

// config is a test configuration container.
type config struct {
	ParentLevel1 `json:"parentLevel1" yaml:"parentLevel1" mapstructure:"parentLevel1" validate:"required"`
	ParentLevel2 `json:"parentLevel2" yaml:"parentLevel2" mapstructure:"parentLevel2" validate:"required"`
}

// ParentLevel1 is an embedded test struct.
// It is best not to use nested struct definitions for JSON and YAML un/marshalling.
type ParentLevel1 struct {
	ChildOne   string `json:"childOne" yaml:"childOne" mapstructure:"childOne" validate:"required"`
	ChildTwo   int    `json:"childTwo" yaml:"childTwo" mapstructure:"childTwo" validate:"required,min=10"`
	ChildThree bool   `json:"childThree" yaml:"childThree" mapstructure:"childThree"`
}

// ParentLevel2 is an embedded test struct.
// It is best not to use nested struct definitions for JSON and YAML un/marshalling.
type ParentLevel2 struct {
	ChildOne   int    `json:"childOne" yaml:"childOne" mapstructure:"childOne" validate:"min=3"`
	ChildTwo   bool   `json:"childTwo" yaml:"childTwo" mapstructure:"childTwo"`
	ChildThree string `json:"childThree" yaml:"childThree" mapstructure:"childThree" validate:"min=3"`
}
