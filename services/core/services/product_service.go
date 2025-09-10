package services

import (
	"context"
	"errors"

	"tutuplapak-core/models"
	"tutuplapak-core/repositories"

	"github.com/google/uuid"
)

type ProductServiceInterface interface {
	CreateProduct(ctx context.Context, req models.CreateProductRequest) (models.ProductResponse, error)
	GetAllProducts(ctx context.Context, filter GetAllProductsFilter) ([]models.ProductResponse, error)
}

type ProductService struct {
	productRepo repositories.ProductRepositoryInterface
	// userRepo	repositories.UserRepositoryInterface
	// authClient  clients.AuthClientInterface
	// fileClient  clients.FileClientInterface
}

func NewProductService(
	productRepo repositories.ProductRepositoryInterface,
	// userRepo repositories.UserRepositoryInterface,
	// authClient clients.AuthClientInterface,
	// fileClient clients.FileClientInterface,
) ProductServiceInterface {
	return &ProductService{
		productRepo: productRepo,
		// userRepo:    userRepo,
		// authClient:  authClient,
		// fileClient:  fileClient,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req models.CreateProductRequest) (models.ProductResponse, error) {
	// user, err := s.userRepo.GetUserExist(ctx, req.UserID)
	// if err != nil {
	// 	return models.ProductResponse{}, errors.New("user not found or invalid token")
	// }

	// file, err := s.fileClient.GetFileByID(ctx, req.FileID)
	// if err != nil {
	// 	return models.ProductResponse{}, errors.New("file not found")
	// }

	exists, err := s.productRepo.CheckSKUExistsByUser(ctx, req.SKU, req.UserID)
	if err != nil {
		return models.ProductResponse{}, err
	}
	if exists {
		return models.ProductResponse{}, errors.New("sku already exists")
	}

	productResp, err := s.productRepo.CreateProduct(ctx, req)
	if err != nil {
		return models.ProductResponse{}, err
	}

	// productResp.FileURI = file.URI
	// productResp.FileThumbnailURI = file.ThumbnailURI

	return productResp, nil
}

// Filter struct — tanpa pointer untuk limit/offset (selalu ada)
type GetAllProductsFilter struct {
	Limit     int
	Offset    int
	ProductID *uuid.UUID
	SKU       *string
	Category  *string
	SortBy    *string
}

func (s *ProductService) GetAllProducts(ctx context.Context, filter GetAllProductsFilter) ([]models.ProductResponse, error) {
	// Panggil repo
	products, err := s.productRepo.GetAllProducts(ctx, repositories.GetAllProductsParams{
		Limit:     filter.Limit,
		Offset:    filter.Offset,
		ProductID: filter.ProductID,
		SKU:       filter.SKU,
		Category:  filter.Category,
		SortBy:    filter.SortBy,
	})
	if err != nil {
		return nil, err
	}

	var responses []models.ProductResponse

	for _, p := range products {
		resp := models.ProductResponse{
			ProductID: p.ID,
			Name:      p.Name,
			Category:  p.Category,
			Qty:       p.Qty,
			Price:     p.Price,
			SKU:       p.SKU,
			FileID:    p.FileID,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
			// FileURI & FileThumbnailURI akan diisi di bawah
		}

		// // Ambil file metadata jika FileID valid
		// if p.FileID != uuid.Nil {
		// 	file, err := s.fileClient.GetFileByID(ctx, p.FileID)
		// 	if err == nil {
		// 		resp.FileURI = file.URI
		// 		resp.FileThumbnailURI = file.ThumbnailURI
		// 	}
		// 	// Jika error, biarkan kosong — sesuai permintaan "ignore if invalid"
		// }

		responses = append(responses, resp)
	}

	return responses, nil
}
