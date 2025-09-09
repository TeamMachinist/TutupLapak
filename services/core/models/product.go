package models

import "github.com/google/uuid"

type Product struct {
	ID               uuid.UUID `json:"productId"`
	Name             string    `json:"name"`
	Category         string    `json:"category"`
	Qty              int       `json:"qty"`
	Price            int       `json:"price"`
	SKU              string    `json:"sku"`
	FileID           uuid.UUID `json:"fileId"`
	FileURI          string    `json:"fileUri"`
	FileThumbnailURI string    `json:"fileThumbnailUri"`
	UserID           uuid.UUID `json:"userId"`
	CreatedAt        string    `json:"createdAt"`
	UpdatedAt        string    `json:"updatedAt"`
}

type CreateProductRequest struct {
	Name     string    `json:"name" validate:"required,min=4,max=32"`
	Category string    `json:"category" validate:"required"`
	Qty      int       `json:"qty" validate:"required,min=1"`
	Price    int       `json:"price" validate:"required,min=100"`
	SKU      string    `json:"sku" validate:"required,max=32"`
	FileID   uuid.UUID `json:"fileId" `
	UserID   uuid.UUID `json:"userId" `
}

type ProductResponse struct {
	ProductID        uuid.UUID `json:"productId"`
	Name             string    `json:"name"`
	Category         string    `json:"category"`
	Qty              int       `json:"qty"`
	Price            int       `json:"price"`
	SKU              string    `json:"sku"`
	FileID           uuid.UUID `json:"fileId"`
	FileURI          string    `json:"fileUri"`
	FileThumbnailURI string    `json:"fileThumbnailUri"`
	CreatedAt        string    `json:"createdAt"`
	UpdatedAt        string    `json:"updatedAt"`
}

type ListProductsRequest struct {
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
	ProductID uuid.UUID `json:"productId"`
	SKU       string    `json:"sku"`
	Category  string    `json:"category"`
	SortBy    string    `json:"sortBy"`
}
