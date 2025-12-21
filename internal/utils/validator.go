package utils

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidEmail = errors.New("invalid email address")
	ErrEmptyField   = errors.New("field cannot be empty")
)

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

// ValidateRequired checks if a string field is not empty
func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(fieldName + " is required")
	}
	return nil
}

// ValidateMinLength checks minimum length
func ValidateMinLength(value string, minLength int, fieldName string) error {
	if len(value) < minLength {
		return errors.New(fieldName + " must be at least " + string(rune(minLength)) + " characters")
	}
	return nil
}

// ValidateMaxLength checks maximum length
func ValidateMaxLength(value string, maxLength int, fieldName string) error {
	if len(value) > maxLength {
		return errors.New(fieldName + " must be at most " + string(rune(maxLength)) + " characters")
	}
	return nil
}

// SanitizeString removes extra whitespace and trims
func SanitizeString(s string) string {
	return strings.TrimSpace(s)
}
