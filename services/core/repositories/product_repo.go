package repositories

import (
	"context"
	"fmt"
	"log"
	"time"

	"tutuplapak-core/config"
	"tutuplapak-core/internal/db"
	"tutuplapak-core/models"

	"github.com/google/uuid"
)

type ProductRepositoryInterface interface {
	CreateProduct(ctx context.Context, req models.ProductRequest) (models.ProductResponse, error)
	CheckSKUExistsByUser(ctx context.Context, sku string, userID uuid.UUID) (uuid.UUID, error)
	GetAllProducts(ctx context.Context, params models.GetAllProductsParams) ([]models.Product, error)
	UpdateProduct(ctx context.Context, params db.UpdateProductParams) (db.UpdateProductRow, error)
	CheckProductOwnership(ctx context.Context, productID uuid.UUID, userID uuid.UUID) (bool, error)
	DeleteProduct(ctx context.Context, productID uuid.UUID, userID uuid.UUID) error
}

type ProductRepository struct {
	db *config.Database
}

func NewProductRepository(database *config.Database) ProductRepositoryInterface {
	return &ProductRepository{db: database}
}

func (r *ProductRepository) CreateProduct(ctx context.Context, req models.ProductRequest) (models.ProductResponse, error) {
	productID := uuid.Must(uuid.NewV7())

	dbProduct, err := r.db.Queries.CreateProduct(ctx, db.CreateProductParams{
		ID:        productID,
		Name:      req.Name,
		Category:  req.Category,
		Qty:       req.Qty,
		Price:     req.Price,
		Sku:       req.SKU,
		FileID:    req.FileID,
		UserID:    req.UserID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
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
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return resp, nil
}

func (r *ProductRepository) CheckSKUExistsByUser(ctx context.Context, sku string, userID uuid.UUID) (uuid.UUID, error) {
	productID, err := r.db.Queries.CheckSKUExistsByUser(ctx, db.CheckSKUExistsByUserParams{
		Sku:    sku,
		UserID: userID,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return productID, nil
}

func (r *ProductRepository) GetAllProducts(ctx context.Context, params models.GetAllProductsParams) ([]models.Product, error) {
	limit := int32(5)
	if params.Limit > 0 {
		limit = int32(params.Limit)
	}

	offset := int32(0)
	if params.Offset >= 0 {
		offset = int32(params.Offset)
	}

	args := db.GetAllProductsParams{
		LimitCount:  limit,
		OffsetCount: offset,
		ProductID:   uuid.Nil,
		Sku:         "",
		Category:    "",
		SortBy:      "newest",
	}

	if params.ProductID != nil {
		args.ProductID = *params.ProductID
	}

	if params.SKU != nil {
		args.Sku = *params.SKU
	}

	if params.Category != nil {
		args.Category = *params.Category
	}
	if params.SortBy != nil {
		args.SortBy = *params.SortBy
	}

	log.Printf("Query Args: %+v", args)

	rows, err := r.db.Queries.GetAllProducts(ctx, args)
	if err != nil {
		fmt.Println("Error fetching products:", err)
		return nil, err
	}

	products := make([]models.Product, len(rows))
	for i, row := range rows {
		products[i] = models.Product{
			ID:               row.ID,
			Name:             row.Name,
			Category:         row.Category,
			Qty:              row.Qty,
			Price:            row.Price,
			SKU:              row.Sku,
			FileID:           row.FileID,
			FileURI:          "",
			FileThumbnailURI: "",
			UserID:           row.UserID,
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
		}
	}

	log.Printf("Retrieved %d products", len(products))
	return products, nil
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, params db.UpdateProductParams) (db.UpdateProductRow, error) {
	return r.db.Queries.UpdateProduct(ctx, params)
}

func (r *ProductRepository) CheckProductOwnership(ctx context.Context, productID uuid.UUID, userID uuid.UUID) (bool, error) {
	result, err := r.db.Queries.CheckProductOwnership(ctx, db.CheckProductOwnershipParams{
		ProductID: productID,
		UserID:    userID,
	})
	if err != nil {
		return false, err
	}

	return result, nil
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, productID uuid.UUID, userID uuid.UUID) error {
	err := r.db.Queries.DeleteProduct(ctx, db.DeleteProductParams{
		ID:     productID,
		UserID: userID,
	})
	return err
}
