package handler

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/teammachinist/tutuplapak/internal"
	"github.com/teammachinist/tutuplapak/services/files/internal/api"
	"github.com/teammachinist/tutuplapak/services/files/internal/logger"
	"github.com/teammachinist/tutuplapak/services/files/internal/model"
	"github.com/teammachinist/tutuplapak/services/files/internal/service"
	"github.com/teammachinist/tutuplapak/services/files/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/nfnt/resize"
)

type FileHandler struct {
	fileService service.FileService
	storage     *storage.MinIOStorage
}

func NewFileHandler(minioStorage *storage.MinIOStorage, fileService service.FileService) *FileHandler {
	return &FileHandler{
		storage:     minioStorage,
		fileService: fileService,
	}
}

func (h *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileIdStr := chi.URLParam(r, "fileId")

	logger.InfoCtx(ctx, "Delete file request", "file_id", fileIdStr)

	// Validate file ID format
	fileId, err := uuid.Parse(fileIdStr)
	if err != nil {
		logger.WarnCtx(ctx, "Invalid file ID format", "file_id", fileIdStr, "error", err)
		api.WriteBadRequest(w, r, "Invalid file ID format")
		return
	}

	// TODO: Use authz package
	// Get user ID from auth middleware
	userID, ok := internal.GetUserIDFromChi(r)
	if !ok {
		logger.ErrorCtx(ctx, "Missing user ID from auth context")
		api.WriteUnauthorized(w, r, "Unauthorized")
		return
	}

	// Verify file ownership and delete
	err = h.fileService.DeleteFiles(ctx, fileId, userID)
	if err != nil {
		if err.Error() == "file not found" || err.Error() == "not found" {
			logger.WarnCtx(ctx, "File not found for deletion", "file_id", fileId, "user_id", userID)
			api.WriteNotFound(w, r, "File not found")
			return
		}
		if err.Error() == "unauthorized" || err.Error() == "forbidden" {
			logger.WarnCtx(ctx, "Unauthorized file deletion attempt", "file_id", fileId, "user_id", userID)
			api.WriteError(w, r, http.StatusForbidden, "You can only delete your own files")
			return
		}

		logger.ErrorCtx(ctx, "Failed to delete file", "file_id", fileId, "user_id", userID, "error", err)
		api.WriteInternalServerError(w, r, "Failed to delete file")
		return
	}

	logger.InfoCtx(ctx, "File deleted successfully", "file_id", fileId, "user_id", userID)

	// Return simple success message as per spec
	response := map[string]string{"message": "File deleted successfully"}
	api.WriteSuccess(w, r, response)
}

func (h *FileHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileIdStr := chi.URLParam(r, "fileId")

	logger.InfoCtx(ctx, "Get file request", "file_id", fileIdStr)

	// Validate file ID format
	fileId, err := uuid.Parse(fileIdStr)
	if err != nil {
		logger.WarnCtx(ctx, "Invalid file ID format", "file_id", fileIdStr, "error", err)
		api.WriteBadRequest(w, r, "Invalid file ID format")
		return
	}

	// Get file with caching
	file, err := h.fileService.GetFile(ctx, fileId)
	if err != nil {
		if err.Error() == "file not found" || err.Error() == "not found" {
			logger.WarnCtx(ctx, "File not found", "file_id", fileId)
			logger.ErrorCtx(ctx, "masuk sini", "error", err)
			api.WriteNotFound(w, r, "File not found")
			return
		}

		logger.ErrorCtx(ctx, "Failed to get file", "file_id", fileId, "error", err)
		api.WriteInternalServerError(w, r, "Failed to retrieve file")
		return
	}

	logger.InfoCtx(ctx, "File retrieved successfully", "file_id", fileId)

	// Format response according to API spec (no user_id exposed)
	response := map[string]interface{}{
		"id":                 file.ID,
		"user_id":            file.UserID,
		"file_uri":           file.FileURI,
		"file_thumbnail_uri": file.FileThumbnailURI,
		"created_at":         file.CreatedAt,
	}

	api.WriteSuccess(w, r, response)
}

