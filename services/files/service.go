package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/teammachinist/tutuplapak/internal/cache"
	"github.com/teammachinist/tutuplapak/internal/database"
	"github.com/teammachinist/tutuplapak/internal/logger"
)

type FileService struct {
	queries *database.Queries
	cache   *cache.RedisCache
}

func NewFileService(queries *database.Queries, cache *cache.RedisCache) FileService {
	return FileService{
		queries: queries,
		cache:   cache,
	}
}

func (s FileService) CreateFiles(ctx context.Context, file File) (File, error) {
	logger.InfoCtx(ctx, "Creating file record", "file_id", file.ID, "user_id", file.UserID)

	// Create file record in database
	params := database.CreateFileParams{
		ID:               file.ID,
		UserID:           file.UserID,
		FileUri:          file.FileURI,
		FileThumbnailUri: file.FileThumbnailURI,
	}

	dbFile, err := s.queries.CreateFile(ctx, params)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to create file in database",
			"error", err,
			"file_id", file.ID,
			"user_id", file.UserID,
		)
		return File{}, fmt.Errorf("failed to create file: %w", err)
	}

	// Convert database file to our File model
	createdFile := File{
		ID:               dbFile.ID,
		UserID:           dbFile.UserID,
		FileURI:          dbFile.FileUri,
		FileThumbnailURI: dbFile.FileThumbnailUri,
		CreatedAt:        dbFile.CreatedAt,
	}

	// Cache the file metadata
	cacheKey := fmt.Sprintf(cache.FileMetadataKey, file.ID.String())
	if err := s.cache.Set(ctx, cacheKey, createdFile, cache.FileMetadataTTL); err != nil {
		logger.WarnCtx(ctx, "Failed to cache file metadata after creation",
			"error", err,
			"file_id", file.ID,
			"cache_key", cacheKey,
		)
		// Don't fail the operation for cache errors
	}

	// Invalidate user's file list cache
	userCacheKey := fmt.Sprintf(cache.UserFileListKey, file.UserID.String())
	if err := s.cache.Delete(ctx, userCacheKey); err != nil {
		logger.WarnCtx(ctx, "Failed to invalidate user file list cache",
			"error", err,
			"user_id", file.UserID,
			"cache_key", userCacheKey,
		)
	}

	// Update file exists cache
	existsCacheKey := fmt.Sprintf(cache.FileExistsKey, file.ID.String())
	if err := s.cache.Set(ctx, existsCacheKey, true, cache.FileExistsTTL); err != nil {
		logger.WarnCtx(ctx, "Failed to cache file existence",
			"error", err,
			"file_id", file.ID,
			"cache_key", existsCacheKey,
		)
	}

	logger.InfoCtx(ctx, "File record created successfully",
		"file_id", createdFile.ID,
		"user_id", createdFile.UserID,
	)

	return createdFile, nil
}

func (s FileService) GetFile(ctx context.Context, fileID uuid.UUID) (File, error) {
	logger.DebugCtx(ctx, "Getting file", "file_id", fileID)

	// Try cache first
	cacheKey := fmt.Sprintf(cache.FileMetadataKey, fileID.String())
	var cachedFile File

	if err := s.cache.Get(ctx, cacheKey, &cachedFile); err == nil {
		logger.DebugCtx(ctx, "File retrieved from cache", "file_id", fileID)
		return cachedFile, nil
	}

	// Cache miss - get from database
	logger.DebugCtx(ctx, "File cache miss - querying database", "file_id", fileID)

	dbFile, err := s.queries.GetFile(ctx, fileID)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.WarnCtx(ctx, "File not found", "file_id", fileID)
			return File{}, fmt.Errorf("file not found")
		}
		logger.ErrorCtx(ctx, "Failed to get file from database", "error", err, "file_id", fileID)
		return File{}, fmt.Errorf("failed to get file: %w", err)
	}

	// Convert database file to our File model
	file := File{
		ID:               dbFile.ID,
		UserID:           dbFile.UserID,
		FileURI:          dbFile.FileUri,
		FileThumbnailURI: dbFile.FileThumbnailUri,
		CreatedAt:        dbFile.CreatedAt,
	}

	// Cache the file metadata
	if err := s.cache.Set(ctx, cacheKey, file, cache.FileMetadataTTL); err != nil {
		logger.WarnCtx(ctx, "Failed to cache file metadata",
			"error", err,
			"file_id", fileID,
			"cache_key", cacheKey,
		)
	}

	logger.InfoCtx(ctx, "File retrieved from database", "file_id", fileID)
	return file, nil
}

func (s FileService) GetFiles(ctx context.Context, fileID []uuid.UUID) ([]database.GetFilesByIDRow, error) {
	logger.DebugCtx(ctx, "Getting file", "file_id", fileID)
	file, err := s.queries.GetFilesByID(ctx, fileID)
	logger.Info("file id", file)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.WarnCtx(ctx, "File not found", "file_id", fileID)
			return nil, fmt.Errorf("file not found")
		}
		logger.ErrorCtx(ctx, "Failed to get file from database", "error", err, "file_id", fileID)
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	return file, nil
}


