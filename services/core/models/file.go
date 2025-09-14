package models

type File struct {
	ID           string `json:"id"`
	URI          string `json:"uri"`
	ThumbnailURI string `json:"thumbnailUri"`
}
