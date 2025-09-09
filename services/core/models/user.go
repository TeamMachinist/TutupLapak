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
