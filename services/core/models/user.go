package models

import "github.com/google/uuid"

type User struct {
	ID                uuid.UUID `json:"id"`
	Email             string    `json:"email"`
	Phone             string    `json:"phone"`
	BankAccountName   string    `json:"bankAccountName"`
	BankAccountHolder string    `json:"bankAccountHolder"`
	BankAccountNumber string    `json:"bankAccountNumber"`
}

type LinkPhoneRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type LinkPhoneResponse struct {
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	FileID            string `json:"fileId"`
	FileURI           string `json:"fileUri"`
	FileThumbnailURI  string `json:"fileThumbnailUri"`
	BankAccountName   string `json:"bankAccountName"`
	BankAccountHolder string `json:"bankAccountHolder"`
	BankAccountNumber string `json:"bankAccountNumber"`
}

type LinkEmailRequest struct {
	Email string `json:"email" binding:"required"`
}

type LinkEmailResponse struct {
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	FileID            string `json:"fileId"`
	FileURI           string `json:"fileUri"`
	FileThumbnailURI  string `json:"fileThumbnailUri"`
	BankAccountName   string `json:"bankAccountName"`
	BankAccountHolder string `json:"bankAccountHolder"`
	BankAccountNumber string `json:"bankAccountNumber"`
}

type UserRequest struct {
	FileID            *string `json:"fileId"`
	BankAccountName   string  `json:"bankAccountName" validate:"required,min=4,max=32"`
	BankAccountHolder string  `json:"bankAccountHolder" validate:"required,min=4,max=32"`
	BankAccountNumber string  `json:"bankAccountNumber" validate:"required,min=4,max=32"`
}

type UserResponse struct {
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	FileID            string `json:"fileId"`
	URI               string `json:"fileUri"`
	ThumbnailURI      string `json:"fileThumbnailUri"`
	BankAccountName   string `json:"bankAccountName"`
	BankAccountHolder string `json:"bankAccountHolder"`
	BankAccountNumber string `json:"bankAccountNumber"`
}
