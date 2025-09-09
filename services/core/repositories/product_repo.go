package repositories

import (
	"context"
	"time"

	"tutuplapak-core/config"
	"tutuplapak-core/internal/db"
	"tutuplapak-core/models"

	"github.com/google/uuid"
)

type ProductRepositoryInterface interface {
	CreateProduct(ctx context.Context, req models.CreateProductRequest) (models.ProductResponse, error)
	CheckSKUExistsByUser(ctx context.Context, sku string, userID uuid.UUID) (bool, error)
	ListProducts(ctx context.Context, req models.ListProductsRequest) ([]models.ProductResponse, error)
}

type ProductRepository struct {
	db *config.Database
}

func NewProductRepository(database *config.Database) ProductRepositoryInterface {
	return &ProductRepository{db: database}
}

func (r *ProductRepository) CreateProduct(ctx context.Context, req models.CreateProductRequest) (models.ProductResponse, error) {
	productID := uuid.New()

	dbProduct, err := r.db.Queries.CreateProduct(ctx, db.CreateProductParams{
		ID:       productID,
		Name:     req.Name,
		Category: req.Category,
		Qty:      int32(req.Qty),
		Price:    int32(req.Price),
		Sku:      req.SKU,
		FileID:   req.FileID,
		UserID:   req.UserID,
	})
	if err != nil {
		return models.ProductResponse{}, err
	}

	resp := models.ProductResponse{
		ProductID:        dbProduct.ID,
		Name:             dbProduct.Name,
		Category:         dbProduct.Category,
		Qty:              int(dbProduct.Qty),
		Price:            int(dbProduct.Price),
		SKU:              dbProduct.Sku,
		FileID:           dbProduct.FileID,
		FileURI:          "",
		FileThumbnailURI: "",
		CreatedAt:        dbProduct.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:        dbProduct.UpdatedAt.Time.Format(time.RFC3339),
	}

	return resp, nil
}

func (r *ProductRepository) CheckSKUExistsByUser(ctx context.Context, sku string, userID uuid.UUID) (bool, error) {
	exists, err := r.db.Queries.CheckSKUExistsByUser(ctx, db.CheckSKUExistsByUserParams{
		Sku:    sku,
		UserID: userID,
	})
	if err != nil {
		return false, err
	}
	return exists, nil
}

// repositories/product_repository.go
func (r *ProductRepository) ListProducts(ctx context.Context, req models.ListProductsRequest) ([]models.ProductResponse, error) {
	params := db.ListProductsParams{
		Column1: req.ProductID,     // $1
		Column2: req.SKU,           // $2
		Column3: req.Category,      // $3
		Column4: req.SortBy,        // $4
		Limit:   int32(req.Limit),  // $5
		Offset:  int32(req.Offset), // $6
	}

	dbProducts, err := r.db.Queries.ListProducts(ctx, params)
	if err != nil {
		return nil, err
	}

	var products []models.ProductResponse
	for _, p := range dbProducts {
		products = append(products, models.ProductResponse{
			ProductID:        p.ID,
			Name:             p.Name,
			Category:         p.Category,
			Qty:              int(p.Qty),
			Price:            int(p.Price),
			SKU:              p.Sku,
			FileID:           p.FileID,
			FileURI:          "",
			FileThumbnailURI: "",
			CreatedAt:        p.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt:        p.UpdatedAt.Time.Format(time.RFC3339),
		})
	}

	return products, nil
}
