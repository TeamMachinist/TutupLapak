package main

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtService interface {
	GenerateToken(payload *model.User) (string, error)
	VerifyToken(tokenString string) (jwt.MapClaims, error)
}

type SecurityConfig struct {
	Key      string
	Duration time.Duration
	Issues   string
}

type jwtService struct {
	config *SecurityConfig
	// Pre-convert key to []byte to avoid repeated conversions
	keyBytes []byte
}

// Custom claims untuk JWT
type JwtTokenClaims struct {
	jwt.RegisteredClaims
	UserId uuid.UUID `json:"user_id"`
}

// GenerateToken implements JwtService.
func (j *jwtService) GenerateToken(payload *model.User) (string, error) {
	now := time.Now()
	claims := JwtTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.Issues,
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.Duration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserId: payload.ID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Use pre-converted keyBytes instead of converting each time
	accessToken, err := token.SignedString(j.keyBytes)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// VerifyToken implements JwtService.
func (j *jwtService) VerifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		// Use pre-converted keyBytes
		return j.keyBytes, nil
	})

	if err != nil {
		return nil, errors.New("failed to parse token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !token.Valid || !ok {
		return nil, errors.New("invalid token or claims")
	}

	// Verify issuer
	if iss, ok := claims["iss"].(string); !ok || iss != j.config.Issues {
		return nil, errors.New("invalid issuer")
	}

	return claims, nil
}

func NewJwtService(config *SecurityConfig) JwtService {
	return &jwtService{
		config:   config,
		keyBytes: []byte(config.Key), // Pre-convert once during initialization
	}
}
