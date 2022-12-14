package validator

import (
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
	var validationErr ErrorValidation
	if err := structValidator.Struct(body); err != nil {
		for _, issue := range err.(validator.ValidationErrors) {
			var ev ErrorField
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
