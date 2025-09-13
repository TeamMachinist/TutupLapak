package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/teammachinist/tutuplapak/internal"
	"github.com/teammachinist/tutuplapak/internal/database"

	"github.com/jackc/pgconn"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type UserService struct {
	userRepo   *UserRepository
	jwtService internal.JWTService
	db         *database.Queries
}

func NewUserService(userRepo *UserRepository, jwtService internal.JWTService, db *database.Queries) *UserService {
	return &UserService{
		userRepo:   userRepo,
		jwtService: jwtService,
		db:         db,
	}
}

func (s *UserService) LoginByPhone(ctx context.Context, phone, password string) (*AuthResponse, error) {
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
	if !CheckPassword(password, userAuth.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	fmt.Println("----------------")
	fmt.Printf("userAuth: %+v\n", userAuth)

	// Get user, hit db directly for development purpose
	user, err := s.db.GetUserByAuthID(ctx, userAuth.ID)

	fmt.Printf("user: %+v\n", user)
	fmt.Println("----------------")

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(user.ID.String())
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &AuthResponse{
		Email: userAuth.Email,
		Phone: userAuth.Phone,
		Token: token,
	}, nil
}

func (s *UserService) RegisterByPhone(ctx context.Context, phone, password string) (*AuthResponse, error) {
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

	// Create user, hit db directly for development purpose
	user, err := s.db.CreateUser(ctx, database.CreateUserParams{
		ID:         uuid.New(),
		UserAuthID: userAuth.ID,
		Phone:      &userAuth.Phone,
	})
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	// Generate token
	token, err := s.jwtService.GenerateToken(user.ID.String())
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &AuthResponse{
		Email: userAuth.Email,
		Phone: userAuth.Phone,
		Token: token,
	}, nil
}

func (s *UserService) LoginWithEmail(ctx context.Context, email, password string) (*AuthResponse, error) {
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
	if !CheckPassword(password, userAuth.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	fmt.Println("----------------")
	fmt.Printf("userAuth: %+v\n", userAuth)

	// Get user, hit db directly for development purpose
	user, err := s.db.GetUserByAuthID(ctx, userAuth.ID)

	fmt.Printf("user: %+v\n", user)
	fmt.Println("----------------")

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(user.ID.String())
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &AuthResponse{
		Email: userAuth.Email,
		Phone: userAuth.Phone,
		Token: token,
	}, nil
}

func (s *UserService) RegisterWithEmail(ctx context.Context, email, password string) (*AuthResponse, error) {
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
	passwordHash, err := HashPassword(password)
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

	// Create user, hit db directly for development purpose
	user, err := s.db.CreateUser(ctx, database.CreateUserParams{
		ID:         uuid.New(),
		UserAuthID: userAuth.ID,
		Email:      &userAuth.Email,
	})

	// Generate token
	token, err := s.jwtService.GenerateToken(user.ID.String())
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &AuthResponse{
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

func (s *UserService) LinkPhone(ctx context.Context, userID, phone string) (*LinkPhoneResponse, error) {
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

	var fileURI, fileThumbnailURI string
	if user.FileID.Valid {
		file, err := s.db.GetFileByIDAndUserID(ctx, database.GetFileByIDAndUserIDParams{
			ID:     user.FileID.Bytes,
			UserID: user.ID,
		})
		if err == nil {
			fileURI = file.FileUri
			fileThumbnailURI = file.FileThumbnailUri
		}
	}

	response := &LinkPhoneResponse{
		FileID:            "",
		FileURI:           fileURI,
		FileThumbnailURI:  fileThumbnailURI,
		BankAccountName:   "",
		BankAccountHolder: "",
		BankAccountNumber: "",
	}

	if user.Email != nil {
		response.Email = *user.Email
	}

	if user.Phone != nil {
		response.Phone = *user.Phone
	}

	if user.BankAccountName != nil {
		response.BankAccountName = *user.BankAccountName
	}
	if user.BankAccountHolder != nil {
		response.BankAccountHolder = *user.BankAccountHolder
	}
	if user.BankAccountNumber != nil {
		response.BankAccountNumber = *user.BankAccountNumber
	}

	if user.FileID.Valid {
		response.FileID = uuid.UUID(user.FileID.Bytes).String()
	}

	return response, nil
}
