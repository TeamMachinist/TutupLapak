package internal

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Errors
var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
	ErrMissingToken = errors.New("authorization token required")
)

// Claims
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// Config
type JWTConfig struct {
	Key      string
	Duration time.Duration
	Issuer   string
}

// Service
type JWTService struct {
	config *JWTConfig
}

func NewJWTService(config *JWTConfig) *JWTService {
	return &JWTService{config: config}
}

func (j *JWTService) GenerateToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.config.Duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.Key))
}

func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.config.Key), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

func (j *JWTService) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrMissingToken
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", ErrInvalidToken
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", ErrMissingToken
	}

	return token, nil
}

// Context helpers
type contextKey string

const UserIDKey contextKey = "user_id"

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// Fiber Middleware
func (j *JWTService) FiberMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		token, err := j.ExtractTokenFromHeader(authHeader)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		claims, err := j.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		c.Locals("user_id", claims.UserID)
		return c.Next()
	}
}

func GetUserIDFromFiber(c *fiber.Ctx) (string, bool) {
	userID := c.Locals("user_id")
	if userID == nil {
		return "", false
	}
	return userID.(string), true
}

// Chi Middleware
func (j *JWTService) ChiMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		token, err := j.ExtractTokenFromHeader(authHeader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		claims, err := j.ValidateToken(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := WithUserID(r.Context(), claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromChi(r *http.Request) (string, bool) {
	return GetUserID(r.Context())
}
