package model

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User Models
type UserAuth struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

type EmailAuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=32"`
}

type PhoneAuthRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=8,max=32"`
}

type AuthResponse struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
	Token string `json:"token"`
}

// Password utility functions
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
