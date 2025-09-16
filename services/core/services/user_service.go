package services

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/teammachinist/tutuplapak/services/core/clients"
	"github.com/teammachinist/tutuplapak/services/core/models"
	"github.com/teammachinist/tutuplapak/services/core/repositories"
)

type UserServiceInterface interface {
	LinkPhone(ctx context.Context, userID uuid.UUID, phone string) (*models.LinkPhoneResponse, error)
}

type UserService struct {
	userRepo   *repositories.UserRepository
	fileClient *clients.FileClient
}

func NewUserService(userRepo *repositories.UserRepository, fileClient *clients.FileClient) *UserService {
	return &UserService{
		userRepo:   userRepo,
		fileClient: fileClient,
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
	if user.FileID.Valid {
		file, err := s.fileClient.GetFileByID(ctx, uuid.UUID(user.FileID.Bytes), userID.String())
		if err == nil {
			fileURI = file.FileURI
			fileThumbnailURI = file.FileThumbnailURI
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

	if user.FileID.Valid {
		response.FileID = uuid.UUID(user.FileID.Bytes).String()
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
