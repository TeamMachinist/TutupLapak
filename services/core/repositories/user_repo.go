package repositories

import (
	"context"

	"github.com/teammachinist/tutuplapak/internal/database"

	"github.com/google/uuid"
)

type UserRepositoryInterface interface {
	GetUsersWithFileId(ctx context.Context, userID uuid.UUID) (database.GetUserWithFileIdRow, error)
	UpdateUser(ctx context.Context, args database.UpdateUserParams) (database.Users, error)
}

type UserRepository struct {
	db database.Querier
}

func NewUserRepository(database database.Querier) UserRepositoryInterface {
	return &UserRepository{db: database}
}

func (r *UserRepository) GetUsersWithFileId(ctx context.Context, userID uuid.UUID) (database.GetUserWithFileIdRow, error) {
	rows, err := r.db.GetUserWithFileId(ctx, userID)
	if err != nil {
		return database.GetUserWithFileIdRow{}, err
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
		Phone: &phone,
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
	result, err := r.db.CheckPhoneExists(ctx, &phone)
	if err != nil {
		return false, err
	}
	return result, nil
}
