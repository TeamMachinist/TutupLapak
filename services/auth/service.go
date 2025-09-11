package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"https://github.com/TeamMachinist/TutupLapak/internal"
	"https://github.com/TeamMachinist/TutupLapak/services/auth"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo   *repository.UserRepository
	cache      *cache.Redis
	userUtils  utils.PasswordHasher
	jwtService JwtService
}

func NewUserService(userRepo *repository.UserRepository, cache *cache.Redis, jwt JwtService) *UserService {
	return &UserService{
		userRepo:   userRepo,
		cache:      cache,
		userUtils:  utils.NewPasswordHasher(),
		jwtService: jwt,
	}
}

func (s *UserService) RegisterNewUser(ctx context.Context, payload model.User) (model.AuthResponse, error) {
	// Business Logic
	// Check if email exists Import get email function

	hashedPassword, err := s.userUtils.EncryptPassword(payload.Password)
	if err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to encrypt password: %v", err)
	}
	payload.Password = hashedPassword

	// Payload exists here
	createdUser, err := s.userRepo.RegisterNewUser(ctx, payload)
	if err != nil {
		if ctx.Err() != nil {
			return model.AuthResponse{}, fmt.Errorf("context error: %v", ctx.Err())
		}
		return model.AuthResponse{}, fmt.Errorf("failed to create user: %v", err)
	}

	token, err := s.jwtService.GenerateToken(&createdUser)
	if err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to generate token: %v", err)
	}

	return model.AuthResponse{Email: createdUser.Email, Token: token}, nil
}

func (s *UserService) Login(ctx context.Context, payload model.User) (model.AuthResponse, error) {
	if payload.Email == "" {
		return model.AuthResponse{}, fmt.Errorf("email is required")
	}
	if payload.Password == "" {
		return model.AuthResponse{}, fmt.Errorf("password is required")
	}

	user, err := s.userRepo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		return model.AuthResponse{}, fmt.Errorf("user not found")
	}

	if err := s.userUtils.ComparePasswordHash(user.Password, payload.Password); err != nil {
		return model.AuthResponse{}, fmt.Errorf("invalid credentials")
	}

	token, err := s.jwtService.GenerateToken(&user)
	if err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to generate token: %v", err)
	}

	return model.AuthResponse{Email: payload.Email, Token: token}, nil
}

type UserService struct {
	userRepo   *repository.UserRepository
	cache      *cache.Redis
	userUtils  utils.PasswordHasher
	jwtService JwtService
}

func NewUserService(userRepo *repository.UserRepository, cache *cache.Redis, jwt JwtService) *UserService {
	return &UserService{
		userRepo:   userRepo,
		cache:      cache,
		userUtils:  utils.NewPasswordHasher(), // Consider making this a singleton if expensive to create
		jwtService: jwt,
	}
}

// RegisterNewUser optimized for performance with value passing
// Uses stack allocation for better cache locality and reduced GC pressure
func (s *UserService) RegisterNewUser(ctx context.Context, payload model.User) (model.AuthResponse, error) {
	// Early context check to avoid unnecessary work
	select {
	case <-ctx.Done():
		return model.AuthResponse{}, ctx.Err()
	default:
	}

	// Check if email exists using cache first, then database
	exists, err := s.checkUserExists(ctx, payload.Email)
	if err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return model.AuthResponse{}, ErrUserExists
	}

	// Hash password - this is CPU intensive, do it early
	hashedPassword, err := s.userUtils.EncryptPassword(payload.Password)
	if err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to encrypt password: %w", err)
	}

	// Create a copy with hashed password (avoiding mutation of original)
	userToCreate := payload
	userToCreate.Password = hashedPassword

	// Another context check before expensive DB operation
	select {
	case <-ctx.Done():
		return model.AuthResponse{}, ctx.Err()
	default:
	}

	// Create user in database
	createdUser, err := s.userRepo.RegisterNewUser(ctx, userToCreate)
	if err != nil {
		// Check if context was cancelled during DB operation
		if ctx.Err() != nil {
			return model.AuthResponse{}, ctx.Err()
		}
		return model.AuthResponse{}, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(&createdUser)
	if err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to generate token: %w", err)
	}

	// Cache user for future lookups (async to not block response)
	go s.cacheUserAsync(createdUser)

	// Return response - using value from created user to ensure consistency
	return model.AuthResponse{
		Email: createdUser.Email,
		Phone: "",
		Token: token,
	}, nil
}

// Login optimized for high-frequency authentication requests
// Uses aggressive caching and optimized value passing
func (s *UserService) Login(ctx context.Context, payload model.User) (model.AuthResponse, error) {
	// Early validation - fail fast for invalid input
	if payload.Email == "" {
		return model.AuthResponse{}, ErrInvalidData
	}
	if payload.Password == "" {
		return model.AuthResponse{}, ErrInvalidData
	}

	// // Check rate limiting using cache (prevents brute force)
	// if s.isRateLimited(ctx, payload.Email) {
	// 	return model.AuthResponse{}, ErrTooManyAttempts
	// }

	// Try to get user from cache first (hot path optimization)
	user, err := s.getUserWithCache(ctx, payload.Email)
	if err != nil {
		return model.AuthResponse{}, ErrUserNotFound
	}

	// Compare password hash - this is CPU intensive
	if err := s.userUtils.ComparePasswordHash(user.Password, payload.Password); err != nil {
		return model.AuthResponse{}, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(&user)
	if err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to generate token: %w", err)
	}

	// Return using payload email to avoid any potential inconsistencies
	return model.AuthResponse{
		Email: payload.Email,
		Token: token,
	}, nil
}

// checkUserExists checks cache first, then database
func (s *UserService) checkUserExists(ctx context.Context, email string) (bool, error) {
	// Check cache first for recent lookups
	cacheKey := fmt.Sprintf("user_exists:%s", email)
	if exists, err := s.cache.Exists(ctx, cacheKey); err == nil && exists {
		return true, nil
	}

	// Check database if not in cache
	exists, err := s.userRepo.CheckUserExists(ctx, email)
	if err != nil {
		return false, err
	}

	// Cache the result for 5 minutes to reduce DB hits
	if exists {
		s.cache.Set(ctx, cacheKey, "1", 5*time.Minute)
	}

	return exists, nil
}

// getUserWithCache attempts cache first, fallback to database
func (s *UserService) getUserWithCache(ctx context.Context, email string) (model.User, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("user:%s", email)
	var user model.User

	if err := s.cache.Get(ctx, cacheKey, &user); err == nil {
		return user, nil
	}

	// Cache miss - get from database
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return model.User{}, err
	}

	// Cache for future requests (15 minutes)
	go func() {
		s.cache.Set(context.Background(), cacheKey, user, 15*time.Minute)
	}()

	return user, nil
}

// cacheUserAsync caches user data asynchronously to not block response
func (s *UserService) cacheUserAsync(user model.User) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%s", user.Email)
	s.cache.Set(ctx, cacheKey, user, 15*time.Minute)

	// Also cache existence for faster duplicate checks
	existsKey := fmt.Sprintf("user_exists:%s", user.Email)
	s.cache.Set(ctx, existsKey, "1", 5*time.Minute)
}
