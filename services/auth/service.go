package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthService struct {
	authRepo   *repository.AuthRepository
	cache      *cache.Redis
	hasher     utils.PasswordHasher
	jwtService JwtService
}

func NewAuthService(authRepo *repository.AuthRepository, cache *cache.Redis, jwt JwtService) *AuthService {
	return &AuthService{
		authRepo:   authRepo,
		cache:      cache,
		hasher:  	utils.NewPasswordHasher(), // Consider making this a singleton if expensive to create
		jwtService: jwt,
	}
}

// RegisterNewUser optimized for performance with value passing
// Uses stack allocation for better cache locality and reduced GC pressure
func (s *AuthService) RegisterWithEmail(ctx context.Context, payload model.EmailAuthRequest) (model.AuthResponse, error) {
	// Early context check to avoid unnecessary work
	select {
	case <-ctx.Done():
		return model.AuthResponse{}, ctx.Err()
	default:
	}

	// Check if email exists using cache first, then database
	exists, err := s.checkExistedUserByEmailWithCache(ctx, payload.Email)
	if err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return model.AuthResponse{}, error.ErrConflict
	}

	// Hash password - this is CPU intensive, do it early
	hashedPassword, err := s.hasher.EncryptPassword(payload.Password)
	if err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to encrypt password: %w", err)
	}

	payload.password := hashedPassword

	// Another context check before expensive DB operation
	select {
	case <-ctx.Done():
		return model.AuthResponse{}, ctx.Err()
	default:
	}

	// Create user in database
	user := UsersAuth{}
	
	var pgUUID pgtype.UUID
	copy(pgUUID.Bytes[:], gUUID[:])
	pgUUID.Status = pgtype.Present
	
	user.ID = pgUUID
	user.Phone = ""
	user.HashedPassword = payload.hashed_password
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	
	createdUser, err := s.authRepo.RegisterNewUser(ctx, user)
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
func (s *AuthService) LoginWithEmail(ctx context.Context, payload model.User) (model.AuthResponse, error) {
	// Early validation - fail fast for invalid input
	if payload.Email == "" {
		return model.AuthResponse{}, errors.ErrBadRequest
	}
	if payload.Password == "" {
		return model.AuthResponse{}, errors.ErrBadRequest
	}

	// // Check rate limiting using cache (prevents brute force)
	// if s.isRateLimited(ctx, payload.Email) {
	// 	return model.AuthResponse{}, ErrTooManyAttempts
	// }

	// Try to get user from cache first (hot path optimization)
	user, err := s.getUserByEmailWithCache(ctx, payload.Email)
	if err != nil {
		return model.AuthResponse{}, errors.ErrNotFound
	}

	// Compare password hash - this is CPU intensive
	if err := s.hasher.ComparePasswordHash(user.Password, payload.Password); err != nil {
		return model.AuthResponse{}, errors.ErrUnauthorized
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(&user)
	if err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to generate token: %w", err)
	}

	// Return using payload email to avoid any potential inconsistencies
	return model.AuthResponse{
		Email: payload.Email,
		Phone: "",
		Token: token,
	}, nil
}

// checkUserExists checks cache first, then database
func (s *AuthService) CheckExistedUserByEmailWithCache(ctx context.Context, email string) (bool, error) {
	// Check cache first for recent lookups
	cacheKey := fmt.Sprintf("user_exists:%s", email)
	if exists, err := s.cache.Exists(ctx, cacheKey); err == nil && exists {
		return true, nil
	}

	// Check database if not in cache
	exists, err := s.authRepo.CheckExistedUserByEmail(ctx, email)
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
func (s *AuthService) getUserByEmailWithCache(ctx context.Context, email string) (model.User, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("user:%s", email)
	var user model.User

	if err := s.cache.Get(ctx, cacheKey, &user); err == nil {
		return user, nil
	}

	// Cache miss - get from database
	user, err := s.authRepo.GetUserByEmail(ctx, email)
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
func (s *AuthService) cacheUserAsync(user model.User) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%s", user.Email)
	s.cache.Set(ctx, cacheKey, user, 15*time.Minute)

	// Also cache existence for faster duplicate checks
	existsKey := fmt.Sprintf("user_exists:%s", user.Email)
	s.cache.Set(ctx, existsKey, "1", 5*time.Minute)
}
