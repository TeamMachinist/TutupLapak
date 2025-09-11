package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AuthRepository struct {
	db *db.Queries
}

func NewAuthRepository(db *db.Queries) *AuthRepository {
	return &AuthRepository{q: q}
}

func (r *AuthRepository) CheckExistedUserByEmail(ctx context.Context, email strings) (bool, error) {
	exist, err := r.db.CheckExistedUserByEmail(ctx, email)

	if err != nil {
		return false, err
	}

	if exist == nil {
		return false, nil
	}

	return true, nil
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email strings) (UsersAuth, error) {
	result, err := r.db.GetUserByEmail(ctx, email)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *AuthRepository) RegisterWithEmail(ctx context.Context, user UsersAuth) error {
	result, err := r.db.RegisterWithEmail(ctx, user.ID, user.Email, user.Phone, user.HashedPassword, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}