package services

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

	"github.com/teammachinist/tutuplapak/internal/cache"
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
	cache      *cache.RedisCache
}

func NewProductService(
	productRepo repositories.ProductRepositoryInterface,
	// userRepo repositories.UserRepositoryInterface,
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

func (s *ProductService) CreateProduct(ctx context.Context, req models.ProductRequest) (models.ProductResponse, error) {

	_, err := s.productRepo.CheckSKUExistsByUser(ctx, req.SKU, req.UserID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return models.ProductResponse{}, fmt.Errorf("failed to check SKU existence: %w", err)
		}
	} else {
		return models.ProductResponse{}, errors.New("sku already exists")
	}

	if req.FileID != uuid.Nil {
		file, err := s.fileClient.GetFileByID(ctx, req.FileID)
		if err != nil {
			return models.ProductResponse{}, errors.New("fileId is not valid / exists")
		}

		log.Printf("[CreateProduct] File validated successfully: ID=%s, URI=%s", req.FileID, file.FileURI)
	}

	productResp, err := s.productRepo.CreateProduct(ctx, req)
	if err != nil {
		return models.ProductResponse{}, fmt.Errorf("failed to create product: %w", err)
	}

	log.Printf("[CreateProduct] Product created successfully with ID: %s", productResp.ProductID)

	if req.FileID != uuid.Nil {
		file, err := s.fileClient.GetFileByID(ctx, req.FileID)
		if err != nil {
			log.Printf("[CreateProduct] Failed to re-fetch file after creation: %v", err)
		} else {
			productResp.FileURI = file.FileURI
			productResp.FileThumbnailURI = file.FileThumbnailURI
			log.Printf("[CreateProduct] Attached file info to product response: URI=%s", file.FileURI)
		}
	}

	return productResp, nil
}

func (s *ProductService) GetAllProducts(ctx context.Context, filter models.GetAllProductsParams) ([]models.ProductResponse, error) {
	key := cache.ProductListKey + generateFilterHash(filter)

	var products []models.ProductResponse

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

	var responses []models.ProductResponse
	for _, p := range productsDB {
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
		file, err := s.fileClient.GetFileByID(ctx, req.FileID)
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

func generateFilterHash(filter models.GetAllProductsParams) string {
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
