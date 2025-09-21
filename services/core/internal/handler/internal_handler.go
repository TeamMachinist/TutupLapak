package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/teammachinist/tutuplapak/services/core/internal/logger"
	"github.com/teammachinist/tutuplapak/services/core/internal/model"
	"github.com/teammachinist/tutuplapak/services/core/internal/service"
)

type InternalHandler struct {
	userService service.UserServiceInterface
}

func NewInternalHandler(userService service.UserServiceInterface) *InternalHandler {
	return &InternalHandler{userService: userService}
}

// CreateUserFromAuth - Called by Auth service to create user record
func (h *InternalHandler) CreateUserFromAuth(c *fiber.Ctx) error {
	ctx := logger.WithRequestID(c.Context())
	logger.InfoCtx(ctx, "Create user from auth request")

	var req model.CreateUserFromAuthRequest
	if err := c.BodyParser(&req); err != nil {
		logger.WarnCtx(ctx, "Invalid create user request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate required fields
	if req.UserAuthID == "" {
		logger.WarnCtx(ctx, "Missing user auth ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user_auth_id is required",
		})
	}

	// Create user
	response, err := h.userService.CreateUserFromAuth(ctx, req)
	if err != nil {
		logger.WarnCtx(ctx, "Failed to create user from auth",
			"user_auth_id", req.UserAuthID,
			"error", err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.InfoCtx(ctx, "User created successfully from auth",
		"user_id", response.ID,
		"user_auth_id", response.UserAuthID)

	return c.Status(fiber.StatusCreated).JSON(response)
}

// CreateUserFromAuth - Called by Auth service to create user record
func (h *InternalHandler) GetUserFromAuth(c *fiber.Ctx) error {
	ctx := logger.WithRequestID(c.Context())
	logger.InfoCtx(ctx, "Get user from auth request")

	userAuthID := c.Params("userAuthID")

	userAuthUUID, err := uuid.Parse(userAuthID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user auth ID"})
	}

	user, err := h.userService.GetUserFromAuth(ctx, userAuthUUID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	return c.JSON(fiber.Map{
		"id":           user.ID,
		"user_auth_id": user.UserAuthID,
	})
}
