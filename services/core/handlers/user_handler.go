package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/teammachinist/tutuplapak/internal"
	"github.com/teammachinist/tutuplapak/services/core/models"
	"github.com/teammachinist/tutuplapak/services/core/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService services.UserServiceInterface
}

func NewUserHandler(userService services.UserServiceInterface) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) LinkPhone(c *fiber.Ctx) error {
	var req models.LinkPhoneRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	validate := validator.New()
	validate.RegisterValidation("phone_format", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		if phone == "" {
			return false
		}

		// Check if phone starts with "+"
		if !strings.HasPrefix(phone, "+") {
			return false
		}

		// Validate phone format (international format)
		phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
		return phoneRegex.MatchString(phone)
	})

	if err := validate.Struct(req); err != nil {
		var details []string
		for _, ve := range err.(validator.ValidationErrors) {
			switch ve.Tag() {
			case "required":
				details = append(details, "phone is required")
			case "phone_format":
				details = append(details, "phone must start with international calling code prefix '+' and be valid format")
			default:
				details = append(details, "phone is not valid")
			}
		}
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation error",
			"details": details,
		})
	}

	// Get user ID from JWT token (already validated by middleware)
	userIDStr, ok := internal.GetUserIDFromFiber(c)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header required",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid user id in token",
		})
	}

	// Call service to link phone
	resp, err := h.userService.LinkPhone(c.Context(), userID, req.Phone)
	if err != nil {
		switch err.Error() {
		case "phone is taken":
			return c.Status(http.StatusConflict).JSON(fiber.Map{
				"error": "Phone is taken",
			})
		case "phone must start with international calling code prefix '+'":
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Phone must start with international calling code prefix '+'",
			})
		case "invalid phone number format":
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid phone number format",
			})
		default:
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}
	}

	return c.Status(http.StatusOK).JSON(resp)
}
