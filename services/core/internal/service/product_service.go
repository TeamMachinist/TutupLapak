package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/teammachinist/tutuplapak/services/core/internal/cache"
	"github.com/teammachinist/tutuplapak/services/core/internal/clients"
	"github.com/teammachinist/tutuplapak/services/core/internal/database"
	"github.com/teammachinist/tutuplapak/services/core/internal/model"
	"github.com/teammachinist/tutuplapak/services/core/internal/repository"

	"github.com/google/uuid"
)

type ProductServiceInterface interface {
	CreateProduct(ctx context.Context, req model.ProductRequest) (model.ProductResponse, error)
	GetAllProducts(ctx context.Context, filter model.GetAllProductsParams) ([]model.ProductResponse, error)
	UpdateProduct(
		ctx context.Context,
		productID uuid.UUID,
		req model.ProductRequest,
		userID uuid.UUID,
	) (model.ProductResponse, error)
	DeleteProduct(ctx context.Context, productID uuid.UUID, userID uuid.UUID) error
}

type ProductService struct {
	productRepo repository.ProductRepositoryInterface
	// userRepo	repository.UserRepositoryInterface
	// authClient  clients.AuthClientInterface
	fileClient clients.FileClientInterface
	cache      *cache.RedisCache
}

func NewProductService(
	productRepo repository.ProductRepositoryInterface,
	// userRepo repository.UserRepositoryInterface,
	// authClient clients.AuthClientInterface,
	fileClient clients.FileClientInterface,
	cache *cache.RedisCache,
) ProductServiceInterface {
	return &ProductService{
		productRepo: productRepo,
		// userRepo:    userRepo,
		// authClient:  authClient,
		fileClient: fileClient,
		cache:      cache,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req model.ProductRequest) (model.ProductResponse, error) {
	// Check if SKU already exists for the user
	_, err := s.productRepo.CheckSKUExistsByUser(ctx, req.SKU, req.UserID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return model.ProductResponse{}, fmt.Errorf("failed to check SKU existence: %w", err)
		}
	} else {
		return model.ProductResponse{}, errors.New("sku already exists")
	}

	var parsedFileId uuid.UUID

	fileID, err := uuid.Parse(req.FileID)
	if err != nil {
		return model.ProductResponse{}, err
	}

	parsedFileId = fileID

	var file *clients.FileMetadataResponse
	if req.FileID != "" {
		f, err := s.fileClient.GetFileByID(ctx, parsedFileId)
		if err != nil {
			return model.ProductResponse{}, errors.New("file not found") // tetap pakai string ini untuk handler
		}
		file = f
	}

	// Create product
	productResp, err := s.productRepo.CreateProduct(ctx, req)
	if err != nil {
		return model.ProductResponse{}, fmt.Errorf("failed to create product: %w", err)
	}

	// Attach file info if present
	if file != nil {
		productResp.FileURI = file.FileURI
		productResp.FileThumbnailURI = file.FileThumbnailURI
	}

	log.Printf("[CreateProduct] Product created successfully with ID: %s", productResp.ProductID)

	return productResp, nil
}

func (s *ProductService) GetAllProducts(ctx context.Context, filter model.GetAllProductsParams) ([]model.ProductResponse, error) {
	key := cache.ProductListKey + generateFilterHash(filter)

	var products []model.ProductResponse

	err := s.cache.Get(ctx, key, &products)
	if err == nil {
		log.Printf("[GetAllProducts] Cache HIT for key: %s", key)
		return products, nil
	}

	log.Printf("[GetAllProducts] Cache MISS for key: %s", key)
	productsDB, err := s.productRepo.GetAllProducts(ctx, filter)
	if err != nil {
		return nil, err
	}

	var responses []model.ProductResponse
	for _, p := range productsDB {
		resp := model.ProductResponse{
			ProductID: p.ID,
			Name:      p.Name,
			Category:  p.Category,
			Qty:       p.Qty,
			Price:     p.Price,
			SKU:       p.SKU,
			FileID:    p.FileID,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		}

		if p.FileID != uuid.Nil {
			file, err := s.fileClient.GetFileByID(ctx, p.FileID)
			if err == nil {
				resp.FileURI = file.FileURI
				resp.FileThumbnailURI = file.FileThumbnailURI
			} else {
				log.Printf("[GetAllProducts] Failed to fetch file %s: %v", p.FileID, err)
			}
		}

		responses = append(responses, resp)
	}

	err = s.cache.Set(ctx, key, responses, cache.ProductListTTL)
	if err != nil {
		log.Printf("[GetAllProducts] Failed to set cache: %v", err)
	}

	return responses, nil
}

func (s *ProductService) UpdateProduct(
	ctx context.Context,
	productID uuid.UUID,
	req model.ProductRequest,
	userID uuid.UUID,
) (model.ProductResponse, error) {

	owned, err := s.productRepo.CheckProductOwnership(ctx, productID, userID)
	if err != nil {
		// fmt.Printf("Error checking ownership: %v\n", err)
		return model.ProductResponse{}, fmt.Errorf("internal error verifying ownership")
	}

	if !owned {
		fmt.Println("User does not own this product")
		return model.ProductResponse{}, errors.New("unauthorized: you don't own this product")
	}
	fmt.Println("error disini")
	existingProduct, err := s.productRepo.CheckSKUExistsByUser(ctx, req.SKU, userID)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return model.ProductResponse{}, err
		}
	} else {
		if existingProduct.ID != productID {
			return model.ProductResponse{}, errors.New("sku already exists for this user")
		}
	}
	var fileMetadata *clients.FileMetadataResponse

	var parsedFileId uuid.UUID

	fileID, err := uuid.Parse(req.FileID)
	if err != nil {
		return model.ProductResponse{}, err
	}

	parsedFileId = fileID

	if req.FileID != "" {
		file, err := s.fileClient.GetFileByID(ctx, parsedFileId)
		if err != nil {
			return model.ProductResponse{}, errors.New("fileId is not valid / exists")
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
		FileID:    parsedFileId,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return model.ProductResponse{}, err
	}

	resp := model.ProductResponse{
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

	if fileMetadata != nil {
		resp.FileURI = fileMetadata.FileURI
		resp.FileThumbnailURI = fileMetadata.FileThumbnailURI
	} else if updatedRow.FileID != uuid.Nil {
		file, err := s.fileClient.GetFileByID(ctx, updatedRow.FileID)
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

func generateFilterHash(filter model.GetAllProductsParams) string {
	// Konversi semua field filter ke string, urutkan agar konsisten
	var parts []string

	if filter.Limit > 0 {
		parts = append(parts, fmt.Sprintf("limit=%d", filter.Limit))
	}
	if filter.Offset > 0 {
		parts = append(parts, fmt.Sprintf("offset=%d", filter.Offset))
	}
	if filter.ProductID != nil {
		parts = append(parts, fmt.Sprintf("productId=%s", filter.ProductID.String()))
	}
	if filter.SKU != nil {
		parts = append(parts, fmt.Sprintf("sku=%s", *filter.SKU))
	}
	if filter.Category != nil {
		parts = append(parts, fmt.Sprintf("category=%s", *filter.Category))
	}
	if filter.SortBy != nil {
		parts = append(parts, fmt.Sprintf("sortBy=%s", *filter.SortBy))
	}

	// Urutkan agar key konsisten meski parameter beda urutan
	sort.Strings(parts)

	// Gabungkan dan hash
	input := strings.Join(parts, "&")
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash[:])
}
