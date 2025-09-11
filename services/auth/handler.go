package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"
	
	"github.com/go-playground/validator/v10"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService *service.AuthService
	validator   *validator.Validate // Reuse validator instance
	emailRegex  *regexp.Regexp      // Compile regex once
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
		emailRegex:  regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`),
	}
}

// RegisterWithEmail - Uses value passing for registration data
// Value passing is preferred here because:
// 1. Registration payload is relatively small
// 2. We want stack allocation for better cache locality
// 3. No GC overhead for short-lived registration data
// 4. Automatic cleanup when function returns
func (h *AuthHandler) RegisterWithEmail(c *gin.Context) {
	requestCtx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Use value passing - entity will be copied to stack for better performance
	var payload model.User
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrBadRequest.Error()})
		return
	}

	// Validate using struct validation
	if err := h.validator.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrBadRequest.Error()})
		return
	}

	// Custom validation using pre-compiled regex
	if !h.isEmailValid(payload.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrBadRequest.Error()})
		return
	}

	if len(payload.Password) < 8 || len(payload.Password) > 32 {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrBadRequest.Error()})
		return
	}

	// Pass by value to service layer for stack allocation benefits
	// The service will copy it once more, but registration is infrequent
	// and we benefit from predictable cleanup and better cache locality
	user, err := h.authService.RegisterWithEmail(requestCtx, payload)
	if err != nil {
		// Handle different types of service layer errors
		switch err {
		case errors.StatusBadRequest:
			c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrBadRequest.Error()})
		case errors.ErrConflict:
			c.JSON(http.StatusConflict, gin.H{"error": errors.ErrConflict.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInternalServerError.Error()})
		}
		return
	}

	// Check context cancellation
	if requestCtx.Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": errors.ErrRequestTimeout.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

// Login - Uses value passing for authentication data
// Value passing is optimal for login because:
// 1. Login payload is small (email + password)
// 2. Authentication happens frequently - stack allocation is faster
// 3. No heap fragmentation from frequent small allocations
// 4. Better CPU cache efficiency for hot path
// 5. Automatic memory cleanup reduces GC pressure
func (h *AuthHandler) LoginWithEmail(c *gin.Context) {
	requestCtx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Use value passing - login credentials are small and accessed frequently
	var payload model.User
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrBadRequest.Error()})
		return
	}

	// Basic validation for required fields
	if payload.Email == "" || payload.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrBadRequest.Error()})
		return
	}

	// Email format validation using pre-compiled regex
	if !h.isEmailValid(payload.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrBadRequest.Error()})
		return
	}

	// Pass by value to service layer
	// Login is a hot path, so we optimize for:
	// - Stack allocation (faster than heap)
	// - Direct memory access (no indirection)
	// - Automatic cleanup (no GC needed)
	user, err := h.authService.LoginWithEmail(requestCtx, payload)
	if err != nil {
		// Handle different types of authentication errors
		switch err {
		case errors.StatusBadRequest:
			c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrBadRequest.Error()})
		case errors.ErrConflict:
			c.JSON(http.StatusConflict, gin.H{"error": errors.ErrConflict.Error()})
		case errors.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": errors.ErrNotFound.Error()})
		case errors.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": errors.ErrUnauthorized.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInternalServerError.Error()})
		}
		return
	}

	// Check context cancellation
	if requestCtx.Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": errors.ErrRequestTimeout.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// isEmailValid uses the pre-compiled regex for better performance
// Pre-compilation eliminates regex compilation overhead on each validation
func (h *AuthHandler) isEmailValid(email string) bool {
	return h.emailRegex.MatchString(email)
}
