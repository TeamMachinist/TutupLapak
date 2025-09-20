package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/teammachinist/tutuplapak/services/core/internal/logger"
)

// Validation errors (only for validation failed responses)
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// HTTP response writers - return data directly as per spec
func WriteJSON(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.ErrorCtx(r.Context(), "Failed to encode JSON response", "error", err, "status_code", statusCode)
		// Fallback to basic error response
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Success responses (200, 201) - return data directly
func WriteSuccess(w http.ResponseWriter, r *http.Request, data interface{}) {
	WriteJSON(w, r, http.StatusOK, data)
}

func WriteCreated(w http.ResponseWriter, r *http.Request, data interface{}) {
	WriteJSON(w, r, http.StatusCreated, data)
}

// Error responses with simple message
func WriteError(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	errorResponse := map[string]string{"error": message}
	WriteJSON(w, r, statusCode, errorResponse)
}

// Validation error response (400) - special format for validation
func WriteValidationError(w http.ResponseWriter, r *http.Request, errors []ValidationError) {
	validationResponse := ValidationErrors{Errors: errors}
	WriteJSON(w, r, http.StatusBadRequest, validationResponse)
}

// Most commonly used error responses
func WriteBadRequest(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusBadRequest, message)
}

func WriteNotFound(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusNotFound, message)
}

func WriteUnauthorized(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusUnauthorized, message)
}

func WriteInternalServerError(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusInternalServerError, message)
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
