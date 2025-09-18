// services/purchase.go
package services

import (
	"context"
	"errors"
	"fmt"

	// "github.com/teammachinist/tutuplapak/clients"
	"github.com/google/uuid"
	"github.com/teammachinist/tutuplapak/services/core/clients"
	"github.com/teammachinist/tutuplapak/services/core/models"
	"github.com/teammachinist/tutuplapak/services/core/repositories"
)

type PurchaseServiceInterface interface {
	CreatePurchase(ctx context.Context, req models.PurchaseRequest) (models.PurchaseResponse, error)
	UploadPaymentProof(ctx context.Context, purchaseId string, req []string) error
}

type PurchaseService struct {
	purchaseRepo repositories.PurchaseRepositoryInterface
	productRepo  repositories.ProductRepositoryInterface
	fileClient   clients.FileClientInterface
}

func (s *PurchaseService) CreatePurchase(ctx context.Context, req models.PurchaseRequest) (models.PurchaseResponse, error) {

	// ✅ Panggil repo — transaksi di-handle di dalamnya
	resp, err := s.purchaseRepo.CreatePurchase(ctx, req)
	if err != nil {
		return models.PurchaseResponse{}, err
	}

	for i, item := range resp.PurchasedItems {
		if item.FileID != uuid.Nil && s.fileClient != nil {
			fileMeta, err := s.fileClient.GetFileByID(ctx, item.FileID)
			if err == nil {
				resp.PurchasedItems[i].FileURI = fileMeta.FileURI
				resp.PurchasedItems[i].FileThumbnailURI = fileMeta.FileThumbnailURI
			}
			// Jika error, biarkan kosong
		}
	}

	return resp, nil
}

// UploadPaymentProof implements PurchaseServiceInterface.
func (s *PurchaseService) UploadPaymentProof(ctx context.Context, purchaseId string, req []string) error {

	// Ambil purchase by ID
	purchase, err := s.purchaseRepo.GetPurchaseByid(ctx, purchaseId)
	if err != nil {
		return fmt.Errorf("failed to get purchase: %w", err)
	}
	if purchase.PurchaseID == uuid.Nil {
		return errors.New("purchase not found")
	}

	// Validasi status unpaid
	if purchase.Status != "unpaid" {
		return errors.New("purchase is not in unpaid status")
	}

	// Validasi jumlah file IDs == jumlah payment details
	if len(req) != len(purchase.PaymentDetails) {
		return fmt.Errorf("expected %d payment proof files, got %d", len(purchase.PaymentDetails), len(req))
	}

	// Validasi file IDs
	_, err = s.fileClient.GetFilesByIDList(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to validate file IDs: %w", err)
	}

	// Kurangi stok produk
	for _, item := range purchase.PurchasedItems {
		err = s.productRepo.UpdateProductQty(ctx, item.ProductID.String(), item.Qty)
		if err != nil {
			return fmt.Errorf("failed to reduce stock for product %s: %w", item.ProductID, err)
		}
	}

	// Update status jadi paid
	err = s.purchaseRepo.UpdatePurchaseStatus(ctx, purchaseId, "paid")
	if err != nil {
		return fmt.Errorf("failed to update purchase status: %w", err)
	}

	return nil
}

func NewPurchaseService(
	purchaseRepo repositories.PurchaseRepositoryInterface,
	productRepo repositories.ProductRepositoryInterface,
	fileClient clients.FileClientInterface,

) PurchaseServiceInterface {
	return &PurchaseService{
		purchaseRepo: purchaseRepo,
		productRepo:  productRepo,
		fileClient:   fileClient,
	}
}
