// Package validation provides validation functionality for the application
package validation

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// Initialize sets up the validator with custom validations
func Initialize() error {
	validate = validator.New()

	// Register custom validation tags
	if err := validate.RegisterValidation("url", validateURL); err != nil {
		return err
	}
	if err := validate.RegisterValidation("no_spaces", validateNoSpaces); err != nil {
		return err
	}
	if err := validate.RegisterValidation("versionformat", validateVersionFormat); err != nil {
		return err
	}
	if err := validate.RegisterValidation("password_complexity", validatePasswordComplexity); err != nil {
		return err
	}

	// Register custom error messages
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		return name
	})

	return nil
}

// GetValidator returns the validator instance
func GetValidator() *validator.Validate {
	if validate == nil {
		if err := Initialize(); err != nil {
			return nil
		}
	}

	return validate
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Validate validates a struct and returns a slice of validation errors
func Validate(i interface{}) []ValidationError {
	if validate == nil {
		Initialize()
	}

	var errors []ValidationError

	// Use the validator's built-in nested validation
	err := validate.Struct(i)
	if err != nil {
		for _, validationErr := range err.(validator.ValidationErrors) {
			errors = append(errors, createValidationError(validationErr))
		}
	}

	return errors
}

// createValidationError creates a ValidationError from a validator.ValidationError
func createValidationError(err validator.FieldError) ValidationError {
	cfg := config.GetInstance().Get()

	var message string
	switch err.Tag() {
	case "required":
		message = fmt.Sprintf("%s is required", err.Field())
	case "email":
		message = fmt.Sprintf("%s must be a valid email address", err.Field())
	case "url":
		message = fmt.Sprintf("%s must be a valid URL", err.Field())
	case "min":
		message = fmt.Sprintf("%s must be at least %s characters long", err.Field(), err.Param())
	case "max":
		message = fmt.Sprintf("%s must be at most %s characters long", err.Field(), err.Param())
	case "no_spaces":
		message = fmt.Sprintf("%s cannot contain spaces", err.Field())
	case "versionformat":
		message = fmt.Sprintf("%s must be in the format x.x.x", err.Field())
	case "password_complexity":
		minLength := cfg.GetInt(config.SecurityPasswordMinLengthKey, 8)
		requireNumber := cfg.GetBool(config.SecurityPasswordRequireNumberKey, true)
		requireSpecial := cfg.GetBool(config.SecurityPasswordRequireSpecialKey, true)
		requireUppercase := cfg.GetBool(config.SecurityPasswordRequireUppercaseKey, true)
		msg := fmt.Sprintf("%s must be at least %v characters long", err.Field(), minLength)
		if requireNumber {
			msg += " and contain at least one number"
		}
		if requireSpecial {
			msg += " and contain at least one special character from the following: " + config.PasswordAllowedSpecialChars
		}
		if requireUppercase {
			msg += " and contain at least one uppercase letter"
		}
		message = msg
	default:
		message = fmt.Sprintf("%s failed validation: %s", err.Field(), err.Tag())
	}

	// The validator library automatically provides the full field path
	// e.g., "Services[0].ServiceName" or "Files[1].FileName"
	fieldPath := err.Field()

	return ValidationError{
		Field:   fieldPath,
		Message: message,
	}
}

// Custom validators

func validateURL(fl validator.FieldLevel) bool {
	field := fl.Field().String()
	if field == "" {
		return true // empty values are handled by 'required' tag
	}

	u, err := url.Parse(field)
	if err != nil {
		return false
	}

	// Require scheme and host
	return u.Scheme != "" && u.Host != ""
}

func validateNoSpaces(fl validator.FieldLevel) bool {
	return !strings.Contains(fl.Field().String(), " ")
}

// Helper functions for common validations

// ValidateAndReturn validates a struct and returns the errors or nil
// This is useful in HTTP handlers to quickly validate and return errors
func ValidateAndReturn(i interface{}) error {
	errors := Validate(i)
	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %v", errors)
	}
	return nil
}

// IsValid checks if a struct is valid without returning specific errors
func IsValid(i interface{}) bool {
	return len(Validate(i)) == 0
}

func validateVersionFormat(fl validator.FieldLevel) bool {
	field := fl.Field().String()
	if field == "" {
		return true // empty values are handled by 'required' tag
	}

	versionRegex := `^[0-9]+\.[0-9]+\.[0-9]+$`
	match := regexp.MustCompile(versionRegex).MatchString(field)

	return match
}

func validatePasswordComplexity(fl validator.FieldLevel) bool {
	field := fl.Field().String()
	if field == "" {
		return true // empty values are handled by 'required' tag
	}
	cfg := config.GetInstance().Get()
	if cfg == nil {
		return false
	}

	minLength := cfg.GetInt(config.SecurityPasswordMinLengthKey, 8)
	requireNumber := cfg.GetBool(config.SecurityPasswordRequireNumberKey, true)
	requireSpecial := cfg.GetBool(config.SecurityPasswordRequireSpecialKey, true)
	requireUppercase := cfg.GetBool(config.SecurityPasswordRequireUppercaseKey, true)

	if requireNumber {
		if !regexp.MustCompile(`[0-9]`).MatchString(field) {
			return false
		}
	}

	if requireSpecial {
		if !regexp.MustCompile(`[` + config.PasswordAllowedSpecialChars + `]`).MatchString(field) {
			return false
		}
	}

	if requireUppercase {
		if !regexp.MustCompile(`[A-Z]`).MatchString(field) {
			return false
		}
	}

	if len(field) < minLength {
		return false
	}

	return true
}
