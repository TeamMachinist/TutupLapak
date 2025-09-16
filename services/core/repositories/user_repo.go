package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/teammachinist/tutuplapak/internal/database"
)

type UserRepository struct {
	db *database.Queries
}

func NewUserRepository(db *database.Queries) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) UpdateUserPhone(ctx context.Context, userID uuid.UUID, phone string) error {
	_, err := r.db.UpdateUserPhone(ctx, database.UpdateUserPhoneParams{
		ID:    userID,
		Phone: phone,
	})
	return err
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*database.Users, error) {
	user, err := r.db.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) CheckPhoneExists(ctx context.Context, phone string) (bool, error) {
	result, err := r.db.CheckPhoneExists(ctx, phone)
	if err != nil {
		return false, err
	}
	return result, nil
}
