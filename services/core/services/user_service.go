package services

import (
	"context"
	"fmt"
	// "errors"
	// "regexp"
	// "strings"

	"github.com/teammachinist/tutuplapak/internal/database"
	"github.com/teammachinist/tutuplapak/services/core/clients"
	"github.com/teammachinist/tutuplapak/services/core/models"
	"github.com/teammachinist/tutuplapak/services/core/repositories"

	"github.com/google/uuid"
)

type UserServiceInterface interface {
	// LinkPhone(ctx context.Context, userID uuid.UUID, phone string) (*models.LinkPhoneResponse, error)
	GetUserWithFileId(ctx context.Context, userID uuid.UUID) (models.UserResponse, error)
	UpdateUser(ctx context.Context, userId uuid.UUID, req models.UserRequest) (models.UserResponse, error)
}

type UserService struct {
	userRepo   repositories.UserRepositoryInterface
	fileClient clients.FileClientInterface
}

func NewUserService(userRepo repositories.UserRepositoryInterface, fileClient clients.FileClientInterface) UserServiceInterface {
	return &UserService{
		userRepo:   userRepo,
		fileClient: fileClient,
	}
}

// type PurchaseServiceInterface interface {
// 	CreatePurchase(ctx context.Context, req models.PurchaseRequest) (models.PurchaseResponse, error)
// }

// type PurchaseService struct {
// 	purchaseRepo repositories.PurchaseRepositoryInterface
// 	fileClient   clients.FileClientInterface
// }

// func NewPurchaseService(
// 	purchaseRepo repositories.PurchaseRepositoryInterface,
// 	fileClient clients.FileClientInterface,
// ) PurchaseServiceInterface {
// 	return &PurchaseService{
// 		purchaseRepo: purchaseRepo,
// 		fileClient:   fileClient,
// 	}
// }

// func (s *UserService) LinkPhone(ctx context.Context, userID uuid.UUID, phone string) (*models.LinkPhoneResponse, error) {
// 	if err := s.validatePhone(phone); err != nil {
// 		return nil, err
// 	}

// 	exists, err := s.userRepo.CheckPhoneExists(ctx, phone)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if exists {
// 		return nil, errors.New("phone is taken")
// 	}

// 	err = s.userRepo.UpdateUserPhone(ctx, userID, phone)
// 	if err != nil {
// 		return nil, err
// 	}

// 	user, err := s.userRepo.GetUserByID(ctx, userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Get file data if fileID exists
// 	var fileURI, fileThumbnailURI string
// 	if user.FileID.Valid {
// 		file, err := s.fileClient.GetFileByID(ctx, uuid.UUID(user.FileID.Bytes), userID.String())
// 		if err == nil {
// 			fileURI = file.FileURI
// 			fileThumbnailURI = file.FileThumbnailURI
// 		}
// 	}

// 	// Build response
// 	response := &models.LinkPhoneResponse{
// 		FileID:            "",
// 		FileURI:           fileURI,
// 		FileThumbnailURI:  fileThumbnailURI,
// 		BankAccountName:   "",
// 		BankAccountHolder: "",
// 		BankAccountNumber: "",
// 	}

// 	if user.Email != nil {
// 		response.Email = *user.Email
// 	}

// 	if user.Phone != nil {
// 		response.Phone = *user.Phone
// 	}

// 	if user.BankAccountName != nil {
// 		response.BankAccountName = *user.BankAccountName
// 	}
// 	if user.BankAccountHolder != nil {
// 		response.BankAccountHolder = *user.BankAccountHolder
// 	}
// 	if user.BankAccountNumber != nil {
// 		response.BankAccountNumber = *user.BankAccountNumber
// 	}

// 	if user.FileID.Valid {
// 		response.FileID = uuid.UUID(user.FileID.Bytes).String()
// 	}

// 	return response, nil
// }

// func (s *UserService) validatePhone(phone string) error {
// 	if phone == "" {
// 		return errors.New("phone is required")
// 	}

// 	if !strings.HasPrefix(phone, "+") {
// 		return errors.New("phone must start with international calling code prefix '+'")
// 	}

// 	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
// 	if !phoneRegex.MatchString(phone) {
// 		return errors.New("invalid phone number format")
// 	}

// 	return nil
// }

func (s *UserService) GetUserWithFileId(ctx context.Context, userID uuid.UUID) (models.UserResponse, error) {
	fmt.Printf("masuk service: %s", userID.String())
	rows, err := s.userRepo.GetUsersWithFileId(ctx, userID)
	if err != nil {
		return models.UserResponse{}, err
	}
	resp := models.UserResponse{
		Email:             rows.Email,
		Phone:             rows.Phone,
		FileID:            rows.ID.String(),
		URI:               rows.FileUri,
		ThumbnailURI:      rows.FileThumbnailUri,
		BankAccountName:   rows.BankAccountName,
		BankAccountHolder: rows.BankAccountHolder,
		BankAccountNumber: rows.BankAccountNumber,
	}
	return resp, nil
}

func (s *UserService) UpdateUser(ctx context.Context,
	userId uuid.UUID,
	req models.UserRequest,
) (models.UserResponse, error) {

	fileID := stringPtrToUUID(req.FileID)

	rows, err := s.userRepo.UpdateUser(ctx, database.UpdateUserParams{
		ID:                userId,
		FileID:            fileID,
		BankAccountName:   req.BankAccountName,
		BankAccountHolder: req.BankAccountHolder,
		BankAccountNumber: req.BankAccountNumber,
	})

	if err != nil {
		return models.UserResponse{}, err
	}

	newFileID := ""
	fileURI := ""
	fileThumbnailURI := ""
	if rows.FileID != uuid.Nil {
		fmt.Print("masuk uuid not Nil")
		fmt.Printf("userID: %s", userId.String())
		fmt.Printf("fileID: %s", rows.FileID.String())
		file, err := s.fileClient.GetFileByID(ctx, rows.FileID, userId.String())
		if err == nil {
			// newFileID = file.ID.String()
			fileURI = file.FileURI
			fileThumbnailURI = file.FileThumbnailURI
		}
		// fmt.Printf("fileID: %s", file.ID.String())
		// fmt.Printf("fileID2: %s", newFileID)
		fmt.Printf("erros: %s", err)
	}

	resp := models.UserResponse{
		Email:             rows.Email,
		Phone:             rows.Phone,
		FileID:            newFileID,
		URI:               fileURI,
		ThumbnailURI:      fileThumbnailURI,
		BankAccountName:   rows.BankAccountName,
		BankAccountHolder: rows.BankAccountHolder,
		BankAccountNumber: rows.BankAccountNumber,
	}
	return resp, nil
}

// Helper: safely convert *string to *uuid.UUID
func stringPtrToUUID(s *string) uuid.UUID {
	if s == nil || *s == "" {
		return uuid.Nil
	}
	parsed, err := uuid.Parse(*s)
	if err != nil {
		return uuid.Nil
	}
	return parsed
}