func (h *FileHandler) GetFiles(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	idParam := r.URL.Query().Get("id")
	fmt.Print(idParam)
	logger.InfoCtx(ctx, "Get file request", "file_id", idParam)

	if idParam == "" {
		logger.WarnCtx(ctx, "Missing file ID parameter")
		api.WriteBadRequest(w, r, "Need the ID parameter")
		return
	}

	ids := strings.Split(idParam, ",")
	for i, id := range ids {
		ids[i] = strings.TrimSpace(id)
	}

	var uuidIDs []uuid.UUID
	for _, id := range ids {
		uid, err := uuid.Parse(id)
		if err != nil {
			logger.WarnCtx(ctx, "Invalid file ID format", "file_id", id, "error", err)
			api.WriteBadRequest(w, r, fmt.Sprintf("Invalid file ID format: %s", id))
			return
		}
		uuidIDs = append(uuidIDs, uid)
	}

	// Dapatkan file dari service
	files, err := h.fileService.GetFiles(ctx, uuidIDs)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to get files", "error", err)
		api.WriteInternalServerError(w, r, "Failed to retrieve files")
		return
	}

	logger.InfoCtx(ctx, "Files retrieved successfully", "file_ids", files)
	api.WriteSuccess(w, r, files)
}
func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	requestCtx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.InfoCtx(requestCtx, "File upload request started")

	// TODO: Use authz package
	// // Get user ID from auth middleware context
	userID, ok := internal.GetUserIDFromChi(r)
	if !ok {
		logger.ErrorCtx(requestCtx, "Missing user ID from auth context")
		api.WriteUnauthorized(w, r, "Unauthorized")
		return
	}

	// Parse multipart form
	file, header, err := r.FormFile("file")
	if err != nil {
		logger.WarnCtx(requestCtx, "Failed to parse multipart form", "error", err)
		api.WriteBadRequest(w, r, "Invalid file upload request")
		return
	}
	defer file.Close()

	logger.InfoCtx(requestCtx, "File upload details",
		"user_id", userID,
		"filename", header.Filename,
		"size", header.Size,
		"content_type", header.Header.Get("Content-Type"),
	)

	// Validate file size (100KiB as per spec)
	const maxFileSize = 100 * 1024
	if header.Size > maxFileSize {
		logger.WarnCtx(requestCtx, "File too large",
			"size", header.Size,
			"max_size", maxFileSize,
			"filename", header.Filename,
		)
		api.WriteBadRequest(w, r, "File size exceeds 100KiB limit")
		return
	}

	// Validate file type by extension
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
		logger.WarnCtx(requestCtx, "Invalid file type",
			"filename", header.Filename,
			"valid_extensions", validExts,
		)
		api.WriteBadRequest(w, r, "Only JPG, JPEG, and PNG files are allowed")
		return
	}

	// Generate file ID and name
	fileId := uuid.Must(uuid.NewV7())
	storageName := fmt.Sprintf("%s_%s", fileId, header.Filename)

	logger.InfoCtx(requestCtx, "Processing file upload",
		"file_id", fileId,
		"storage_name", storageName,
	)

	// Upload original image to MinIO
	uri, err := h.storage.UploadFile(requestCtx, file, header, header.Size, storageName)
	if err != nil {
		logger.ErrorCtx(requestCtx, "Failed to upload original file to MinIO",
			"error", err,
			"file_id", fileId,
			"filename", header.Filename,
		)
		api.WriteInternalServerError(w, r, "Failed to upload file")
		return
	}

	logger.InfoCtx(requestCtx, "Original file uploaded to MinIO", "file_id", fileId, "uri", uri)

	// Read file content for thumbnail generation
	buffer, err := io.ReadAll(file)
	if err != nil {
		logger.ErrorCtx(requestCtx, "Failed to read file buffer", "error", err, "file_id", fileId)
		api.WriteInternalServerError(w, r, "Failed to process file")
		return
	}

	// Create thumbnail
	compressedImage, imageSize, err := h.compressImageNFNT(buffer, 10, "uploads")
	if err != nil {
		logger.ErrorCtx(requestCtx, "Failed to create thumbnail", "error", err, "file_id", fileId)
		api.WriteInternalServerError(w, r, "Failed to create thumbnail")
		return
	}

	compressedReader := bytes.NewReader(compressedImage)
	compressedImageName := "compressed-" + storageName

	// Upload thumbnail to MinIO
	compressedImageUri, err := h.storage.UploadFile(requestCtx, compressedReader, header, imageSize, compressedImageName)
	if err != nil {
		logger.ErrorCtx(requestCtx, "Failed to upload thumbnail to MinIO",
			"error", err,
			"file_id", fileId,
			"compressed_name", compressedImageName,
		)
		api.WriteInternalServerError(w, r, "Failed to upload thumbnail")
		return
	}

	logger.InfoCtx(requestCtx, "Thumbnail uploaded to MinIO", "file_id", fileId, "uri", compressedImageUri)

	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		logger.ErrorCtx(requestCtx, "Invalid user ID format", "user_id", userID, "error", err)
		api.WriteBadRequest(w, r, "Invalid user ID")
		return
	}

	// Create file record
	fileData := model.File{
		ID:               fileId,
		UserID:           userUUID,
		FileURI:          uri,
		FileThumbnailURI: compressedImageUri,
	}

	createdFile, err := h.fileService.CreateFiles(requestCtx, fileData)
	if err != nil {
		logger.ErrorCtx(requestCtx, "Failed to create file record",
			"error", err,
			"file_id", fileId,
			"user_id", userID,
		)
		api.WriteInternalServerError(w, r, "Failed to save file record")
		return
	}

	logger.InfoCtx(requestCtx, "File upload completed successfully",
		"file_id", fileId,
		"user_id", userID,
		"original_uri", uri,
		"thumbnail_uri", compressedImageUri,
	)

	// Response matches spec format (no user_id exposed) - return data directly
	response := map[string]interface{}{
		"id":                 createdFile.ID,
		"file_uri":           createdFile.FileURI,
		"file_thumbnail_uri": createdFile.FileThumbnailURI,
	}

	api.WriteSuccess(w, r, response)
}

// NFNT compression function - same approach as bimg
func (h *FileHandler) compressImageNFNT(buffer []byte, quality int, dirname string) ([]byte, int64, error) {
	// Decode image from buffer (similar to bimg.NewImage)
	img, format, err := image.Decode(bytes.NewReader(buffer))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode image: %w", err)
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
			return nil, 0, fmt.Errorf("failed to encode compressed image: %w", err)
		}
	}

	return processed.Bytes(), int64(processed.Len()), nil
}
