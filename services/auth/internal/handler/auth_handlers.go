package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/teammachinist/tutuplapak/services/auth/internal/logger"
	"github.com/teammachinist/tutuplapak/services/auth/internal/model"
	"github.com/teammachinist/tutuplapak/services/auth/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) LoginByPhone(c *gin.Context) {
	ctx := logger.WithRequestID(c.Request.Context())
	logger.InfoCtx(ctx, "Login attempt", "method", "phone")

	var req model.PhoneAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WarnCtx(ctx, "Invalid request body", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	loginResponse, err := h.userService.LoginByPhone(ctx, req.Phone, req.Password)
	if err != nil {
		logger.WarnCtx(ctx, "Login failed", "method", "phone", "error", err.Error())
		h.handleAuthError(c, err)
		return
	}

	logger.InfoCtx(ctx, "Login successful", "method", "phone")
	c.JSON(http.StatusOK, loginResponse)
}

func (h *UserHandler) RegisterByPhone(c *gin.Context) {
	ctx := logger.WithRequestID(c.Request.Context())
	logger.InfoCtx(ctx, "Registration attempt", "method", "phone")

	var req model.PhoneAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WarnCtx(ctx, "Invalid request body", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.userService.RegisterByPhone(ctx, req.Phone, req.Password)
	if err != nil {
		logger.WarnCtx(ctx, "Registration failed", "method", "phone", "error", err.Error())
		h.handleAuthError(c, err)
		return
	}

	logger.InfoCtx(ctx, "Registration successful", "method", "phone")
	c.JSON(http.StatusCreated, resp)
}

func (h *UserHandler) LoginWithEmail(c *gin.Context) {
	ctx := logger.WithRequestID(c.Request.Context())
	logger.InfoCtx(ctx, "Login attempt", "method", "email")

	var req model.EmailAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WarnCtx(ctx, "Invalid request body", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	loginResponse, err := h.userService.LoginWithEmail(ctx, req.Email, req.Password)
	if err != nil {
		logger.WarnCtx(ctx, "Login failed", "method", "email", "error", err.Error())
		h.handleAuthError(c, err)
		return
	}

	logger.InfoCtx(ctx, "Login successful", "method", "email")
	c.JSON(http.StatusOK, loginResponse)
}

func (h *UserHandler) RegisterWithEmail(c *gin.Context) {
	ctx := logger.WithRequestID(c.Request.Context())
	logger.InfoCtx(ctx, "Registration attempt", "method", "email")

	var req model.EmailAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WarnCtx(ctx, "Invalid request body", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.userService.RegisterWithEmail(ctx, req.Email, req.Password)
	if err != nil {
		logger.WarnCtx(ctx, "Registration failed", "method", "email", "error", err.Error())
		h.handleAuthError(c, err)
		return
	}

	logger.InfoCtx(ctx, "Registration successful", "method", "email")
	c.JSON(http.StatusCreated, resp)
}

// Helper to reduce duplication and centralize error handling
func (h *UserHandler) handleAuthError(c *gin.Context, err error) {
	errorMsg := err.Error()

	switch errorMsg {
	case "email not found", "phone not found":
		c.JSON(http.StatusNotFound, gin.H{"error": errorMsg})
	case "invalid credentials":
		c.JSON(http.StatusUnauthorized, gin.H{"error": errorMsg})
	case "email address already exists", "phone number already exists":
		c.JSON(http.StatusConflict, gin.H{"error": errorMsg})
	default:
		// Check for validation errors (format/length issues)
		if strings.Contains(errorMsg, "format") ||
			strings.Contains(errorMsg, "characters") ||
			strings.Contains(errorMsg, "required") ||
			strings.Contains(errorMsg, "must start") {
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		} else {
			// Unknown error - could be system issue
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
	}
}