func (s FileService) DeleteFiles(ctx context.Context, fileID uuid.UUID, userID string) error {
	logger.InfoCtx(ctx, "Deleting file", "file_id", fileID, "user_id", userID)

	// First, verify the file exists and belongs to the user
	file, err := s.GetFile(ctx, fileID)
	if err != nil {
		return err // This will return "file not found" if file doesn't exist
	}

	// Check ownership
	if file.UserID.String() != userID {
		logger.WarnCtx(ctx, "Unauthorized file deletion attempt",
			"file_id", fileID,
			"requesting_user", userID,
			"file_owner", file.UserID.String(),
		)
		return fmt.Errorf("unauthorized")
	}

	// Delete from database
	err = s.queries.DeleteFile(ctx, fileID)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to delete file from database",
			"error", err,
			"file_id", fileID,
		)
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Remove from caches
	cacheKey := fmt.Sprintf(cache.FileMetadataKey, fileID.String())
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		logger.WarnCtx(ctx, "Failed to remove file from cache",
			"error", err,
			"file_id", fileID,
			"cache_key", cacheKey,
		)
	}

	// Remove from exists cache
	existsCacheKey := fmt.Sprintf(cache.FileExistsKey, fileID.String())
	if err := s.cache.Delete(ctx, existsCacheKey); err != nil {
		logger.WarnCtx(ctx, "Failed to remove file existence from cache",
			"error", err,
			"file_id", fileID,
			"cache_key", existsCacheKey,
		)
	}

	// Invalidate user's file list cache
	userCacheKey := fmt.Sprintf(cache.UserFileListKey, userID)
	if err := s.cache.Delete(ctx, userCacheKey); err != nil {
		logger.WarnCtx(ctx, "Failed to invalidate user file list cache",
			"error", err,
			"user_id", userID,
			"cache_key", userCacheKey,
		)
	}

	logger.InfoCtx(ctx, "File deleted successfully", "file_id", fileID, "user_id", userID)
	return nil
}

// GetUserFiles retrieves all files for a specific user (with caching)
func (s FileService) GetUserFiles(ctx context.Context, userID string) ([]File, error) {
	logger.DebugCtx(ctx, "Getting user files", "user_id", userID)

	cacheKey := fmt.Sprintf(cache.UserFileListKey, userID)
	var files []File

	err := s.cache.GetOrSet(ctx, cacheKey, &files, cache.FileListTTL, func() (interface{}, error) {
		logger.DebugCtx(ctx, "User files cache miss - querying database", "user_id", userID)

		userUUID, err := uuid.Parse(userID)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID: %w", err)
		}

		dbFiles, err := s.queries.GetFilesByUser(ctx, userUUID)
		if err != nil {
			logger.ErrorCtx(ctx, "Failed to get user files from database", "error", err, "user_id", userID)
			return nil, fmt.Errorf("failed to get user files: %w", err)
		}

		// Convert database files to our File model
		fileList := make([]File, len(dbFiles))
		for i, dbFile := range dbFiles {
			fileList[i] = File{
				ID:               dbFile.ID,
				UserID:           dbFile.UserID,
				FileURI:          dbFile.FileUri,
				FileThumbnailURI: dbFile.FileThumbnailUri,
				CreatedAt:        dbFile.CreatedAt,
			}
		}

		return fileList, nil
	})

	if err != nil {
		return nil, err
	}

	logger.InfoCtx(ctx, "User files retrieved successfully", "user_id", userID, "count", len(files))
	return files, nil
}

// FileExists checks if a file exists (with caching)
func (s FileService) FileExists(ctx context.Context, fileID uuid.UUID) (bool, error) {
	logger.DebugCtx(ctx, "Checking file existence", "file_id", fileID)

	// Try cache first
	cacheKey := fmt.Sprintf(cache.FileExistsKey, fileID.String())
	var exists bool

	if err := s.cache.Get(ctx, cacheKey, &exists); err == nil {
		logger.DebugCtx(ctx, "File existence retrieved from cache", "file_id", fileID, "exists", exists)
		return exists, nil
	}

	// Cache miss - check database
	logger.DebugCtx(ctx, "File existence cache miss - querying database", "file_id", fileID)

	_, err := s.queries.GetFile(ctx, fileID)
	if err != nil {
		if err == sql.ErrNoRows {
			exists = false
		} else {
			logger.ErrorCtx(ctx, "Failed to check file existence", "error", err, "file_id", fileID)
			return false, fmt.Errorf("failed to check file existence: %w", err)
		}
	} else {
		exists = true
	}

	// Cache the existence result
	if err := s.cache.Set(ctx, cacheKey, exists, cache.FileExistsTTL); err != nil {
		logger.WarnCtx(ctx, "Failed to cache file existence",
			"error", err,
			"file_id", fileID,
			"cache_key", cacheKey,
		)
	}

	logger.DebugCtx(ctx, "File existence checked", "file_id", fileID, "exists", exists)
	return exists, nil
}
