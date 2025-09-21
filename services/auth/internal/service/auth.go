package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"

	"github.com/teammachinist/tutuplapak/services/auth/internal/cache"
	"github.com/teammachinist/tutuplapak/services/auth/internal/database"
	"github.com/teammachinist/tutuplapak/services/auth/internal/logger"
	"github.com/teammachinist/tutuplapak/services/auth/internal/model"
	"github.com/teammachinist/tutuplapak/services/auth/internal/repository"
	"github.com/teammachinist/tutuplapak/services/auth/pkg/authz"
)

type UserService struct {
	userRepo       *repository.UserRepository
	db             *database.Queries
	jwtConfig      *JWTConfig
	coreServiceURL string
	httpClient     *http.Client
	cache          *cache.RedisCache
}

type JWTConfig struct {
	Key      string
	Duration time.Duration
	Issuer   string
}

type TokenClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

const (
	UserAuthByEmailKey = "userauth:email:%s"
	UserAuthByPhoneKey = "userauth:phone:%s"
	TokenValidationKey = "token:valid:%s"
)

const (
	UserAuthTTL        = 15 * time.Minute
	TokenValidationTTL = 5 * time.Minute
)

func NewUserService(userRepo *repository.UserRepository, db *database.Queries, jwtConfig *JWTConfig, coreServiceURL string, cache *cache.RedisCache) *UserService {
	return &UserService{
		userRepo:       userRepo,
		db:             db,
		jwtConfig:      jwtConfig,
		coreServiceURL: coreServiceURL,
		cache:          cache,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *UserService) GenerateToken(userID string) (string, error) {
	claims := TokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtConfig.Duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.jwtConfig.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtConfig.Key))
}

func (s *UserService) ValidateTokenInternal(tokenString string) (*authz.Claims, error) {
	// Simple cache key (first 16 chars of token for security)
	cacheKey := fmt.Sprintf(TokenValidationKey, tokenString[:min(16, len(tokenString))])

	// Try cache first
	var cachedClaims authz.Claims
	if err := s.cache.Get(context.Background(), cacheKey, &cachedClaims); err == nil {
		return &cachedClaims, nil
	}

	// Cache miss - validate token
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtConfig.Key), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token has expired")
		}
		return nil, errors.New("invalid token")
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		result := &authz.Claims{UserID: claims.UserID}

		// Cache valid token for short time
		s.cache.Set(context.Background(), cacheKey, result, TokenValidationTTL)

		return result, nil
	}

	return nil, errors.New("invalid token")
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

	// Try cache first
	cacheKey := fmt.Sprintf(UserAuthByEmailKey, email)
	var cachedUser model.UserAuth

	var userAuth *model.UserAuth
	var err error

	if cacheErr := s.cache.Get(ctx, cacheKey, &cachedUser); cacheErr == nil {
		userAuth = &cachedUser
	} else {
		// Cache miss - get from DB
		userAuth, err = s.userRepo.GetUserAuthByEmail(ctx, email)
		if err != nil {
			return nil, errors.New("email not found")
		}

		// Cache for future use
		s.cache.Set(ctx, cacheKey, userAuth, UserAuthTTL)
	}

	// Check password
	if !model.CheckPassword(password, userAuth.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	coreUser, err := s.GetUserFromCore(userAuth.ID.String())
	if err != nil {
		return nil, errors.New("user profile not found")
	}

	// Generate token with Core's user.id
	token, err := s.GenerateToken(coreUser.ID)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &model.AuthResponse{
		Email: userAuth.Email,
		Phone: userAuth.Phone,
		Token: token,
	}, nil
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

	// Try cache first
	cacheKey := fmt.Sprintf(UserAuthByPhoneKey, phone)
	var cachedUser model.UserAuth

	var userAuth *model.UserAuth
	var err error

	if cacheErr := s.cache.Get(ctx, cacheKey, &cachedUser); cacheErr == nil {
		userAuth = &cachedUser
	} else {
		// Cache miss - get from DB
		userAuth, err = s.userRepo.GetUserByPhone(ctx, phone)
		if err != nil {
			return nil, errors.New("phone not found")
		}

		// Cache for future use
		s.cache.Set(ctx, cacheKey, userAuth, UserAuthTTL)
	}

	// Check password
	if !model.CheckPassword(password, userAuth.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	coreUser, err := s.GetUserFromCore(userAuth.ID.String())
	if err != nil {
		return nil, errors.New("user profile not found")
	}

	// Generate token with Core's user.id
	token, err := s.GenerateToken(coreUser.ID)
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

	// Wait for core service to create user and return user.id
	coreUser, err := s.CreateUserInCoreSync(ctx, userAuth.ID.String(), email, "")
	if err != nil {
		logger.WarnCtx(ctx, "Failed to create user in core service", "user_auth_id", userAuth.ID.String(), "error", err.Error())
		// Rollback: Delete the userAuth record
		if rollbackErr := s.userRepo.DeleteUserAuth(ctx, userAuth.ID); rollbackErr != nil {
			// Log rollback failure but don't change the original error
			logger.WarnCtx(ctx, "Failed to rollback userAuth creation", "user_auth_id", userAuth.ID.String(), "error", rollbackErr)
		}
		return nil, errors.New("failed to create user profile")
	}

	// Generate token with coreUser ID
	token, err := s.GenerateToken(coreUser.ID)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	// Cache the new user
	cacheKey := fmt.Sprintf(UserAuthByEmailKey, email)
	s.cache.Set(ctx, cacheKey, userAuth, UserAuthTTL)

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

	// Wait for core service to create user and return user.id
	coreUser, err := s.CreateUserInCoreSync(ctx, userAuth.ID.String(), "", phone)
	if err != nil {
		logger.WarnCtx(ctx, "Failed to create user in core service", "user_auth_id", userAuth.ID.String(), "error", err.Error())
		// Rollback: Delete the userAuth record
		if rollbackErr := s.userRepo.DeleteUserAuth(ctx, userAuth.ID); rollbackErr != nil {
			// Log rollback failure but don't change the original error
			logger.WarnCtx(ctx, "Failed to rollback userAuth creation", "user_auth_id", userAuth.ID.String(), "error", rollbackErr)
		}
		return nil, errors.New("failed to create user profile")
	}

	// Generate token with user_auth_id
	token, err := s.GenerateToken(coreUser.ID)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	// Cache the new user
	cacheKey := fmt.Sprintf(UserAuthByPhoneKey, phone)
	s.cache.Set(ctx, cacheKey, userAuth, UserAuthTTL)

	return &model.AuthResponse{
		Email: userAuth.Email,
		Phone: userAuth.Phone,
		Token: token,
	}, nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Validation methods - no logging here
func (s *UserService) validatePhone(phone string) error {
	if phone == "" {
		return errors.New("phone is required")
	}

	if !strings.HasPrefix(phone, "+") {
		return errors.New("phone must start with international calling code prefix '+'")
	}

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

// LinkPhone - Link phone number to existing user auth
func (s *UserService) LinkPhone(ctx context.Context, userAuthID, phone string) error {
	if err := s.validatePhone(phone); err != nil {
		return err
	}

	// Check if phone already exists (different user)
	exists, err := s.userRepo.CheckPhoneExists(ctx, phone)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("phone number already exists")
	}

	userAuthUUID, err := uuid.Parse(userAuthID)
	if err != nil {
		return errors.New("invalid user_auth_id format")
	}

	err = s.userRepo.UpdateUserAuthPhone(ctx, userAuthUUID, phone)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return errors.New("user auth not found")
		}
		return err
	}

	// Invalidate cache for this user
	// Optionally add cache by getting the updated user auth data

	return nil
}

