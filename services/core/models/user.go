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

type UserRequest struct {
	FileID            string    `json:"fileId"`
	BankAccountName   string    `json:"bankAccountName"`
	BankAccountHolder string    `json:"bankAccountHolder"`
	BankAccountNumber string    `json:"bankAccountNumber"`
}

type UserResponse struct {
	FileID            uuid.UUID `json:"fileId"`
	URI               string    `json:"uri"`
	ThumbnailURI      string    `json:"thumbnailUri"`
	Email             string    `json:"email"`
	Phone             string    `json:"phone"`
	BankAccountName   string    `json:"bankAccountName"`
	BankAccountHolder string    `json:"bankAccountHolder"`
	BankAccountNumber string    `json:"bankAccountNumber"`
}
