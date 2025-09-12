package main

import (
	"context"
	"tutuplapak-files/db"
)	

type FileRepository struct{
	db *db.Queries
}

func NewFileRepository(db *db.Queries) FileRepository {
	return FileRepository{db: db}
}

func (q *FileRepository) CreateFiles(ctx context.Context, payload db.CreateFilesParams) (db.File, error){
	row, err := q.db.CreateFiles(ctx, payload)
	if err != nil{
		return db.File{}, err
	}
	return row, nil
}
func (q *FileRepository) DeleteFiles(ctx context.Context, fileid string) (error){
	err := q.db.DeleteFiles(ctx, fileid)
	if err != nil{
		return nil
	}
	return nil
}
func (q *FileRepository) GetFiles(ctx context.Context, fileid string) (db.File, error){
	row, err := q.db.GetFiles(ctx, fileid)
	if err != nil{
		return db.File{}, err
	}
	return row, nil
}
func (q *FileRepository) ListFiles(ctx context.Context) ([]db.File, error){
	row, err := q.db.ListFiles(ctx)
	if err != nil{
		return nil, err
	}
	return row, nil
}