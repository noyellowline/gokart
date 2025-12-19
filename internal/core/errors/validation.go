package errors

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// FormatValidation formats validator.ValidationErrors into a readable error message.
// If err is not a validator.ValidationErrors, returns it unchanged.
func FormatValidation(err error) error {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	var messages []string
	for _, e := range validationErrors {
		messages = append(messages,
			fmt.Sprintf("%s: %s (value: %v)", e.Field(), e.Tag(), e.Value()))
	}

	return fmt.Errorf("validation failed:\n  - %s",
		strings.Join(messages, "\n  - "))
}
