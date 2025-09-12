package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"tutuplapak-files/db"

	"github.com/go-chi/chi/v5"
)

type FileHandler struct {
	fileService FileService
	storage *MinIOStorage

}

func NewFileHandler(minioStorage *MinIOStorage, fileService FileService) *FileHandler {
	return &FileHandler{
		storage: minioStorage,
		fileService: fileService,
	}
}

func (h *FileHandler) CreateFiles(w http.ResponseWriter,r *http.Request,  payload db.CreateFilesParams) {
newFile, err := h.fileService.CreateFiles(r.Context(), payload)
if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))	
	return 
}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(newFile)
 
}

func (h *FileHandler) DeleteFiles(w http.ResponseWriter,r *http.Request) {
	fileId := chi.URLParam(r, "fileid")
	err := h.fileService.DeleteFiles(r.Context(), fileId)
	if err!=nil{

		return 
	}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("File Deleted"))

}
func (h *FileHandler) GetFiles(w http.ResponseWriter,r *http.Request){
	fileId := chi.URLParam(r, "fileid")
	file, err := h.fileService.GetFiles(r.Context(), fileId)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))	
		return 
	}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(file)
}
func (h *FileHandler) ListFiles(w http.ResponseWriter,r *http.Request){
	requestCtx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	listFile, err := h.fileService.ListFiles(requestCtx)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))		
		return 
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(listFile)
}


func (h *FileHandler) UploadFile(w http.ResponseWriter,r *http.Request) {
	requestCtx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	file, header, err := r.FormFile("picture")
	
	ctx := r.Context()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("File is Required"))
		return
	}
	defer file.Close()

	// Validate file size (max 10MB)
	if header.Size > 10*1024*1024 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("File size exceeds 10MB limit"))
		return
	}

	// Validate file type (jpeg/jpg/png)
	contentType := header.Header.Get("Content-Type")
	allowedTypes := []string{"image/jpeg", "image/jpg", "image/png"}

	isValidType := false
	for _, allowedType := range allowedTypes {
		if strings.Contains(contentType, allowedType) {
			isValidType = true
			break
		}
	}

	if !isValidType {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Only JPEG and PNG files are allowed"))
		
		return
	}
	id := time.Now().Unix()
	fileId :=   fmt.Sprintf("%d", id) 
	filename := fmt.Sprintf("%s_%s", fileId, header.Filename)
	// Upload original image to MinIO
	uri, err := h.storage.UploadFile(ctx, file, header, header.Size, filename)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to upload File"))
		w.Write([]byte(err.Error()))
		return
	}


	buffer, err := io.ReadAll(file) //save the data inside buffer
	if err != nil{
			panic(err)
	}

	// upload compressed image
	compressedImage, imageSize, err := CompressImage(buffer, 10, "uploads")
	if err !=nil{
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to upload File"))

	}
	compressedReader := bytes.NewReader(compressedImage)
	compressedImageName := "compressed-" + filename
	
	//upload compressed image to MinIO
	compressedImageUri, err := h.storage.UploadFile(ctx, compressedReader, header, imageSize, compressedImageName)
	
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to upload Compressed Image"))
		w.Write([]byte(err.Error()))

		return
	}
	w.WriteHeader(http.StatusOK)
	data := db.CreateFilesParams{
		Fileid: fileId,
		Fileuri: uri,
		Filethumbnailuri: compressedImageUri,
	}
	h.fileService.CreateFiles(requestCtx, data)
	json.NewEncoder(w).Encode(data)

}