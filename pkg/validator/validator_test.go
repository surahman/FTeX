package validator

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorField_Error(t *testing.T) {
	const errorStr = "Field: %s, Tag: %s, Value: %s\n"
	testCases := []struct {
		name     string
		input    *FieldError
		expected string
	}{
		{
			"Empty error",
			&FieldError{},
			fmt.Sprintf(errorStr, "", "", "%!s(<nil>)"),
		}, {
			"Field only error",
			&FieldError{Field: "field"},
			fmt.Sprintf(errorStr, "field", "", "%!s(<nil>)"),
		}, {
			"Field and Tag only error",
			&FieldError{Field: "field", Tag: "tag"},
			fmt.Sprintf(errorStr, "field", "tag", "%!s(<nil>)"),
		}, {
			"Field, Tag, and Value error",
			&FieldError{Field: "field", Tag: "tag", Value: "value"},
			fmt.Sprintf(errorStr, "field", "tag", "value"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.expected, testCase.input.Error())
		})
	}
}

func TestErrorValidation_Error(t *testing.T) {
	genExpected := func(errs ...string) string {
		var buffer bytes.Buffer
		for _, item := range errs {
			buffer.WriteString(item)
		}

		return buffer.String()
	}

	const errorStr = "Field: %s, Tag: %s, Value: %s\n"

	testCases := []struct {
		name     string
		input    *ValidationError
		expected string
	}{
		{
			"Empty error",
			&ValidationError{Errors: []*FieldError{}},
			genExpected(""),
		}, {
			"Single error",
			&ValidationError{Errors: []*FieldError{
				{Field: "field 1", Tag: "tag 1", Value: "value 1"},
			}},
			genExpected(fmt.Sprintf(errorStr, "field 1", "tag 1", "value 1")),
		}, {
			"Two errors",
			&ValidationError{Errors: []*FieldError{
				{Field: "field 1", Tag: "tag 1", Value: "value 1"},
				{Field: "field 2", Tag: "tag 2", Value: "value 2"},
			}},
			genExpected(fmt.Sprintf(errorStr, "field 1", "tag 1", "value 1"),
				fmt.Sprintf(errorStr, "field 2", "tag 2", "value 2")),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.expected, testCase.input.Error())
		})
	}
}

func TestValidateStruct(t *testing.T) {
	type ValidationTestStruct struct {
		AlphaNum string `validate:"required,alphanum,min=6"`
		Integer  int    `validate:"required,numeric,min=1,max=10"`
		Vector   []int  `validate:"required,min=1,max=3,dive,min=2,max=10"`
	}

	testCases := []struct {
		name        string
		input       *ValidationTestStruct
		expectedLen int
		expectErr   require.ErrorAssertionFunc
	}{
		{
			"No error",
			&ValidationTestStruct{AlphaNum: "alphanum1", Integer: 5, Vector: []int{2, 3, 4}},
			0,
			require.NoError,
		}, {
			"All missing",
			&ValidationTestStruct{},
			3,
			require.Error,
		}, {
			"Alphanum error",
			&ValidationTestStruct{AlphaNum: "alpha # 1", Integer: 5, Vector: []int{2, 3, 4}},
			1,
			require.Error,
		}, {
			"Integer range error",
			&ValidationTestStruct{AlphaNum: "alphanum1", Integer: 11, Vector: []int{2, 3, 4}},
			1,
			require.Error,
		}, {
			"Vector range error",
			&ValidationTestStruct{AlphaNum: "alphanum1", Integer: 5, Vector: []int{2, 1, 4}},
			1,
			require.Error,
		}, {
			"Vector range error",
			&ValidationTestStruct{AlphaNum: "alphanum1", Integer: 5, Vector: []int{2, 3, 4, 5}},
			1,
			require.Error,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := ValidateStruct(testCase.input)
			testCase.expectErr(t, err)

			validationErr := &ValidationError{}
			if errors.As(err, &validationErr) {
				require.Equal(t, testCase.expectedLen, len(validationErr.Errors))
			}
		})
	}
}
