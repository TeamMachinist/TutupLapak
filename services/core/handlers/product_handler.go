package handlers

import (
	"net/http"
	"reflect"
	"strings"

	"tutuplapak-core/models"
	"tutuplapak-core/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ProductHandler struct {
	productService services.ProductServiceInterface
}

func NewProductHandler(productService services.ProductServiceInterface) *ProductHandler {
	return &ProductHandler{productService: productService}
}

func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var req models.CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON payload",
		})
	}

	validate := validator.New()

	validate.RegisterValidation("category_enum", func(fl validator.FieldLevel) bool {
		category := fl.Field().String()
		allowed := map[string]bool{
			"Food":      true,
			"Beverage":  true,
			"Clothes":   true,
			"Furniture": true,
			"Tools":     true,
		}
		return allowed[category]
	})

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
			case "category_enum":
				details = append(details, fieldName+" must be one of: Food, Beverage, Clothes, Furniture, Tools")
			default:
				details = append(details, fieldName+" is not valid")
			}
		}
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation error",
			"details": details,
		})
	}

	// userIDStr := c.Locals("userID").(string)
	// userID, err := uuid.Parse(userIDStr)
	// if err != nil {
	// 	return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
	// 		"error": "Invalid or missing token",
	// 	})
	// }
	// req.UserID = userID

	req.UserID = uuid.MustParse("11111111-1111-1111-1111-111111111111") // UUID dummy valid

	productResp, err := h.productService.CreateProduct(c.Context(), req)
	if err != nil {
		switch err.Error() {
		case "sku already exists":
			return c.Status(http.StatusConflict).JSON(fiber.Map{
				"error": "sku already exists",
			})
		case "user not found or invalid token":
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		case "file not found":
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "fileId is not valid / exists",
			})
		default:
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Server error",
			})
		}
	}
	return c.Status(http.StatusCreated).JSON(productResp)
}

func getJSONTagName(fieldPath string) string {
	parts := strings.Split(fieldPath, ".")
	if len(parts) == 0 {
		return fieldPath
	}
	fieldName := parts[len(parts)-1]

	t := reflect.TypeOf(models.CreateProductRequest{})
	if f, ok := t.FieldByName(fieldName); ok {
		jsonTag := f.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			if idx := strings.Index(jsonTag, ","); idx != -1 {
				jsonTag = jsonTag[:idx]
			}
			return jsonTag
		}
	}
	return fieldName
}
