package repositories

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/teammachinist/tutuplapak/internal/database"
	"github.com/teammachinist/tutuplapak/services/core/models"

	"github.com/google/uuid"
)

type ProductRepositoryInterface interface {
	CreateProduct(ctx context.Context, req models.ProductRequest) (models.ProductResponse, error)
	CheckSKUExistsByUser(ctx context.Context, sku string, userID uuid.UUID) (CheckSKUExistsByUserRow, error)
	GetAllProducts(ctx context.Context, params models.GetAllProductsParams) ([]models.Product, error)
	UpdateProduct(ctx context.Context, params database.UpdateProductParams) (database.UpdateProductRow, error)
	CheckProductOwnership(ctx context.Context, productID uuid.UUID, userID uuid.UUID) (bool, error)
	DeleteProduct(ctx context.Context, productID uuid.UUID, userID uuid.UUID) error
}

type ProductRepository struct {
	db database.Querier
}

func NewProductRepository(database database.Querier) ProductRepositoryInterface {
	return &ProductRepository{db: database}
}

func (r *ProductRepository) CreateProduct(ctx context.Context, req models.ProductRequest) (models.ProductResponse, error) {
	productID := uuid.Must(uuid.NewV7())

	dbProduct, err := r.db.CreateProduct(ctx, database.CreateProductParams{
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

type CheckSKUExistsByUserRow struct {
	ID  uuid.UUID `json:"id"`
	Sku string    `json:"sku"`
}

func (r *ProductRepository) CheckSKUExistsByUser(ctx context.Context, sku string, userID uuid.UUID) (CheckSKUExistsByUserRow, error) {
	row, err := r.db.CheckSKUExistsByUser(ctx, database.CheckSKUExistsByUserParams{
		Sku:    sku,
		UserID: userID,
	})
	if err != nil {
		return CheckSKUExistsByUserRow{}, err
	}
	return CheckSKUExistsByUserRow{
		ID:  row.ID,
		Sku: row.Sku,
	}, nil
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

	args := database.GetAllProductsParams{
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

	rows, err := r.db.GetAllProducts(ctx, args)
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

func (r *ProductRepository) UpdateProduct(ctx context.Context, params database.UpdateProductParams) (database.UpdateProductRow, error) {
	return r.db.UpdateProduct(ctx, params)
}

func (r *ProductRepository) CheckProductOwnership(ctx context.Context, productID uuid.UUID, userID uuid.UUID) (bool, error) {
	result, err := r.db.CheckProductOwnership(ctx, database.CheckProductOwnershipParams{
		ProductID: productID,
		UserID:    userID,
	})
	if err != nil {
		return false, err
	}

	return result, nil
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, productID uuid.UUID, userID uuid.UUID) error {
	err := r.db.DeleteProduct(ctx, database.DeleteProductParams{
		ID:     productID,
		UserID: userID,
	})
	return err
}
