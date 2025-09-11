package auth

import "github.com/google/uuid"

type EmailAuthRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type PhoneAuthRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
	Token string `json:"token"`
}
