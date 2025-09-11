package main

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/teammachinist/tutuplapak/internal"

	"github.com/jackc/pgconn"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type UserService struct {
	userRepo   *UserRepository
	jwtService internal.JWTService
}

func NewUserService(userRepo *UserRepository, jwtService internal.JWTService) *UserService {
	return &UserService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (s *UserService) LoginByPhone(ctx context.Context, phone, password string) (*LoginResponse, error) {
	// Validate phone format
	if err := s.validatePhone(phone); err != nil {
		return nil, err
	}

	// Validate password
	if err := s.validatePassword(password); err != nil {
		return nil, err
	}

	// Get user by phone
	user, err := s.userRepo.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, errors.New("phone not found")
	}

	// Check password
	if !CheckPassword(password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &LoginResponse{
		Phone: user.Phone,
		Token: token,
	}, nil
}

func (s *UserService) RegisterByPhone(ctx context.Context, phone, password string) (*LoginResponse, error) {
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
	passwordHash, err := HashPassword(password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user
	user, err := s.userRepo.CreateUserByPhone(ctx, phone, passwordHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, errors.New("phone number already exists")
			}
		}
		return nil, err
	}

	// Generate token
	token, err := s.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &LoginResponse{
		Phone: user.Phone,
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
