// handlers/purchase.go
package handler

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/teammachinist/tutuplapak/services/core/internal/logger"
	"github.com/teammachinist/tutuplapak/services/core/internal/model"
	"github.com/teammachinist/tutuplapak/services/core/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PurchaseHandler struct {
	purchaseService service.PurchaseServiceInterface
}

func NewPurchaseHandler(purchaseService service.PurchaseServiceInterface) *PurchaseHandler {
	return &PurchaseHandler{purchaseService: purchaseService}
}

func (h *PurchaseHandler) CreatePurchase(c *fiber.Ctx) error {
	ctx := c.Context()

	var req model.PurchaseRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	// Validasi dasar
	if len(req.PurchasedItems) == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "purchasedItems is required"})
	}

	for _, item := range req.PurchasedItems {
		if item.ProductID == uuid.Nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "productId is required"})
		}
		if item.Qty < 1 { // ← min: 2, bukan 1
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "qty must be at least 2"})
		}
	}

	if len(req.SenderName) < 4 || len(req.SenderName) > 55 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "senderName must be 4-55 characters"})
	}

	if req.SenderContactType != "email" && req.SenderContactType != "phone" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "senderContactType must be 'email' or 'phone'"})
	}

	if req.SenderContactDetail == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "senderContactDetail is required"})
	}

	// Validasi contact detail
	switch req.SenderContactType {
	case "email":
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(req.SenderContactDetail) {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid email format"})
		}
	case "phone":
		phoneRegex := regexp.MustCompile(`^\+?[0-9\s\-\(\)]{7,15}$`)
		if !phoneRegex.MatchString(req.SenderContactDetail) {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid phone number format"})
		}
	}

	resp, err := h.purchaseService.CreatePurchase(ctx, req)
	if err != nil {
		switch {
		case err.Error() == "product not found":
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid product ID"})
		case strings.HasPrefix(err.Error(), "insufficient stock for"):
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "quantity exceeds available stock"})
		default:
			logger.WarnCtx(ctx, "Failed to create purchase", "error", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
		}
	}

	return c.Status(http.StatusCreated).JSON(resp)
}

func (h *PurchaseHandler) UploadPaymentProof(c *fiber.Ctx) error {
	ctx := c.Context()
	logger.InfoCtx(ctx, "Upload payment proof request")

	purchaseId := c.Params("purchaseId")

	var body struct {
		FileIds []string `json:"fileIds" validate:"required,min=1,dive"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if len(body.FileIds) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "fileIds must not be empty",
		})
	}

	err := h.purchaseService.UploadPaymentProof(c.Context(), purchaseId, body.FileIds)
	if err != nil {
		// Handle error secara spesifik

		if strings.Contains(err.Error(), "purchase not found") {
			return c.SendStatus(fiber.StatusNotFound)
		}

		if strings.Contains(err.Error(), "purchase is already paid") ||
			strings.Contains(err.Error(), "expected") ||
			strings.Contains(err.Error(), "invalid or non-existent file IDs") ||
			strings.Contains(err.Error(), "invalid request body") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		logger.ErrorCtx(ctx, "Failed to upload payment proof", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "payment proof processed successfully, purchase marked as paid and stock reduced",
	})
}