// LinkEmail - Link email to existing user auth
func (s *UserService) LinkEmail(ctx context.Context, userAuthID, email string) error {
	// Validate email format
	if err := s.validateEmail(email); err != nil {
		return err
	}

	// Check if email already exists (different user)
	exists, err := s.userRepo.CheckExistedUserAuthByEmail(ctx, email)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("email address already exists")
	}

	userAuthUUID, err := uuid.Parse(userAuthID)
	if err != nil {
		return errors.New("invalid user_auth_id format")
	}

	// Update user auth with email
	err = s.userRepo.UpdateUserAuthEmail(ctx, userAuthUUID, email)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return errors.New("user auth not found")
		}
		return err
	}

	// Invalidate cache for this user
	// Optionally add cache by getting the updated user auth data

	return nil
}

// HTTP calls to Core will be placed here first for a quick refactor
// CreateUserInCoreSync - Call core service and return the created user with user.id
type CoreUser struct {
	ID         string `json:"id"`
	UserAuthID string `json:"user_auth_id"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
}

func (s *UserService) CreateUserInCoreSync(ctx context.Context, userAuthID, email, phone string) (*CoreUser, error) {
	payload := map[string]interface{}{
		"user_auth_id": userAuthID,
		"email":        email,
		"phone":        phone,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.coreServiceURL+"/internal/user", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call core service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		// Try to read error response
		var errorBody map[string]interface{}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&errorBody); decodeErr == nil {
			if errorMsg, ok := errorBody["error"].(string); ok {
				return nil, fmt.Errorf("core service error (status %d): %s", resp.StatusCode, errorMsg)
			}
		}
		return nil, fmt.Errorf("core service returned status %d", resp.StatusCode)
	}

	// Parse the created user response
	var coreUser CoreUser
	if err := json.NewDecoder(resp.Body).Decode(&coreUser); err != nil {
		return nil, fmt.Errorf("failed to decode core service response: %w", err)
	}

	// Validate we got the user ID
	if coreUser.ID == "" {
		return nil, fmt.Errorf("core service did not return user ID")
	}

	return &coreUser, nil
}

// GetUserFromCore - Get user.id from core service using user_auth_id
func (s *UserService) GetUserFromCore(userAuthID string) (*CoreUser, error) {
	req, err := http.NewRequest("GET", s.coreServiceURL+"/internal/user/"+userAuthID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call core service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Try to read error response
		var errorBody map[string]interface{}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&errorBody); decodeErr == nil {
			if errorMsg, ok := errorBody["error"].(string); ok {
				return nil, fmt.Errorf("core service error (status %d): %s", resp.StatusCode, errorMsg)
			}
		}
		return nil, fmt.Errorf("core service returned status %d", resp.StatusCode)
	}

	// Parse the user response
	var coreUser CoreUser
	if err := json.NewDecoder(resp.Body).Decode(&coreUser); err != nil {
		return nil, fmt.Errorf("failed to decode core service response: %w", err)
	}

	// Validate we got the user ID
	if coreUser.ID == "" {
		return nil, fmt.Errorf("core service did not return user ID")
	}

	return &coreUser, nil
}
