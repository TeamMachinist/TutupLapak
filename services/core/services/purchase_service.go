// services/purchase.go
package services

import (
	"context"

	// "github.com/teammachinist/tutuplapak/clients"
	"github.com/teammachinist/tutuplapak/services/core/models"
	"github.com/teammachinist/tutuplapak/services/core/repositories"
)

type PurchaseServiceInterface interface {
	CreatePurchase(ctx context.Context, req models.PurchaseRequest) (models.PurchaseResponse, error)
}

type PurchaseService struct {
	purchaseRepo repositories.PurchaseRepositoryInterface
	// fileClient   clients.FileClientInterface // ← Ganti dari ProductService ke FileClient
}

func NewPurchaseService(
	purchaseRepo repositories.PurchaseRepositoryInterface,
	// fileClient clients.FileClientInterface, // ← Inject fileClient
) PurchaseServiceInterface {
	return &PurchaseService{
		purchaseRepo: purchaseRepo,
		// fileClient:   fileClient,
	}
}

func (s *PurchaseService) CreatePurchase(ctx context.Context, req models.PurchaseRequest) (models.PurchaseResponse, error) {

	// ✅ Panggil repo — transaksi di-handle di dalamnya
	resp, err := s.purchaseRepo.CreatePurchase(ctx, req)
	if err != nil {
		return models.PurchaseResponse{}, err
	}

	// // ✅ ISI FILE URI & THUMBNAIL — langsung dari fileClient
	// for i, item := range resp.PurchasedItems {
	// 	if item.FileID != uuid.Nil && s.fileClient != nil {
	// 		fileMeta, err := s.fileClient.GetFileByID(ctx, item.FileID)
	// 		if err == nil {
	// 			resp.PurchasedItems[i].FileURI = fileMeta.URI
	// 			resp.PurchasedItems[i].FileThumbnailURI = fileMeta.ThumbnailURI
	// 		}
	// 		// Jika error, biarkan string kosong
	// 	}
	// }

	return resp, nil
}
