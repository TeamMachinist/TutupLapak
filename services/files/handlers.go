package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type FileHandler struct {
	storage *MinIOStorage
}

func NewFileHandler(minioStorage *MinIOStorage) *FileHandler {
	return &FileHandler{
		storage: minioStorage,
	}
}

func (h *FileHandler) UploadFile(w http.ResponseWriter,r *http.Request) {
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

	// Upload original image to MinIO
	uri, err := h.storage.UploadFile(ctx, file, header, header.Size)
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
	//upload compressed image to MinIO
	compressedImageUri, err := h.storage.UploadFile(ctx, compressedReader, header, imageSize)
	
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to upload Compressed Image"))
		w.Write([]byte(err.Error()))

		return
	}
	w.WriteHeader(http.StatusOK)
	 data := map[string]interface{}{
        "fileid": "",
        "fileUri": uri,
		"fileThumbnailUri": compressedImageUri,
    }
	json.NewEncoder(w).Encode(data)

}