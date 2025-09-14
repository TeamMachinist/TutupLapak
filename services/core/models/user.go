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
