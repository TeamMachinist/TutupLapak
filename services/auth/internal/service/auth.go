package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/teammachinist/tutuplapak/internal"
	"github.com/teammachinist/tutuplapak/services/auth/internal/database"
	"github.com/teammachinist/tutuplapak/services/auth/internal/model"
	"github.com/teammachinist/tutuplapak/services/auth/internal/repository"

	"github.com/jackc/pgconn"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type UserService struct {
	userRepo *repository.UserRepository
	// TODO: Use authz package
	jwtService internal.JWTService
	db         *database.Queries
}

func NewUserService(userRepo *repository.UserRepository, jwtService internal.JWTService, db *database.Queries) *UserService {
	return &UserService{
		userRepo:   userRepo,
		jwtService: jwtService,
		db:         db,
	}
}

func (s *UserService) LoginByPhone(ctx context.Context, phone, password string) (*model.AuthResponse, error) {
	// Validate phone format
	if err := s.validatePhone(phone); err != nil {
		return nil, err
	}

	// Validate password
	if err := s.validatePassword(password); err != nil {
		return nil, err
	}

	// Validate user auth by phone
	userAuth, err := s.userRepo.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, errors.New("phone not found")
	}

	// Check password
	if !model.CheckPassword(password, userAuth.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	// // TODO: Opt 1 - Call user service to get users(id)
	// // TODO: Opt 2 - Put users_auth(id) to token
	// user, err := s.db.GetUserAuthByID(ctx, userAuth.ID)

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(userAuth.ID.String())
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &model.AuthResponse{
		Email: userAuth.Email,
		Phone: userAuth.Phone,
		Token: token,
	}, nil
}

func (s *UserService) RegisterByPhone(ctx context.Context, phone, password string) (*model.AuthResponse, error) {
	// Validate inputs
	if err := s.validatePhone(phone); err != nil {
		return nil, err
	}
	if err := s.validatePassword(password); err != nil {
		return nil, err
	}

	// Check if phone already exists
	exists, err := s.userRepo.CheckPhoneExists(ctx, phone)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("phone number already exists")
	}

	// Hash password
	passwordHash, err := model.HashPassword(password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user auth
	userAuth, err := s.userRepo.CreateUserByPhone(ctx, phone, passwordHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, errors.New("phone number already exists")
			}
		}
		return nil, err
	}

	// // TODO: Call user service to create new user
	// user, err := s.db.CreateUserByPhone(ctx, database.CreateUserByPhoneParams{
	// 	ID:         uuid.New(),
	// 	UserAuthID: userAuth.ID,
	// 	Phone:      userAuth.Phone,
	// })
	// if err != nil {
	// 	return nil, errors.New("failed to create user")
	// }

	// Generate token
	token, err := s.jwtService.GenerateToken(userAuth.ID.String())
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &model.AuthResponse{
		Email: userAuth.Email,
		Phone: userAuth.Phone,
		Token: token,
	}, nil
}

func (s *UserService) LoginWithEmail(ctx context.Context, email, password string) (*model.AuthResponse, error) {
	// Validate email format
	if err := s.validateEmail(email); err != nil {
		return nil, err
	}

	// Validate password
	if err := s.validatePassword(password); err != nil {
		return nil, err
	}

	// Validate user auth by email
	userAuth, err := s.userRepo.GetUserAuthByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("email not found")
	}

	// Check password
	if !model.CheckPassword(password, userAuth.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	// // TODO: Opt 1 - Call user service to get users(id)
	// // TODO: Opt 2 - Put users_auth(id) to token
	// user, err := s.db.GetUserAuthByID(ctx, userAuth.ID)

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(userAuth.ID.String())
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &model.AuthResponse{
		Email: userAuth.Email,
		Phone: userAuth.Phone,
		Token: token,
	}, nil
}

func (s *UserService) RegisterWithEmail(ctx context.Context, email, password string) (*model.AuthResponse, error) {
	// Validate email format
	if err := s.validateEmail(email); err != nil {
		return nil, err
	}

	// Validate password
	if err := s.validatePassword(password); err != nil {
		return nil, err
	}

	// Check if email already exists
	exists, err := s.userRepo.CheckExistedUserAuthByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email address already exists")
	}

	// Hash password
	passwordHash, err := model.HashPassword(password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user auth
	userAuth, err := s.userRepo.RegisterWithEmail(ctx, email, passwordHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, errors.New("email address already exists")
			}
		}
		return nil, err
	}

	// // TODO: Call user service to create new user
	// user, err := s.db.CreateUserByEmail(ctx, database.CreateUserByEmailParams{
	// 	ID:         uuid.New(),
	// 	UserAuthID: userAuth.ID,
	// 	Email:      userAuth.Email,
	// })
	// if err != nil {
	// 	return nil, errors.New("failed to create user")
	// }

	// Generate token
	token, err := s.jwtService.GenerateToken(userAuth.ID.String())
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &model.AuthResponse{
		Email: userAuth.Email,
		Phone: userAuth.Phone,
		Token: token,
	}, nil
}

func (s *UserService) validatePhone(phone string) error {
	if phone == "" {
		return errors.New("phone is required")
	}

	// Check if phone starts with "+"
	if !strings.HasPrefix(phone, "+") {
		return errors.New("phone must start with international calling code prefix '+'")
	}

	// Validate phone format (international format)
	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	if !phoneRegex.MatchString(phone) {
		return errors.New("invalid phone number format")
	}

	return nil
}

func (s *UserService) validateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	// Validate email format (international format)
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email address format")
	}

	return nil
}

func (s *UserService) validatePassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}

	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if len(password) > 32 {
		return errors.New("password must be at most 32 characters long")
	}

	return nil
}
