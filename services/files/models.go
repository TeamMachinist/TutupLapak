package main

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID               uuid.UUID `json:"productId"`
	UserID           uuid.UUID `json:"userId"`
	FileURI          string    `json:"fileUri"`
	FileThumbnailURI string    `json:"fileThumbnailUri"`
	CreatedAt        time.Time `json:"createdAt"`
}
