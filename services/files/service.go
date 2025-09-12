package main

import (
	"context"
	"mime/multipart"
	"tutuplapak-files/db"

	"github.com/h2non/bimg"
)

type FileStorage interface {
	Upload(ctx context.Context, objectKey, contentType string, file multipart.File, size int64) (string, error)
	CompressImage(buffer []byte, quality int, dirname string) ([]byte, int64, error)
}

type FileService struct {
	userRepo FileRepository
	db      *db.Queries
	// storage FileStorage
}

func NewFileService(db *db.Queries,  userRepo FileRepository) FileService {
	return FileService{db: db, userRepo: userRepo}
}

func (s *FileService) CreateFiles(ctx context.Context, payload db.CreateFilesParams) (db.File, error){
newFile, err := s.userRepo.CreateFiles(ctx, payload)
if err != nil{
	return db.File{}, err
}
return newFile, nil
}

func (s *FileService) DeleteFiles(ctx context.Context, fileId string) (error){
	err := s.db.DeleteFiles(ctx, fileId)
	if err!=nil{
		return err
	}
	return nil
}
func (s *FileService) GetFiles(ctx context.Context, fileid string) (db.File, error){
	file, err := s.db.GetFiles(ctx, fileid)
	if err != nil{
		return db.File{}, err
	}
	return file, nil
}
func (s *FileService) ListFiles(ctx context.Context) ([]db.File, error){
	listFile, err := s.db.ListFiles(ctx)
	if err != nil{
		return []db.File{}, err
	}
	return listFile, nil
}

func CompressImage(buffer []byte, quality int, dirname string) ([]byte, int64,error){
	converted, err := bimg.NewImage(buffer).Convert(bimg.JPEG) // convert image to JPEG
	if err !=nil {
		return nil, 0, err
	}
	//compress the image
	processed, err := bimg.NewImage(converted).Process(bimg.Options{Quality: quality, StripMetadata: true}) 
	if err != nil{
		return nil, 0, err
	}

	return processed, int64(len(processed)) ,nil

}