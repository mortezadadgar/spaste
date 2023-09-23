package validator

import (
	"fmt"
	"strings"
	"sync"
)

type Validator struct {
	errors []map[errType]string
	mu     sync.Mutex
}

// New returns a instance of Validator.
func New() *Validator {
	return &Validator{
		errors: make([]map[errType]string, 0),
	}
}

const (
	isEqualErr    = "is not equal to expected value"
	isBlankErr    = "is blank"
	validationErr = "validation failed"
)

type errType int

const (
	equalErrType errType = iota
	blankErrType
)

func (e errType) String() string {
	switch e {
	case equalErrType:
		return isEqualErr
	case blankErrType:
		return isBlankErr
	}

	return ""
}

// IsEqual reports whether integer value is equal to expectedValue.
func (v *Validator) IsEqual(name string, value int, expectedValue int) {
	if value == expectedValue {
		v.addError(equalErrType, name)
	}
}

// IsBlank reports whether string value is a empty string.
func (v *Validator) IsBlank(name string, value string) {
	if strings.TrimSpace(value) == "" {
		v.addError(blankErrType, name)
	}
}

// Valid reports whether value is valid or not.
func (v *Validator) Valid() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if len(v.errors) == 0 {
		return nil
	}

	err := v.formatError()

	v.clearErrors()

	return err
}

func (v *Validator) addError(t errType, name string) {
	error := make(map[errType]string, 1)
	error[t] = name
	v.mu.Lock()
	v.errors = append(v.errors, error)
	v.mu.Unlock()
}

func (v *Validator) clearErrors() {
	v.errors = make([]map[errType]string, 0)
}

func (v *Validator) formatError() error {
	var error string

	for i, e := range v.errors {
		for k, v := range e {
			if i > 0 {
				error += " "
			}

			error += v + ": " + k.String()

			if i != len(e)-1 {
				error += ","
			}
		}
	}

	return fmt.Errorf("%s: %v", validationErr, error)
}
