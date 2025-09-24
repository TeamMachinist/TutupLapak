package handler

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/teammachinist/tutuplapak/services/auth/pkg/authz"
	"github.com/teammachinist/tutuplapak/services/core/internal/logger"
	"github.com/teammachinist/tutuplapak/services/core/internal/model"
	"github.com/teammachinist/tutuplapak/services/core/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ProductHandler struct {
	productService service.ProductServiceInterface
}

func NewProductHandler(productService service.ProductServiceInterface) *ProductHandler {
	return &ProductHandler{productService: productService}
}

func (h *ProductHandler) GetAllProducts(c *fiber.Ctx) error {
	ctx := c.Context()

	limit := 5
	offset := 0

	var productID *uuid.UUID
	var sku *string
	var category *string
	var sortBy *string

	if limStr := c.Query("limit"); limStr != "" {
		if l, err := strconv.Atoi(limStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offStr := c.Query("offset"); offStr != "" {
		if o, err := strconv.Atoi(offStr); err == nil && o >= 0 {
			offset = o
		}
	}

	if pidStr := c.Query("productId"); pidStr != "" {
		if pid, err := uuid.Parse(pidStr); err == nil {
			productID = &pid
		}
	}

	if s := c.Query("sku"); s != "" {
		sku = &s
	}

	if cat := c.Query("category"); cat != "" {
		allowedCategories := map[string]bool{
			"Food":      true,
			"Beverage":  true,
			"Clothes":   true,
			"Furniture": true,
			"Tools":     true,
		}
		if allowedCategories[cat] {
			category = &cat
		}
	}

	if sb := c.Query("sortBy"); sb != "" {
		validSorts := map[string]bool{
			"newest":    true,
			"oldest":    true,
			"cheapest":  true,
			"expensive": true,
		}
		if validSorts[strings.ToLower(sb)] {
			lower := strings.ToLower(sb)
			sortBy = &lower
		}
	}

	filter := model.GetAllProductsParams{
		Limit:     limit,
		Offset:    offset,
		ProductID: productID,
		SKU:       sku,
		Category:  category,
		SortBy:    sortBy,
	}

	products, err := h.productService.GetAllProducts(ctx, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch products",
		})
	}

	if len(products) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "no products found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(products)
}

func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	ctx := c.Context()

	var req model.ProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
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
			fieldName := getJSONTagName(ve.StructNamespace(), reflect.TypeOf(req))
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": details,
		})
	}

	// Get user ID from authz
	userIDStr, ok := authz.GetUserIDFromFiber(c)
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

	req.UserID = userID

	productResp, err := h.productService.CreateProduct(c.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows): // This should not happen here
			// Ignore; handled above
		case strings.Contains(err.Error(), "sku already exists"):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "sku already exists",
			})
		case strings.Contains(err.Error(), "file not found"):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "fileId is not valid or does not exist",
			})
		case strings.Contains(err.Error(), "user not found or invalid token"):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		default:
			logger.WarnCtx(ctx, "Failed to create product", "error", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "internal server error",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(productResp)
}

func getJSONTagName(fieldPath string, structType reflect.Type) string {
	parts := strings.Split(fieldPath, ".")
	if len(parts) == 0 {
		return fieldPath
	}
	fieldName := parts[len(parts)-1]

	if f, ok := structType.FieldByName(fieldName); ok {
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

// UpdateProduct menangani PUT /products/:productId
func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	ctx := c.Context()

	productIDStr := c.Params("productId")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "invalid product id format",
		})
	}

	// Get user ID from authz
	userIDStr, ok := authz.GetUserIDFromFiber(c)
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

	var req model.ProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	// userID := uuid.MustParse("11111111-1111-1111-1111-111111111111") // UUID dummy valid

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
			fieldName := getJSONTagName(ve.StructNamespace(), reflect.TypeOf(req))
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

	resp, err := h.productService.UpdateProduct(ctx, productID, req, userID)
	if err != nil {
		// Handle error langsung dengan errors.Is atau string match
		switch {
		case errors.Is(err, context.Canceled):
			return c.Status(fiber.StatusRequestTimeout).JSON(fiber.Map{"error": "request timeout"})
		case errors.Is(err, context.DeadlineExceeded):
			return c.Status(fiber.StatusRequestTimeout).JSON(fiber.Map{"error": "request timeout"})
		case err.Error() == "unauthorized: you don't own this product":
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		case err.Error() == "sku already exists for your account":
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		case err.Error() == "fileId is not valid / exists":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case err.Error() == "product not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case err.Error() == "product doesn't exist":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		default:
			c.App().Config().ErrorHandler(c, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "internal server error",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	ctx := c.Context()

	productIDStr := c.Params("productId")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "invalid product id format",
		})
	}

	// Get user ID from authz
	userIDStr, ok := authz.GetUserIDFromFiber(c)
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
	// userID := uuid.MustParse("11111111-1111-1111-1111-111111111111") // UUID dummy valid

	err = h.productService.DeleteProduct(ctx, productID, userID)
	if err != nil {
		switch {
		case err.Error() == "unauthorized: you don't own this product":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		case strings.Contains(err.Error(), "no rows in result set"):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "product not found",
			})
		case err.Error() == "product doesn't exist":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "internal server error",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(nil)
}
