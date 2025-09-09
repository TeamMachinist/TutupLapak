package services

import (
	"context"
	"errors"

	"tutuplapak-core/models"
	"tutuplapak-core/repositories"
)

type ProductServiceInterface interface {
	CreateProduct(ctx context.Context, req models.CreateProductRequest) (models.ProductResponse, error)
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

func (s *ProductService) ListProducts(ctx context.Context, req models.ListProductsRequest) ([]models.ProductResponse, error) {
	products, err := s.productRepo.ListProducts(ctx, req)
	if err != nil {
		return nil, err
	}
	return products, nil
}
