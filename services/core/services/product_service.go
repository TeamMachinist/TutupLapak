package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/teammachinist/tutuplapak/internal/database"
	"github.com/teammachinist/tutuplapak/services/core/clients"
	"github.com/teammachinist/tutuplapak/services/core/models"
	"github.com/teammachinist/tutuplapak/services/core/repositories"

	"github.com/google/uuid"
)

type ProductServiceInterface interface {
	CreateProduct(ctx context.Context, req models.ProductRequest) (models.ProductResponse, error)
	GetAllProducts(ctx context.Context, filter models.GetAllProductsParams) ([]models.ProductResponse, error)
	UpdateProduct(
		ctx context.Context,
		productID uuid.UUID,
		req models.ProductRequest,
		userID uuid.UUID,
	) (models.ProductResponse, error)
	DeleteProduct(ctx context.Context, productID uuid.UUID, userID uuid.UUID) error
}

type ProductService struct {
	productRepo repositories.ProductRepositoryInterface
	// userRepo	repositories.UserRepositoryInterface
	// authClient  clients.AuthClientInterface
	fileClient clients.FileClientInterface
}

func NewProductService(
	productRepo repositories.ProductRepositoryInterface,
	// userRepo repositories.UserRepositoryInterface,
	// authClient clients.AuthClientInterface,
	fileClient clients.FileClientInterface,
) ProductServiceInterface {
	return &ProductService{
		productRepo: productRepo,
		// userRepo:    userRepo,
		// authClient:  authClient,
		fileClient: fileClient,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req models.ProductRequest) (models.ProductResponse, error) {

	_, err := s.productRepo.CheckSKUExistsByUser(ctx, req.SKU, req.UserID)
	if err == nil {
		return models.ProductResponse{}, errors.New("sku already exists")
	}

	productResp, err := s.productRepo.CreateProduct(ctx, req)
	if err != nil {
		return models.ProductResponse{}, err
	}

	if req.FileID != uuid.Nil {
		file, err := s.fileClient.GetFileByID(ctx, req.FileID, req.UserID.String())
		if err != nil {
			return models.ProductResponse{}, errors.New("fileId is not valid / exists")
		}

		if file.UserID != req.UserID.String() {
			return models.ProductResponse{}, errors.New("fileId is not valid / exists")
		}

		productResp.FileURI = file.FileURI
		productResp.FileThumbnailURI = file.FileThumbnailURI
	}

	return productResp, nil
}

func (s *ProductService) GetAllProducts(ctx context.Context, filter models.GetAllProductsParams) ([]models.ProductResponse, error) {
	products, err := s.productRepo.GetAllProducts(ctx, filter)
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

		if p.FileID != uuid.Nil {
			file, err := s.fileClient.GetFileByID(ctx, p.FileID, p.UserID.String()) // ← p.UserID!
			if err == nil {
				resp.FileURI = file.FileURI
				resp.FileThumbnailURI = file.FileThumbnailURI
			}
		}

		responses = append(responses, resp)
	}

	return responses, nil
}

func (s *ProductService) UpdateProduct(
	ctx context.Context,
	productID uuid.UUID,
	req models.ProductRequest,
	userID uuid.UUID,
) (models.ProductResponse, error) {

	owned, err := s.productRepo.CheckProductOwnership(ctx, productID, userID)
	if err != nil {
		// fmt.Printf("Error checking ownership: %v\n", err)
		return models.ProductResponse{}, fmt.Errorf("internal error verifying ownership")
	}

	if !owned {
		fmt.Println("User does not own this product")
		return models.ProductResponse{}, errors.New("unauthorized: you don't own this product")
	}
	fmt.Println("error disini")
	existingProduct, err := s.productRepo.CheckSKUExistsByUser(ctx, req.SKU, userID)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return models.ProductResponse{}, err
		}
	} else {
		if existingProduct.ID != productID {
			return models.ProductResponse{}, errors.New("sku already exists for this user")
		}
	}
	var fileMetadata *clients.FileMetadataResponse

	// Validasi & ambil metadata file jika user ingin ubah file
	if req.FileID != uuid.Nil {
		file, err := s.fileClient.GetFileByID(ctx, req.FileID, userID.String())
		if err != nil {
			return models.ProductResponse{}, errors.New("fileId is not valid / exists")
		}
		if file.UserID != userID.String() {
			return models.ProductResponse{}, errors.New("fileId is not valid / exists")
		}

		fileMetadata = file
	}

	updatedRow, err := s.productRepo.UpdateProduct(ctx, database.UpdateProductParams{
		ID:        productID,
		Name:      req.Name,
		Category:  req.Category,
		Qty:       req.Qty,
		Price:     req.Price,
		Sku:       req.SKU,
		FileID:    req.FileID,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return models.ProductResponse{}, err
	}

	// 5. Mapping ke ProductResponse (by value)
	resp := models.ProductResponse{
		ProductID:        updatedRow.ID,
		Name:             updatedRow.Name,
		Category:         updatedRow.Category,
		Qty:              int(updatedRow.Qty),
		Price:            int(updatedRow.Price),
		SKU:              updatedRow.Sku,
		FileID:           updatedRow.FileID,
		CreatedAt:        updatedRow.CreatedAt,
		UpdatedAt:        updatedRow.UpdatedAt,
		FileURI:          "",
		FileThumbnailURI: "",
	}

	// Gunakan metadata yang sudah diambil, atau ambil baru jika diperlukan
	if fileMetadata != nil {
		resp.FileURI = fileMetadata.FileURI
		resp.FileThumbnailURI = fileMetadata.FileThumbnailURI
	} else if updatedRow.FileID != uuid.Nil {
		// Ambil metadata file lama (jika tidak diubah)
		file, err := s.fileClient.GetFileByID(ctx, updatedRow.FileID, userID.String())
		if err == nil {

			resp.FileURI = file.FileURI
			resp.FileThumbnailURI = file.FileThumbnailURI
		}
	}

	return resp, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, productID uuid.UUID, userID uuid.UUID) error {
	owned, err := s.productRepo.CheckProductOwnership(ctx, productID, userID)
	if err != nil {
		// fmt.Printf("Error checking ownership: %v\n", err)
		return fmt.Errorf("internal error verifying ownership")
	}

	if !owned {
		fmt.Println("User does not own this product")
		return errors.New("unauthorized: you don't own this product")
	}

	err = s.productRepo.DeleteProduct(ctx, productID, userID)
	if err != nil {
		return err
	}

	return nil
}
