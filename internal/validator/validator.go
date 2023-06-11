package validator

import (
	"fmt"
	"strings"
)

type Validator struct {
	errors []error
}

// New returns a instance of Validator.
func New() Validator {
	return Validator{}
}

// IsEqual reports whether value is equal to expectedValue.
func (v *Validator) IsEqual(name string, value int, expectedValue int) {
	if value == expectedValue {
		v.addError(fmt.Errorf("%s: is not equal to %d", name, expectedValue))
	}
}

// IsBlank reports whether value is a empty string.
func (v *Validator) IsBlank(name string, value string) {
	if strings.TrimSpace(value) == "" {
		v.addError(fmt.Errorf("%s: is blank", name))
	}
}

// IsValid reports whether value is valid or not.
func (v *Validator) IsValid() error {
	if len(v.errors) == 0 {
		return nil
	}

	var errors string
	for _, error := range v.errors {
		errors += " " + error.Error()
	}

	return fmt.Errorf(errors)
}

func (v *Validator) addError(err error) {
	v.errors = append(v.errors, err)
}
