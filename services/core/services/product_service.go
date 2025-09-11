package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"tutuplapak-core/internal/db"
	"tutuplapak-core/models"
	"tutuplapak-core/repositories"

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

func (s *ProductService) CreateProduct(ctx context.Context, req models.ProductRequest) (models.ProductResponse, error) {
	// user, err := s.userRepo.GetUserExist(ctx, req.UserID)
	// if err != nil {
	// 	return models.ProductResponse{}, errors.New("user not found or invalid token")
	// }

	// file, err := s.fileClient.GetFileByID(ctx, req.FileID)
	// if err != nil {
	// 	return models.ProductResponse{}, errors.New("file not found")
	// }

	_, err := s.productRepo.CheckSKUExistsByUser(ctx, req.SKU, req.UserID)
	if err == nil {
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

func (s *ProductService) UpdateProduct(
	ctx context.Context,
	productID uuid.UUID,
	req models.ProductRequest,
	userID uuid.UUID,
) (models.ProductResponse, error) {

	// owned, err := s.productRepo.CheckProductOwnership(ctx, productID, userID)
	// if err != nil {
	// 	return models.ProductResponse{}, err
	// }
	// if !owned {
	// 	return models.ProductResponse{}, errors.New("unauthorized: you don't own this product")
	// }

	// ??? is this neccesaryy?
	existingProductID, err := s.productRepo.CheckSKUExistsByUser(ctx, req.SKU, userID)
	if err == nil {
		if existingProductID != productID {
			return models.ProductResponse{}, errors.New("sku already exists")
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return models.ProductResponse{}, err
	}

	// // 3. Validasi fileId jika tidak zero
	// if req.FileID != uuid.Nil {
	// 	_, err := s.fileClient.GetFileByID(ctx, req.FileID)
	// 	if err != nil {
	// 		return models.ProductResponse{}, errors.New("file not found or invalid")
	// 	}
	// }

	updatedRow, err := s.productRepo.UpdateProduct(ctx, db.UpdateProductParams{
		ID:        productID,
		Name:      req.Name,
		Category:  req.Category,
		Qty:       int32(req.Qty),
		Price:     int32(req.Price),
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

	// // 6. Ambil metadata file jika ada
	// if updatedRow.FileID != uuid.Nil {
	// 	file, err := s.fileClient.GetFileByID(ctx, updatedRow.FileID)
	// 	if err == nil {
	// 		resp.FileURI = file.URI
	// 		resp.FileThumbnailURI = file.ThumbnailURI
	// 	}
	// 	// Jika error, biarkan string kosong — tidak return error
	// }

	return resp, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, productID uuid.UUID, userID uuid.UUID) error {
	// owned, err := s.productRepo.CheckProductOwnership(ctx, productID, userID)
	// if err != nil {
	// 	return err
	// }
	// if !owned {
	// 	return errors.New("unauthorized: you don't own this product")
	// }

	err := s.productRepo.DeleteProduct(ctx, productID, userID)
	if err != nil {
		return err
	}

	return nil
}
