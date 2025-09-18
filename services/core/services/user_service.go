package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/teammachinist/tutuplapak/internal/cache"
	"github.com/teammachinist/tutuplapak/internal/database"
	"github.com/teammachinist/tutuplapak/services/core/clients"
	"github.com/teammachinist/tutuplapak/services/core/models"
	"github.com/teammachinist/tutuplapak/services/core/repositories"

	"github.com/google/uuid"
)

type UserServiceInterface interface {
	LinkPhone(ctx context.Context, userID uuid.UUID, phone string) (*models.LinkPhoneResponse, error)
	LinkEmail(ctx context.Context, userID uuid.UUID, email string) (*models.LinkEmailResponse, error)
	GetUserWithFileId(ctx context.Context, userID uuid.UUID) (models.UserResponse, error)
	UpdateUser(ctx context.Context, userId uuid.UUID, req models.UserRequest) (models.UserResponse, error)
}

type UserService struct {
	userRepo   repositories.UserRepositoryInterface
	fileClient clients.FileClientInterface
	cache      *cache.RedisCache
}

func NewUserService(
	userRepo repositories.UserRepositoryInterface,
	fileClient clients.FileClientInterface,
	cache *cache.RedisCache,
) UserServiceInterface {
	return &UserService{
		userRepo:   userRepo,
		fileClient: fileClient,
		cache:      cache,
	}
}

func (s *UserService) LinkPhone(ctx context.Context, userID uuid.UUID, phone string) (*models.LinkPhoneResponse, error) {
	if err := s.validatePhone(phone); err != nil {
		return nil, err
	}

	exists, err := s.userRepo.CheckPhoneExists(ctx, phone)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("phone is taken")
	}

	err = s.userRepo.UpdateUserPhone(ctx, userID, phone)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get file data if fileID exists
	var fileURI, fileThumbnailURI string
	if user.FileID != nil {
		if err := uuid.Validate(user.FileID.String()); err == nil {
			file, err := s.fileClient.GetFileByID(ctx, *user.FileID)
			if err == nil {
				fileURI = file.FileURI
				fileThumbnailURI = file.FileThumbnailURI
			}
		}
	}

	// Build response
	response := &models.LinkPhoneResponse{
		FileID:            "",
		FileURI:           fileURI,
		FileThumbnailURI:  fileThumbnailURI,
		BankAccountName:   "",
		BankAccountHolder: "",
		BankAccountNumber: "",
	}

	if user.Email != "" {
		response.Email = user.Email
	}

	if user.Phone != "" {
		response.Phone = user.Phone
	}

	if user.BankAccountName != "" {
		response.BankAccountName = user.BankAccountName
	}
	if user.BankAccountHolder != "" {
		response.BankAccountHolder = user.BankAccountHolder
	}
	if user.BankAccountNumber != "" {
		response.BankAccountNumber = user.BankAccountNumber
	}

	if user.FileID != nil {
		if err := uuid.Validate(user.FileID.String()); err == nil {
			response.FileID = user.FileID.String()
		}
	}

	return response, nil
}

func (s *UserService) LinkEmail(ctx context.Context, userID uuid.UUID, email string) (*models.LinkEmailResponse, error) {
	if err := s.validateEmail(email); err != nil {
		return nil, err
	}

	exists, err := s.userRepo.CheckEmailExists(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email is taken")
	}

	err = s.userRepo.UpdateUserEmail(ctx, userID, email)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get file data if fileID exists
	var fileURI, fileThumbnailURI string
	if user.FileID != nil {
		if err := uuid.Validate(user.FileID.String()); err == nil {
			file, err := s.fileClient.GetFileByID(ctx, *user.FileID)
			if err == nil {
				fileURI = file.FileURI
				fileThumbnailURI = file.FileThumbnailURI
			}
		}
	}
	fmt.Printf("error apa nih: %s", err)

	// Build response
	response := &models.LinkEmailResponse{
		FileID:            "",
		FileURI:           fileURI,
		FileThumbnailURI:  fileThumbnailURI,
		BankAccountName:   "",
		BankAccountHolder: "",
		BankAccountNumber: "",
	}

	if user.Email != "" {
		response.Email = user.Email
	}

	if user.Phone != "" {
		response.Phone = user.Phone
	}

	if user.BankAccountName != "" {
		response.BankAccountName = user.BankAccountName
	}
	if user.BankAccountHolder != "" {
		response.BankAccountHolder = user.BankAccountHolder
	}
	if user.BankAccountNumber != "" {
		response.BankAccountNumber = user.BankAccountNumber
	}

	if user.FileID != nil {
		if err := uuid.Validate(user.FileID.String()); err == nil {
			response.FileID = user.FileID.String()
		}
	}

	return response, nil
}

func (s *UserService) validatePhone(phone string) error {
	if phone == "" {
		return errors.New("phone is required")
	}

	if !strings.HasPrefix(phone, "+") {
		return errors.New("phone must start with international calling code prefix '+'")
	}

	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	if !phoneRegex.MatchString(phone) {
		return errors.New("invalid phone number format")
	}

	return nil
}

