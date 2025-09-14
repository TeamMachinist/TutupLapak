package main

import (
	"context"

	"github.com/teammachinist/tutuplapak/internal/database"
)

type UserRepository struct {
	db *database.Queries
}

func NewUserRepository(db *database.Queries) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CheckPhoneExists(ctx context.Context, phone string) (bool, error) {
	result, err := r.db.CheckPhoneExists(ctx, &phone)
	if err != nil {
		return false, err
	}
	return result, nil
}

func (r *UserRepository) GetUserByPhone(ctx context.Context, phone string) (*UserAuth, error) {
	user, err := r.db.GetUserAuthByPhone(ctx, &phone)
	if err != nil {
		return nil, err
	}

	return &UserAuth{
		ID:           user.ID,
		Phone:        *user.Phone,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt.Time,
	}, nil
}

func (r *UserRepository) CreateUserByPhone(ctx context.Context, phone, passwordHash string) (*UserAuth, error) {
	params := database.CreateUserAuthParams{
		Phone:        &phone,
		PasswordHash: passwordHash,
	}

	user, err := r.db.CreateUserAuth(ctx, params)
	if err != nil {
		return nil, err
	}

	return &UserAuth{
		ID:           user.ID,
		Phone:        *user.Phone,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt.Time,
	}, nil
}

func (r *UserRepository) CheckExistedUserAuthByEmail(ctx context.Context, email string) (bool, error) {
	result, err := r.db.CheckExistedUserAuthByEmail(ctx, &email)
	if err != nil {
		return false, err
	}
	return result, nil
}

func (r *UserRepository) GetUserAuthByEmail(ctx context.Context, email string) (*UserAuth, error) {
	user, err := r.db.GetUserAuthByEmail(ctx, &email)
	if err != nil {
		return nil, err
	}

	return &UserAuth{
		ID:           user.ID,
		Email:        *user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt.Time,
	}, nil
}

func (r *UserRepository) RegisterWithEmail(ctx context.Context, email, passwordHash string) (*UserAuth, error) {
	params := database.RegisterWithEmailParams{
		Email:        &email,
		PasswordHash: passwordHash,
	}

	user, err := r.db.RegisterWithEmail(ctx, params)
	if err != nil {
		return nil, err
	}

	return &UserAuth{
		ID:           user.ID,
		Email:        *user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt.Time,
	}, nil
}
