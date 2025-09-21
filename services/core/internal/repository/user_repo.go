package repository

import (
	"context"

	"github.com/teammachinist/tutuplapak/services/core/internal/database"

	"github.com/google/uuid"
)

type UserRepositoryInterface interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (database.Users, error)
	UpdateUser(ctx context.Context, args database.UpdateUserParams) (database.Users, error)
	UpdateUserPhone(ctx context.Context, userID uuid.UUID, phone string) error
	CheckPhoneExists(ctx context.Context, phone string) (bool, error)
	CheckEmailExists(ctx context.Context, email string) (bool, error)
	UpdateUserEmail(ctx context.Context, userID uuid.UUID, email string) error
	GetUserByAuthID(ctx context.Context, userAuthID uuid.UUID) (database.Users, error)
	CreateUserFromUserAuth(ctx context.Context, userID, userAuthID uuid.UUID, email, phone string) (database.Users, error)
}

type UserRepository struct {
	db database.Querier
}

func NewUserRepository(database database.Querier) UserRepositoryInterface {
	return &UserRepository{db: database}
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (database.Users, error) {
	rows, err := r.db.GetUserByID(ctx, userID)
	if err != nil {
		return database.Users{}, err
	}
	return rows, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, args database.UpdateUserParams) (database.Users, error) {
	rows, err := r.db.UpdateUser(ctx, args)

	if err != nil {
		return database.Users{}, err
	}
	return rows, nil

}

func (r *UserRepository) UpdateUserPhone(ctx context.Context, userID uuid.UUID, phone string) error {
	_, err := r.db.UpdateUserPhone(ctx, database.UpdateUserPhoneParams{
		ID:    userID,
		Phone: phone,
	})
	return err
}

func (r *UserRepository) CheckPhoneExists(ctx context.Context, phone string) (bool, error) {
	result, err := r.db.CheckPhoneExists(ctx, phone)
	if err != nil {
		return false, err
	}
	return result, nil
}

func (r *UserRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	result, err := r.db.CheckEmailExists(ctx, email)
	if err != nil {
		return false, err
	}
	return result, nil
}

func (r *UserRepository) UpdateUserEmail(ctx context.Context, userID uuid.UUID, email string) error {
	_, err := r.db.UpdateUserEmail(ctx, database.UpdateUserEmailParams{
		ID:    userID,
		Email: email,
	})
	return err
}

func (r *UserRepository) GetUserByAuthID(ctx context.Context, userAuthID uuid.UUID) (database.Users, error) {
	rows, err := r.db.GetUserByAuthID(ctx, userAuthID)
	if err != nil {
		return database.Users{}, err
	}
	return rows, nil
}

func (r *UserRepository) CreateUserFromUserAuth(ctx context.Context, userID, userAuthID uuid.UUID, email, phone string) (database.Users, error) {
	result, err := r.db.CreateUserFromUserAuth(ctx, database.CreateUserFromUserAuthParams{
		ID:                userID,
		UserAuthID:        userAuthID,
		FileID:            nil,
		Email:             email,
		Phone:             phone,
		BankAccountName:   "",
		BankAccountHolder: "",
		BankAccountNumber: "",
	})
	if err != nil {
		return database.Users{}, err
	}

	return result, nil
}
