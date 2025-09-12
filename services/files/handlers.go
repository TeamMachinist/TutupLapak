package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/teammachinist/tutuplapak/internal"

	"github.com/go-chi/chi/v5"
	// "github.com/h2non/bimg"
	"github.com/nfnt/resize"
)

type FileHandler struct {
	fileService FileService
	storage     *MinIOStorage
}

func NewFileHandler(minioStorage *MinIOStorage, fileService FileService) *FileHandler {
	return &FileHandler{
		storage:     minioStorage,
		fileService: fileService,
	}
}

func (h *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	fileIdStr := chi.URLParam(r, "fileId")

	fileId, err := uuid.Parse(fileIdStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.fileService.DeleteFiles(r.Context(), fileId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File Deleted"))
}

func (h *FileHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	fileIdStr := chi.URLParam(r, "fileId")

	fileId, err := uuid.Parse(fileIdStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, err := h.fileService.GetFile(r.Context(), fileId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(file)
}

func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	requestCtx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Get user ID from auth middleware context (for internal tracking)
	userID, ok := internal.GetUserIDFromChi(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	file, header, err := r.FormFile("file")
	ctx := r.Context()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest) // Spec-compliant error
		return
	}
	defer file.Close()

	// Spec says max 100KiB, not 10MB
	if header.Size > 100*1024 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Validate file type - check extension (more reliable than content-type header)
	filename := strings.ToLower(header.Filename)
	validExts := []string{".jpg", ".jpeg", ".png"}

	isValidType := false
	for _, ext := range validExts {
		if strings.HasSuffix(filename, ext) {
			isValidType = true
			break
		}
	}

	if !isValidType {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	fileId := uuid.Must(uuid.NewV7())
	filename = fmt.Sprintf("%s_%s", fileId, header.Filename)

	// Upload original image to MinIO
	uri, err := h.storage.UploadFile(ctx, file, header, header.Size, filename)
	if err != nil {
		log.Println("error uploading file:", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	buffer, err := io.ReadAll(file)
	if err != nil {
		log.Println("error read buffer:", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	// Create thumbnail
	compressedImage, imageSize, err := h.compressImageNFNT(buffer, 10, "uploads")
	if err != nil {
		log.Println("error compress image nfnt:", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	compressedReader := bytes.NewReader(compressedImage)
	compressedImageName := "compressed-" + filename

	// Upload thumbnail to MinIO
	compressedImageUri, err := h.storage.UploadFile(ctx, compressedReader, header, imageSize, compressedImageName)
	if err != nil {
		log.Println("error uploading compressed image:", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Println("error parsing user ID:", err)
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	// Create file record with user ID (internal tracking)
	data := File{
		ID:               fileId,
		UserID:           userUUID, // Track ownership internally
		FileURI:          uri,
		FileThumbnailURI: compressedImageUri,
	}

	_, err = h.fileService.CreateFiles(requestCtx, data)
	if err != nil {
		log.Println("error creating file:", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	// Response matches spec format (no user_id exposed)
	response := map[string]interface{}{
		"id":                 data.ID,
		"file_uri":           data.FileURI,
		"file_thumbnail_uri": data.FileThumbnailURI,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Spec says "Ok" response
	json.NewEncoder(w).Encode(response)
}

// Original bimg compression function (commented)
// func (h *FileHandler) compressImageBimg(buffer []byte, quality int, dirname string) ([]byte, int64, error) {
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

// NFNT compression function - same approach as bimg
func (h *FileHandler) compressImageNFNT(buffer []byte, quality int, dirname string) ([]byte, int64, error) {
	// Decode image from buffer (similar to bimg.NewImage)
	img, format, err := image.Decode(bytes.NewReader(buffer))
	if err != nil {
		return nil, 0, err
	}

	// Resize for compression (similar to bimg quality reduction)
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
