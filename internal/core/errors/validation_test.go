package errors

import (
	"errors"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestFormatValidation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
		wantMsg []string
	}{
		{
			name: "valid struct",
			input: struct {
				Name string `validate:"required"`
			}{Name: "test"},
			wantErr: false,
		},
		{
			name: "missing required field",
			input: struct {
				Name string `validate:"required"`
			}{},
			wantErr: true,
			wantMsg: []string{"Name", "required"},
		},
		{
			name: "invalid email",
			input: struct {
				Email string `validate:"email"`
			}{Email: "invalid"},
			wantErr: true,
			wantMsg: []string{"Email", "email"},
		},
		{
			name: "multiple errors",
			input: struct {
				Name  string `validate:"required"`
				Email string `validate:"email"`
			}{},
			wantErr: true,
			wantMsg: []string{"Name", "required", "Email", "email"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.input)
			result := FormatValidation(err)

			if tt.wantErr {
				if result == nil {
					t.Fatal("expected error, got nil")
				}
				msg := result.Error()
				for _, want := range tt.wantMsg {
					if !strings.Contains(msg, want) {
						t.Errorf("error message missing %q: %s", want, msg)
					}
				}
			} else if result != nil {
				t.Errorf("expected no error, got: %v", result)
			}
		})
	}
}

func TestFormatValidation_NonValidatorError(t *testing.T) {
	err := errors.New("generic error")
	result := FormatValidation(err)

	if result != err {
		t.Errorf("expected same error back, got different: %v", result)
	}
}
