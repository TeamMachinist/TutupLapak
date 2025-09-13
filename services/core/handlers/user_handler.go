package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/teammachinist/tutuplapak/services/core/models"
	"github.com/teammachinist/tutuplapak/services/core/services"
)

type UserHandler struct {
	userService services.UserServiceInterface
}

func NewUserHandler(userService services.UserServiceInterface) UserHandler {
	return UserHandler{userService: userService}
}

func (h *UserHandler) GetUserWithFileId(c *fiber.Ctx) error {
	var userID = uuid.MustParse("00000000-0000-0000-0000-000000000012")
	fmt.Printf("masuk handler: %s", userID.String())
	ctx := c.Context()
	rows, err := h.userService.GetUserWithFileId(ctx, userID)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(rows)
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	ctx := c.Context()

	userIdStr := c.Params("id")
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product id format",
		})
	}

	var req models.UserRequest
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	rows, err := h.userService.UpdateUser(ctx, userId, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Server Error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(rows)
}
