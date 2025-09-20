package repository

import (
	"context"

	"github.com/teammachinist/tutuplapak/services/auth/internal/database"
	"github.com/teammachinist/tutuplapak/services/auth/internal/model"
)

type UserRepository struct {
	db *database.Queries
}

func NewUserRepository(db *database.Queries) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CheckPhoneExists(ctx context.Context, phone string) (bool, error) {
	result, err := r.db.CheckPhoneExists(ctx, phone)
	if err != nil {
		return false, err
	}
	return result, nil
}

func (r *UserRepository) GetUserByPhone(ctx context.Context, phone string) (*model.UserAuth, error) {
	user, err := r.db.GetUserAuthByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}

	return &model.UserAuth{
		ID:           user.ID,
		Phone:        user.Phone,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
	}, nil
}

func (r *UserRepository) CreateUserByPhone(ctx context.Context, phone, passwordHash string) (*model.UserAuth, error) {
	params := database.CreateUserByPhoneParams{
		Phone:        phone,
		PasswordHash: passwordHash,
	}

	user, err := r.db.CreateUserByPhone(ctx, params)
	if err != nil {
		return nil, err
	}

	return &model.UserAuth{
		ID:           user.ID,
		Phone:        user.Phone,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
	}, nil
}

func (r *UserRepository) CheckExistedUserAuthByEmail(ctx context.Context, email string) (bool, error) {
	result, err := r.db.CheckEmailExists(ctx, email)
	if err != nil {
		return false, err
	}
	return result, nil
}

func (r *UserRepository) GetUserAuthByEmail(ctx context.Context, email string) (*model.UserAuth, error) {
	user, err := r.db.GetUserAuthByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &model.UserAuth{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
	}, nil
}

func (r *UserRepository) RegisterWithEmail(ctx context.Context, email, passwordHash string) (*model.UserAuth, error) {
	params := database.CreateUserByEmailParams{
		Email:        email,
		PasswordHash: passwordHash,
	}

	user, err := r.db.CreateUserByEmail(ctx, params)
	if err != nil {
		return nil, err
	}

	return &model.UserAuth{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
	}, nil
}
