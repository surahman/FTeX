package validator

import (
	"bytes"
	"fmt"
)

// ErrorField contains information on JSON validation errors.
type ErrorField struct {
	Field string `json:"field" yaml:"field"` // Field name where the validation error occurred.
	Tag   string `json:"tag" yaml:"tag"`     // The reason for the validation failure.
	Value any    `json:"value" yaml:"value"` // The value(s) associated with the failure.
}

// Error will output the validation error for a single structs data member.
func (err *ErrorField) Error() string {
	return fmt.Sprintf("Field: %s, Tag: %s, Value: %s\n", err.Field, err.Tag, err.Value)
}

// ErrorValidation contains all the validation errors found in a struct.
type ErrorValidation struct {
	Errors []*ErrorField `json:"validation_errors" yaml:"validation_errors"` // A list of all data members that failed validation.
}

// Error will output the validation error for all struct data members.
func (err *ErrorValidation) Error() string {
	var buffer bytes.Buffer
	for _, item := range err.Errors {
		buffer.WriteString(item.Error())
	}
	return buffer.String()
}
