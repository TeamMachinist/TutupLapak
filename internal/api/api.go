package api

import (
	"regexp"
	"strings"
)

// Standard API response format
type Response struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Response constructors
func Success(data interface{}) Response {
	return Response{Status: "success", Data: data}
}

func Error(message string) Response {
	return Response{Status: "error", Error: message}
}

func ValidationFailed(errors []ValidationError) Response {
	return Response{
		Status: "error",
		Error:  "validation failed",
		Data:   ValidationErrors{Errors: errors},
	}
}

// Common validation functions
func ValidateEmail(email string) *ValidationError {
	if email == "" {
		return &ValidationError{Field: "email", Message: "email is required"}
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return &ValidationError{Field: "email", Message: "invalid email format"}
	}

	return nil
}

func ValidatePhone(phone string) *ValidationError {
	if phone == "" {
		return &ValidationError{Field: "phone", Message: "phone is required"}
	}

	if !strings.HasPrefix(phone, "+") {
		return &ValidationError{Field: "phone", Message: "phone must start with international calling code prefix '+'"}
	}

	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	if !phoneRegex.MatchString(phone) {
		return &ValidationError{Field: "phone", Message: "invalid phone number format"}
	}

	return nil
}

func ValidatePassword(password string) *ValidationError {
	if password == "" {
		return &ValidationError{Field: "password", Message: "password is required"}
	}

	if len(password) < 8 {
		return &ValidationError{Field: "password", Message: "password must be at least 8 characters long"}
	}

	if len(password) > 32 {
		return &ValidationError{Field: "password", Message: "password must be at most 32 characters long"}
	}

	return nil
}

func ValidateRequired(field, value string) *ValidationError {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{Field: field, Message: field + " is required"}
	}
	return nil
}

// Helper to validate multiple fields
func ValidateFields(validations ...func() *ValidationError) []ValidationError {
	var errors []ValidationError

	for _, validate := range validations {
		if err := validate(); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}
