package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"strings"

	"github.com/google/uuid"
	"github.com/teammachinist/tutuplapak/internal/database"

	// "github.com/h2non/bimg"
	"github.com/nfnt/resize"
)

type FileStorage interface {
	Upload(ctx context.Context, objectKey, contentType string, file multipart.File, size int64) (string, error)
	CompressImage(buffer []byte, quality int, dirname string) ([]byte, int64, error)
}

type FileService struct {
	db *database.Queries
	// userRepo FileRepository
	// storage FileStorage
}

func NewFileService(db *database.Queries) FileService {
	return FileService{db: db}
}

func (s *FileService) CreateFiles(ctx context.Context, file File) (database.Files, error) {
	newFile, err := s.db.CreateFile(ctx, database.CreateFileParams{
		ID:               file.ID,
		UserID:           file.UserID,
		FileUri:          file.FileURI,
		FileThumbnailUri: file.FileThumbnailURI,
	})
	if err != nil {
		return database.Files{}, err
	}
	return newFile, nil
}

func (s *FileService) DeleteFiles(ctx context.Context, fileId uuid.UUID) error {
	file, _ := s.GetFile(ctx, fileId)
	if file == (database.Files{}) {
		return fmt.Errorf("file doesn't exist")
	}

	err := s.db.DeleteFile(ctx, fileId)
	if err != nil {
		return err
	}
	return nil
}

func (s *FileService) GetFile(ctx context.Context, fileId uuid.UUID) (database.Files, error) {
	file, err := s.db.GetFile(ctx, fileId)
	if err != nil {
		return database.Files{}, err
	}
	return file, nil
}

// Original bimg implementation (commented for Docker compatibility)
// func CompressImage(buffer []byte, quality int, dirname string) ([]byte, int64, error) {
// 	converted, err := bimg.NewImage(buffer).Convert(bimg.JPEG) // convert image to JPEG
// 	if err != nil {
// 		return nil, 0, err
// 	}
// 	//compress the image
// 	processed, err := bimg.NewImage(converted).Process(bimg.Options{Quality: quality, StripMetadata: true})
// 	if err != nil {
// 		return nil, 0, err
// 	}

// 	return processed, int64(len(processed)), nil
// }

// NFNT implementation
func CompressImage(buffer []byte, quality int, dirname string) ([]byte, int64, error) {
	// Decode image from buffer (similar to bimg.NewImage)
	img, format, err := image.Decode(bytes.NewReader(buffer))
	if err != nil {
		return nil, 0, err
	}

	// Resize image for compression (similar to bimg.Process with quality)
	// Using 800x600 max to reduce file size like quality compression
	resized := resize.Thumbnail(800, 600, img, resize.Lanczos3)

	// Convert to JPEG (similar to bimg.Convert(bimg.JPEG))
	var processed bytes.Buffer
	err = jpeg.Encode(&processed, resized, &jpeg.Options{Quality: quality})
	if err != nil {
		// Fallback to original format if JPEG fails
		if strings.ToLower(format) == "png" {
			processed.Reset()
			err = png.Encode(&processed, resized)
		}
		if err != nil {
			return nil, 0, err
		}
	}

	return processed.Bytes(), int64(processed.Len()), nil
}
