package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/teammachinist/tutuplapak/services/core/internal/cache"
	"github.com/teammachinist/tutuplapak/services/core/internal/clients"
	"github.com/teammachinist/tutuplapak/services/core/internal/database"
	"github.com/teammachinist/tutuplapak/services/core/internal/model"
	"github.com/teammachinist/tutuplapak/services/core/internal/repository"

	"github.com/google/uuid"
)

type UserServiceInterface interface {
	LinkPhone(ctx context.Context, userID uuid.UUID, phone string) (*model.LinkPhoneResponse, error)
	LinkEmail(ctx context.Context, userID uuid.UUID, email string) (*model.LinkEmailResponse, error)
	GetUserWithFileId(ctx context.Context, userID uuid.UUID) (model.UserResponse, error)
	UpdateUser(ctx context.Context, userId uuid.UUID, req model.UserRequest) (model.UserResponse, error)
}

type UserService struct {
	userRepo   repository.UserRepositoryInterface
	fileClient clients.FileClientInterface
	cache      *cache.RedisCache
}

func NewUserService(
	userRepo repository.UserRepositoryInterface,
	fileClient clients.FileClientInterface,
	cache *cache.RedisCache,
) UserServiceInterface {
	return &UserService{
		userRepo:   userRepo,
		fileClient: fileClient,
		cache:      cache,
	}
}

func (s *UserService) LinkPhone(ctx context.Context, userID uuid.UUID, phone string) (*model.LinkPhoneResponse, error) {
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

	// TODO: Update phone in users_auth via authz package or http_client

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
	response := &model.LinkPhoneResponse{
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

func (s *UserService) LinkEmail(ctx context.Context, userID uuid.UUID, email string) (*model.LinkEmailResponse, error) {
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

	// TODO: Update email in users_auth via authz package or http_client

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
	response := &model.LinkEmailResponse{
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

func (s *UserService) GetUserWithFileId(ctx context.Context, userID uuid.UUID) (model.UserResponse, error) {
	var userFile model.UserResponse
	if err := s.cache.Get(ctx, fmt.Sprintf(cache.UserFileListKey, userID.String()), &userFile); err == nil {
		return model.UserResponse{
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

	rows, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return model.UserResponse{}, errors.New("user does not exist")
	}

	fileID := ""
	fileURI := ""
	fileThumbnailURI := ""

	if rows.FileID != nil {
		var file *clients.FileMetadataResponse
		if err := s.cache.Get(ctx, cache.FileMetadataKey, file); err != nil {
			clientFile, err := s.fileClient.GetFileByID(ctx, *rows.FileID)
			if err != nil {
				return model.UserResponse{}, err
			}
			file = clientFile
		}
		fileID = file.ID.String()
		fileURI = file.FileURI
		fileThumbnailURI = file.FileThumbnailURI
	}

	resp := model.UserResponse{
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

func (s *UserService) UpdateUser(ctx context.Context,
	userID uuid.UUID,
	req model.UserRequest,
) (model.UserResponse, error) {
	var userFile model.UserResponse

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

	// check exist & ownership
	if req.FileID != nil && *req.FileID != "" && *req.FileID != uuid.Nil.String() {
		// Check file exist
		var file *clients.FileMetadataResponse
		if err := s.cache.Get(ctx, cache.FileMetadataKey, file); err != nil {
			clientFile, err := s.fileClient.GetFileByID(ctx, *fileUUID)
			if err != nil {
				return model.UserResponse{}, err
			}
			file = clientFile
		}

		// Check Ownership
		if userID.String() != file.UserID {
			return model.UserResponse{}, errors.New("unauthorized: you don't own this file")
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
			return model.UserResponse{}, err
		}

		resp := model.UserResponse{
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
		return model.UserResponse{}, err
	}

	resp := model.UserResponse{
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
