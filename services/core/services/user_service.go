package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/teammachinist/tutuplapak/internal/database"
	"github.com/teammachinist/tutuplapak/services/core/models"
	"github.com/teammachinist/tutuplapak/services/core/repositories"
)

type UserServiceInterface interface {
	GetUserWithFileId(ctx context.Context, userID uuid.UUID) (database.GetUserWithFileIdRow, error)
	UpdateUser(ctx context.Context, userId uuid.UUID, req models.UserRequest) (database.Users, error)
}

type UserService struct {
	userRepo repositories.UserRepositoryInterface
}

func NewUserService(userRepo repositories.UserRepositoryInterface) UserServiceInterface {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetUserWithFileId(ctx context.Context, userID uuid.UUID) (database.GetUserWithFileIdRow, error) {
	fmt.Printf("masuk service: %s", userID.String())
	rows, err := s.userRepo.GetUsersWithFileId(ctx, userID)
	if err != nil {
		return database.GetUserWithFileIdRow{}, err
	}
	return rows, nil
}

func (s *UserService) UpdateUser(ctx context.Context,
	userId uuid.UUID,
	req models.UserRequest,
) (database.Users, error) {

	rows, err := s.userRepo.UpdateUser(ctx, database.UpdateUserParams{
		ID:                userId,
		// FileID:            req.FileID,
		BankAccountName:   &req.BankAccountName,
		BankAccountHolder: &req.BankAccountHolder,
		BankAccountNumber: &req.BankAccountNumber,
	})

	if err != nil {
		return database.Users{}, err
	}
	return rows, nil

}
