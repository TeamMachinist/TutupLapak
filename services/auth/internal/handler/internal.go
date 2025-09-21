package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/teammachinist/tutuplapak/services/auth/internal/logger"
	"github.com/teammachinist/tutuplapak/services/auth/internal/service"
	"github.com/teammachinist/tutuplapak/services/auth/pkg/authz"
)

type InternalHandler struct {
	userService *service.UserService
}

func NewInternalHandler(userService *service.UserService) *InternalHandler {
	return &InternalHandler{
		userService: userService,
	}
}

// UpdateUserAuthRequest for updating user auth info
type UpdateUserAuthRequest struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
}

// ValidateToken - For other services to validate tokens via HTTP
// POST /internal/validate
func (h *InternalHandler) ValidateToken(c *gin.Context) {
	ctx := logger.WithRequestID(c.Request.Context())

	var req authz.ValidateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WarnCtx(ctx, "Invalid token validation request", "error", err.Error())
		c.JSON(http.StatusBadRequest, authz.ValidateTokenResponse{
			Valid: false,
			Error: "invalid request",
		})
		return
	}

	claims, err := h.userService.ValidateTokenInternal(req.Token)
	if err != nil {
		logger.DebugCtx(ctx, "Token validation failed", "error", err.Error())
		c.JSON(http.StatusOK, authz.ValidateTokenResponse{
			Valid: false,
			Error: err.Error(),
		})
		return
	}

	logger.DebugCtx(ctx, "Token validation successful", "user_auth_id", claims.UserID)
	c.JSON(http.StatusOK, authz.ValidateTokenResponse{
		Valid:  true,
		UserID: claims.UserID,
	})
}

// UpdateUserAuth - Update user auth phone or email
// PUT /internal/userauth/:userAuthID
func (h *InternalHandler) UpdateUserAuth(c *gin.Context) {
	ctx := logger.WithRequestID(c.Request.Context())
	userAuthID := c.Param("userAuthID")

	if userAuthID == "" {
		logger.WarnCtx(ctx, "Missing user auth ID in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "user auth ID required"})
		return
	}

	var req UpdateUserAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WarnCtx(ctx, "Invalid update user auth request", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate that at least one field is provided
	if req.Phone == "" && req.Email == "" {
		logger.WarnCtx(ctx, "No update fields provided", "user_auth_id", userAuthID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone or email required"})
		return
	}

	// Handle phone update
	if req.Phone != "" {
		logger.InfoCtx(ctx, "Updating user auth phone", "user_auth_id", userAuthID)
		if err := h.userService.LinkPhone(ctx, userAuthID, req.Phone); err != nil {
			logger.WarnCtx(ctx, "Failed to update phone", "user_auth_id", userAuthID, "error", err.Error())
			h.handleUpdateError(c, err)
			return
		}
		logger.InfoCtx(ctx, "Phone updated successfully", "user_auth_id", userAuthID)
	}

	// Handle email update
	if req.Email != "" {
		logger.InfoCtx(ctx, "Updating user auth email", "user_auth_id", userAuthID)
		if err := h.userService.LinkEmail(ctx, userAuthID, req.Email); err != nil {
			logger.WarnCtx(ctx, "Failed to update email", "user_auth_id", userAuthID, "error", err.Error())
			h.handleUpdateError(c, err)
			return
		}
		logger.InfoCtx(ctx, "Email updated successfully", "user_auth_id", userAuthID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated successfully"})
}

// Helper to handle update errors
func (h *InternalHandler) handleUpdateError(c *gin.Context, err error) {
	errorMsg := err.Error()

	switch errorMsg {
	case "user not found":
		c.JSON(http.StatusNotFound, gin.H{"error": errorMsg})
	case "email address already exists", "phone number already exists":
		c.JSON(http.StatusConflict, gin.H{"error": errorMsg})
	case "invalid email address format", "invalid phone number format",
		"phone must start with international calling code prefix '+'":
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

// RegisterInternalRoutes - Register internal API routes
func (h *InternalHandler) RegisterInternalRoutes(router *gin.Engine) {
	internal := router.Group("/internal")
	{
		internal.POST("/validate", h.ValidateToken)
		internal.PUT("/userauth/:userAuthID", h.UpdateUserAuth)
	}
}
