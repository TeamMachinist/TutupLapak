// models/purchase.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type PurchaseRequest struct {
	PurchasedItems      []PurchaseItemRequest `json:"purchasedItems"`
	SenderName          string                `json:"senderName"`
	SenderContactType   string                `json:"senderContactType"`
	SenderContactDetail string                `json:"senderContactDetail"`
}

type PurchaseItemRequest struct {
	ProductID uuid.UUID `json:"productId"`
	Qty       int       `json:"qty"`
}

type PurchasedItemSnapshot struct {
	ProductID uuid.UUID `json:"productId" db:"product_id"`
	Name      string    `json:"name" db:"name"`
	Category  string    `json:"category" db:"category"`
	Qty       int       `json:"qty" db:"qty"`
	Price     int       `json:"price" db:"price"`
	SKU       string    `json:"sku" db:"sku"`
	FileID    uuid.UUID `json:"fileId" db:"file_id"`
	SellerID  uuid.UUID `json:"sellerId" db:"seller_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type PurchaseResponse struct {
	PurchaseID     uuid.UUID         `json:"purchaseId" db:"id"`
	PurchasedItems []ProductResponse `json:"purchasedItems" db:"purchased_items"`
	TotalPrice     int               `json:"totalPrice" db:"total_price"`
	PaymentDetails []PaymentDetail   `json:"paymentDetails" db:"payment_details"`
	CreatedAt      time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time         `json:"updatedAt" db:"updated_at"`
}

type PaymentDetail struct {
	BankAccountName   string `json:"bankAccountName" db:"bank_account_name"`
	BankAccountHolder string `json:"bankAccountHolder" db:"bank_account_holder"`
	BankAccountNumber string `json:"bankAccountNumber" db:"bank_account_number"`
	TotalPrice        int    `json:"totalPrice" db:"total_price"`
}
