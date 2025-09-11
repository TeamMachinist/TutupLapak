package models

import (
	"time"

	"github.com/google/uuid"
)

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
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type ProductRequest struct {
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
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type GetAllProductsParams struct {
	Limit     int
	Offset    int
	ProductID *uuid.UUID
	SKU       *string
	Category  *string
	SortBy    *string
}
