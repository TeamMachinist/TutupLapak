package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/teammachinist/tutuplapak/services/core/internal/clients"
	"github.com/teammachinist/tutuplapak/services/core/internal/database"
	"github.com/teammachinist/tutuplapak/services/core/internal/model"
	"github.com/teammachinist/tutuplapak/services/core/internal/repository"

	"github.com/google/uuid"
)

type PurchaseServiceInterface interface {
	CreatePurchase(ctx context.Context, req model.PurchaseRequest) (model.PurchaseResponse, error)
	UploadPaymentProof(ctx context.Context, purchaseId string, req []string) error
}

type PurchaseService struct {
	purchaseRepo repository.PurchaseRepositoryInterface
	productRepo  repository.ProductRepositoryInterface
	fileClient   clients.FileClientInterface
}

func (s *PurchaseService) CreatePurchase(ctx context.Context, req model.PurchaseRequest) (model.PurchaseResponse, error) {

	resp, err := s.purchaseRepo.CreatePurchase(ctx, req)
	if err != nil {
		return model.PurchaseResponse{}, err
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
	if purchase.Status != model.PurchaseStatus(database.PurchaseStatusUnpaid) {
		return errors.New("purchase is already paid")
	}

	// Validasi jumlah file IDs == jumlah payment details
	if len(req) != len(purchase.PaymentDetails) {
		return fmt.Errorf("expected %d payment proof files, got %d", len(purchase.PaymentDetails), len(req))
	}

	// Validasi file IDs
	_, err = s.fileClient.GetFilesByIDList(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "invalid") {
			return fmt.Errorf("invalid or non-existent file IDs")
		}
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
	err = s.purchaseRepo.UpdatePurchaseStatus(ctx, purchaseId, database.PurchaseStatusPaid)
	if err != nil {
		return fmt.Errorf("failed to update purchase status: %w", err)
	}

	return nil
}

func NewPurchaseService(
	purchaseRepo repository.PurchaseRepositoryInterface,
	productRepo repository.ProductRepositoryInterface,
	fileClient clients.FileClientInterface,

) PurchaseServiceInterface {
	return &PurchaseService{
		purchaseRepo: purchaseRepo,
		productRepo:  productRepo,
		fileClient:   fileClient,
	}
}
