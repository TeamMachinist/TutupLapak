// handlers/purchase.go
package handlers

import (
	"context"
	"net/http"
	"regexp"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/teammachinist/tutuplapak/services/core/models"
	"github.com/teammachinist/tutuplapak/services/core/services"
)

type PurchaseHandler struct {
	purchaseService services.PurchaseServiceInterface
}

func NewPurchaseHandler(purchaseService services.PurchaseServiceInterface) *PurchaseHandler {
	return &PurchaseHandler{purchaseService: purchaseService}
}

func (h *PurchaseHandler) CreatePurchase(c *fiber.Ctx) error {
	var req models.PurchaseRequest

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
		if item.Qty < 1 {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "qty must be at least 1"})
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

	resp, err := h.purchaseService.CreatePurchase(context.Background(), req)
	if err != nil {
		switch {
		case err.Error() == "product not found",
			err.Error()[:22] == "insufficient stock for",
			err.Error() == "failed to update stock":
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
		}
	}

	return c.Status(http.StatusCreated).JSON(resp)
}

func (h *PurchaseHandler) UploadPaymentProof(c *fiber.Ctx) error {
	purchaseId := c.Params("purchaseId") // string

	var body struct {
		FileIds []string `json:"fileIds" validate:"required,min=1,dive,uuid4"`
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "payment proof processed successfully, purchase marked as paid and stock reduced",
	})
}