func (s *UserService) GetUserWithFileId(ctx context.Context, userID uuid.UUID) (models.UserResponse, error) {
	fmt.Printf("masuk service: %s", userID.String())
	var userFile models.UserResponse
	if err := s.cache.Get(ctx, fmt.Sprintf(cache.UserFileListKey, userID.String()), &userFile); err == nil {
		return models.UserResponse{
			Email:             userFile.Email,
			Phone:             userFile.Phone,
			FileID:            userFile.FileID,
			URI:               userFile.URI,
			ThumbnailURI:      userFile.ThumbnailURI,
			BankAccountName:   userFile.BankAccountName,
			BankAccountHolder: userFile.BankAccountHolder,
			BankAccountNumber: userFile.BankAccountNumber,
		}, nil
	}

	// TODO: Pisah query
	rows, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return models.UserResponse{}, errors.New("user does not exist")
	}

	var file *clients.FileMetadataResponse
	if err := s.cache.Get(ctx, cache.FileMetadataKey, file); err != nil {
		clientFile, err := s.fileClient.GetFileByID(ctx, *rows.FileID)
		if err != nil {
			return models.UserResponse{}, err
		}
		file = clientFile
	}

	resp := models.UserResponse{
		Email:             rows.Email,
		Phone:             rows.Phone,
		FileID:            rows.FileID.String(),
		URI:               file.FileURI,
		ThumbnailURI:      file.FileThumbnailURI,
		BankAccountName:   rows.BankAccountName,
		BankAccountHolder: rows.BankAccountHolder,
		BankAccountNumber: rows.BankAccountNumber,
	}
	if err = s.cache.Set(ctx, fmt.Sprintf(cache.UserFileListKey, userID.String()), resp, cache.FileListTTL); err != nil {
		fmt.Printf("[UserFileList] Failed to set cache: %v", err)
	}
	return resp, nil
}

func (s *UserService) UpdateUser(ctx context.Context,
	userID uuid.UUID,
	req models.UserRequest,
) (models.UserResponse, error) {
	var userFile models.UserResponse

	fileUUID := stringPtrToUUID(req.FileID)

	fileID := ""
	fileURI := ""
	fileThumbnailURI := ""

	if err := s.cache.Get(ctx, fmt.Sprintf(cache.UserFileListKey, userID.String()), &userFile); err == nil {
		if userFile.BankAccountName == req.BankAccountName &&
			userFile.BankAccountHolder == req.BankAccountHolder &&
			userFile.BankAccountNumber == req.BankAccountNumber {
			if req.FileID != nil {
				if userFile.FileID == *req.FileID {
					return userFile, nil
				}
			} else {
				return userFile, nil
			}
		}
	}
	// if err := s.cache.Get(ctx, cacheKey, &cachedFile); err == nil {
	// 	logger.DebugCtx(ctx, "File retrieved from cache", "file_id", fileID)
	// 	return cachedFile, nil
	// }

	// check exist & ownership
	if req.FileID != nil && *req.FileID != "" && *req.FileID != uuid.Nil.String() {
		// Check file exist
		var file *clients.FileMetadataResponse
		fmt.Printf("masuk sini")
		if err := s.cache.Get(ctx, cache.FileMetadataKey, file); err != nil {
			fmt.Printf("masuk error cache")
			clientFile, err := s.fileClient.GetFileByID(ctx, *fileUUID)
			if err != nil {
				fmt.Printf("masuk error client: %s", err)
				return models.UserResponse{}, err
			}
			file = clientFile
		}

		// Check Ownership
		fmt.Printf("userId.String: %s", userID.String())
		fmt.Printf("file.UserID: %s", file.UserID)
		if userID.String() != file.UserID {
			return models.UserResponse{}, errors.New("unauthorized: you don't own this file")
		}

		fileID = file.ID.String()
		fileURI = file.FileURI
		fileThumbnailURI = file.FileThumbnailURI

		rows, err := s.userRepo.UpdateUser(ctx, database.UpdateUserParams{
			ID:                userID,
			FileID:            fileUUID,
			BankAccountName:   req.BankAccountName,
			BankAccountHolder: req.BankAccountHolder,
			BankAccountNumber: req.BankAccountNumber,
		})

		if err != nil {
			return models.UserResponse{}, err
		}

		resp := models.UserResponse{
			Email:             rows.Email,
			Phone:             rows.Phone,
			FileID:            fileID,
			URI:               fileURI,
			ThumbnailURI:      fileThumbnailURI,
			BankAccountName:   rows.BankAccountName,
			BankAccountHolder: rows.BankAccountHolder,
			BankAccountNumber: rows.BankAccountNumber,
		}

		if err = s.cache.Set(ctx, fmt.Sprintf(cache.UserFileListKey, userID.String()), resp, cache.FileListTTL); err != nil {
			fmt.Printf("[UserFileList] Failed to set cache: %v", err)
		}

		return resp, nil
	}

	rows, err := s.userRepo.UpdateUser(ctx, database.UpdateUserParams{
		ID:                userID,
		BankAccountName:   req.BankAccountName,
		BankAccountHolder: req.BankAccountHolder,
		BankAccountNumber: req.BankAccountNumber,
	})

	if err != nil {
		return models.UserResponse{}, err
	}

	resp := models.UserResponse{
		Email:             rows.Email,
		Phone:             rows.Phone,
		FileID:            fileID,
		URI:               fileURI,
		ThumbnailURI:      fileThumbnailURI,
		BankAccountName:   rows.BankAccountName,
		BankAccountHolder: rows.BankAccountHolder,
		BankAccountNumber: rows.BankAccountNumber,
	}

	if err = s.cache.Set(ctx, fmt.Sprintf(cache.UserFileListKey, userID.String()), resp, cache.FileListTTL); err != nil {
		fmt.Printf("[UserFileList] Failed to set cache: %v", err)
	}

	return resp, nil
}

// Helper: safely convert *string to *uuid.UUID
func stringPtrToUUID(s *string) *uuid.UUID {
	if s == nil || *s == "" {
		return nil
	}
	parsed, err := uuid.Parse(*s)
	if err != nil {
		return nil
	}
	return &parsed
}

func (s *UserService) validateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	// Basic email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}
