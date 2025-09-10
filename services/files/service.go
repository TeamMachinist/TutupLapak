package main

import (
	"context"
	"database/sql"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/google/uuid"
	"github.com/h2non/bimg"
)

type FileStorage interface {
	Upload(ctx context.Context, objectKey, contentType string, file multipart.File, size int64) (string, error)
}

type FileService struct {
	db      *sql.DB
	storage FileStorage
}

func NewFileService(db *sql.DB, storage FileStorage) *FileService {
	return &FileService{db: db, storage: storage}
}

func (s *FileService) UploadFile(ctx context.Context, userID uuid.UUID, fh *multipart.FileHeader) (string, error) {
	file, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	objectKey := fmt.Sprintf("uploads/%s_%s", userID.String(), fh.Filename)

	key, err := s.storage.Upload(ctx, objectKey, fh.Header.Get("Content-Type"), file, fh.Size)
	if err != nil {
		return "", err
	}

	// Save metadata in DB
	_, err = s.db.ExecContext(ctx, `
        INSERT INTO files (user_id, object_key, mime_type, size)
        VALUES ($1, $2, $3, $4)`,
		userID, key, fh.Header.Get("Content-Type"), fh.Size)
	if err != nil {
		return "", err
	}

	return key, nil
}

func imageProcessing(buffer []byte, quality int, dirname string) (string, error){
	filename := strings.Replace(uuid.New().String()+ ".jpeg", "-","",-1) //image name after converted
	converted, err := bimg.NewImage(buffer).Convert(bimg.JPEG) // convert image to JPEG
	if err !=nil {
		return filename, err
	}
	//compress the image
	processed, err := bimg.NewImage(converted).Process(bimg.Options{Quality: quality, StripMetadata: true}) 
	if err != nil{
		return filename, err
	}
	//Save image in  "uploads" folder
	writeError := bimg.Write(fmt.Sprintf("./" + dirname + "/%s", filename), processed) 
	if writeError != nil{
		return filename, writeError
	}
	return filename, nil

}