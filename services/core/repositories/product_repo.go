package repositories

import (
	"context"
	"log"
	"time"

	"tutuplapak-core/config"
	"tutuplapak-core/internal/db"
	"tutuplapak-core/models"

	"github.com/google/uuid"
)

type ProductRepositoryInterface interface {
	CreateProduct(ctx context.Context, req models.CreateProductRequest) (models.ProductResponse, error)
	CheckSKUExistsByUser(ctx context.Context, sku string, userID uuid.UUID) (bool, error)
	GetAllProducts(ctx context.Context, params GetAllProductsParams) ([]models.Product, error)
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
		ID:        productID,
		Name:      req.Name,
		Category:  req.Category,
		Qty:       int32(req.Qty),
		Price:     int32(req.Price),
		Sku:       req.SKU,
		FileID:    req.FileID,
		UserID:    req.UserID,
		CreatedAt: time.Now().UTC(), // ✅ WAJIB diisi dari Go
		UpdatedAt: time.Now().UTC(), // ✅ WAJIB diisi dari Go
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

type GetAllProductsParams struct {
	Limit     int
	Offset    int
	ProductID *uuid.UUID
	SKU       *string
	Category  *string
	SortBy    *string
}

func (r *ProductRepository) GetAllProducts(ctx context.Context, params GetAllProductsParams) ([]models.Product, error) {
	limit := int32(5)
	if params.Limit > 0 {
		limit = int32(params.Limit)
	}

	offset := int32(0)
	if params.Offset >= 0 {
		offset = int32(params.Offset)
	}

	args := db.GetAllProductsParams{
		Column4: limit,
		Column5: offset,
	}

	// Untuk ProductID: jika nil → biarkan Column1 = nil (default)
	// Jika tidak nil → assign nilai UUID
	if params.ProductID != nil {
		args.Column1 = *params.ProductID // assign uuid.UUID
	} // else: biarkan nil → NULLIF(nil, ...) → NULL → COALESCE → p.id = p.id

	// Untuk SKU: jika nil → assign "" (sentinel)
	// Jika tidak nil → assign nilai string
	if params.SKU != nil {
		args.Column2 = *params.SKU
	} else {
		args.Column2 = "" // sentinel untuk NULLIF
	}

	// Untuk Category: jika nil → assign ""
	// Jika tidak nil → assign nilai string
	if params.Category != nil {
		args.Column3 = *params.Category
	} else {
		args.Column3 = "" // sentinel untuk NULLIF
	}

	// Untuk SortBy: jika nil → assign ""
	// Jika tidak nil → assign nilai string
	if params.SortBy != nil {
		args.Column6 = *params.SortBy
	} else {
		args.Column6 = "" // default sort (created_at DESC)
	}

	log.Printf("Query Args: %+v", args) // Debug log

	rows, err := r.db.Queries.GetAllProducts(ctx, args)
	if err != nil {
		return nil, err
	}

	var products []models.Product
	for _, row := range rows {
		products = append(products, models.Product{
			ID:        row.ID,
			Name:      row.Name,
			Category:  row.Category,
			Qty:       int(row.Qty),
			Price:     int(row.Price),
			SKU:       row.Sku,
			FileID:    row.FileID,
			UserID:    row.UserID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}

	log.Printf("Retrieved %d products", len(products)) // Debug log

	return products, nil
}
