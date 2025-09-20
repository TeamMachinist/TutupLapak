package handler

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/teammachinist/tutuplapak/internal"
	"github.com/teammachinist/tutuplapak/services/core/internal/model"
	"github.com/teammachinist/tutuplapak/services/core/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService service.UserServiceInterface
}

func NewUserHandler(userService service.UserServiceInterface) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) LinkPhone(c *fiber.Ctx) error {
	var req model.LinkPhoneRequest
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
	// TODO: Use authz package
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

func (h *UserHandler) LinkEmail(c *fiber.Ctx) error {
	var req model.LinkEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	validate := validator.New()
	validate.RegisterValidation("email_format", func(fl validator.FieldLevel) bool {
		email := fl.Field().String()
		if email == "" {
			return false
		}

		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		return emailRegex.MatchString(email)
	})

	if err := validate.Struct(req); err != nil {
		var details []string
		for _, ve := range err.(validator.ValidationErrors) {
			switch ve.Tag() {
			case "required":
				details = append(details, "email is required")
			case "email_format":
				details = append(details, "email must be in valid email format")
			default:
				details = append(details, "email is not valid")
			}
		}
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation error",
			"details": details,
		})
	}

	// TODO: Use authz package
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

	resp, err := h.userService.LinkEmail(c.Context(), userID, req.Email)
	if err != nil {
		switch err.Error() {
		case "email is taken":
			return c.Status(http.StatusConflict).JSON(fiber.Map{
				"error": "Email is taken",
			})
		case "email is required":
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Email is required",
			})
		case "invalid email format":
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid email format",
			})
		default:
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}
	}

	return c.Status(http.StatusOK).JSON(resp)
}

func (h *UserHandler) GetUserWithFileId(c *fiber.Ctx) error {
	// TODO: Use authz package
	userIDStr, ok := internal.GetUserIDFromFiber(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized: user not authenticated",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid user id in token",
		})
	}

	ctx := c.Context()
	rows, err := h.userService.GetUserWithFileId(ctx, userID)
	if err != nil {
		switch err.Error() {
		case "user not found":
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not found",
			})
		default:
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Server error",
			})
		}
	}
	return c.Status(fiber.StatusOK).JSON(rows)
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	ctx := c.Context()

	// TODO: Use authz package
	userIDStr, ok := internal.GetUserIDFromFiber(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized: user not authenticated",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid user id in token",
		})
	}

	req := model.UserRequest{}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if req.FileID != nil {
		if err := uuid.Validate(*req.FileID); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Validation error",
				"details": "fileId is not valid",
			})
		}
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		var details []string
		for _, ve := range err.(validator.ValidationErrors) {
			fieldName := getJSONTagName(ve.StructNamespace())
			switch ve.Tag() {
			case "required":
				details = append(details, fieldName+" is required")
			case "min":
				details = append(details, fieldName+" must be at least "+ve.Param())
			case "max":
				details = append(details, fieldName+" must be at most "+ve.Param())
			default:
				details = append(details, fieldName+" is not valid")
			}
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation error",
			"details": details,
		})
	}

	rows, err := h.userService.UpdateUser(ctx, userID, req)
	if err != nil {
		switch err.Error() {
		case "file not found":
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "fileId is not valid / exists",
			})
		case "unauthorized: you don't own this file":
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"error": "you don't own this file",
			})
		default:
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Server error",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(rows)
}
