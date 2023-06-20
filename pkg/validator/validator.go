package validator

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

// structValidator is the validator instance that is used for structure validation.
var structValidator *validator.Validate

// init the struct validator.
func init() {
	structValidator = validator.New()
}

// ValidateStruct will validate a struct and list all deficiencies.
func ValidateStruct(body any) error {
	var (
		validationErr ValidationError
		errs          validator.ValidationErrors
	)

	if errors.As(structValidator.Struct(body), &errs) {
		for _, issue := range errs {
			var ev FieldError
			ev.Field = issue.Field()
			ev.Tag = issue.Tag()
			ev.Value = issue.Value()
			validationErr.Errors = append(validationErr.Errors, &ev)
		}
	}

	if validationErr.Errors == nil {
		return nil
	}

	return &validationErr
}
