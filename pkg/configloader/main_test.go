package configloader

// graphqlConfigTestData is a map of HTTP GraphQL configuration test data.
var testConfigTestData = configTestData()

// testFilename name of test config file.
var testFilename = "TestFileName"

// testPrefix is the testing prefix for environment variables.
var testPrefix = "TESTPREFIX"

// config is a test configuration container.
type config struct {
	ParentLevel1 `json:"parentLevel1" mapstructure:"parentLevel1" validate:"required" yaml:"parentLevel1"`
	ParentLevel2 `json:"parentLevel2" mapstructure:"parentLevel2" validate:"required" yaml:"parentLevel2"`
}

// ParentLevel1 is an embedded test struct.
// It is best not to use nested struct definitions for JSON and YAML un/marshalling.
type ParentLevel1 struct {
	ChildOne   string `json:"childOne"   mapstructure:"childOne"   validate:"required"        yaml:"childOne"`
	ChildTwo   int    `json:"childTwo"   mapstructure:"childTwo"   validate:"required,min=10" yaml:"childTwo"`
	ChildThree bool   `json:"childThree" mapstructure:"childThree" yaml:"childThree"`
}

// ParentLevel2 is an embedded test struct.
// It is best not to use nested struct definitions for JSON and YAML un/marshalling.
type ParentLevel2 struct {
	ChildOne   int    `json:"childOne"   mapstructure:"childOne"   validate:"min=3" yaml:"childOne"`
	ChildTwo   bool   `json:"childTwo"   mapstructure:"childTwo"   yaml:"childTwo"`
	ChildThree string `json:"childThree" mapstructure:"childThree" validate:"min=3" yaml:"childThree"`
}
