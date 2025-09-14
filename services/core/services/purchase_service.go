// services/purchase.go
package services

import (
	"context"

	// "github.com/teammachinist/tutuplapak/clients"
	"github.com/google/uuid"
	"github.com/teammachinist/tutuplapak/services/core/clients"
	"github.com/teammachinist/tutuplapak/services/core/models"
	"github.com/teammachinist/tutuplapak/services/core/repositories"
)

type PurchaseServiceInterface interface {
	CreatePurchase(ctx context.Context, req models.PurchaseRequest) (models.PurchaseResponse, error)
}

type PurchaseService struct {
	purchaseRepo repositories.PurchaseRepositoryInterface
	fileClient   clients.FileClientInterface
}

func NewPurchaseService(
	purchaseRepo repositories.PurchaseRepositoryInterface,
	fileClient clients.FileClientInterface,
) PurchaseServiceInterface {
	return &PurchaseService{
		purchaseRepo: purchaseRepo,
		fileClient:   fileClient,
	}
}

func (s *PurchaseService) CreatePurchase(ctx context.Context, req models.PurchaseRequest) (models.PurchaseResponse, error) {

	// ✅ Panggil repo — transaksi di-handle di dalamnya
	resp, err := s.purchaseRepo.CreatePurchase(ctx, req)
	if err != nil {
		return models.PurchaseResponse{}, err
	}

	for i, item := range resp.PurchasedItems {
		if item.FileID != uuid.Nil && s.fileClient != nil {
			fileMeta, err := s.fileClient.GetFileByID(ctx, item.FileID, item.UserID.String())
			if err == nil {
				resp.PurchasedItems[i].FileURI = fileMeta.FileURI
				resp.PurchasedItems[i].FileThumbnailURI = fileMeta.FileThumbnailURI
			}
			// Jika error, biarkan kosong
		}
	}

	return resp, nil
}
