// repositories/purchase.go
package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/teammachinist/tutuplapak/internal/database"
	"github.com/teammachinist/tutuplapak/services/core/models"
)

type PurchaseRepositoryInterface interface {
	CreatePurchase(ctx context.Context, req models.PurchaseRequest) (models.PurchaseResponse, error)
}

type PurchaseRepository struct {
	db *pgxpool.Pool
}

func NewPurchaseRepository(db *pgxpool.Pool) PurchaseRepositoryInterface {
	return &PurchaseRepository{db: db}
}

func (r *PurchaseRepository) CreatePurchase(ctx context.Context, req models.PurchaseRequest) (models.PurchaseResponse, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return models.PurchaseResponse{}, err
	}
	defer tx.Rollback(ctx)

	q := database.New(tx)
	now := time.Now().UTC()
	purchaseID := uuid.Must(uuid.NewV7())

	var snapshots []models.PurchasedItemSnapshot
	sellerTotals := make(map[uuid.UUID]int)
	var purchasedItems []models.ProductResponse // Akan diisi tanpa FileURI dulu

	for _, itemReq := range req.PurchasedItems {
		//  Ambil produk langsung dari DB (dalam transaksi)
		productInTx, err := q.GetProductByID(ctx, itemReq.ProductID)
		if err != nil {
			return models.PurchaseResponse{}, errors.New("product not found in transaction")
		}

		//  Validasi stok
		if itemReq.Qty > int(productInTx.Qty) {
			return models.PurchaseResponse{}, errors.New("insufficient stock for product: " + productInTx.Name)
		}

		//  Update stok (update stock ketika upload file)
		// updateParams := database.UpdateProductQtyParams{
		// 	ID:  itemReq.ProductID,
		// 	Qty: itemReq.Qty,
		// }
		// rowsAffected, err := q.UpdateProductQty(ctx, updateParams)
		// if err != nil {
		// 	return models.PurchaseResponse{}, err
		// }
		// if rowsAffected == 0 {
		// 	return models.PurchaseResponse{}, errors.New("failed to update stock")
		// }

		//  Simpan snapshot — tanpa FileURI (akan diisi di service)
		snapshot := models.PurchasedItemSnapshot{
			ProductID: itemReq.ProductID,
			Name:      productInTx.Name,
			Category:  productInTx.Category,
			Qty:       itemReq.Qty,
			Price:     int(productInTx.Price),
			SKU:       productInTx.Sku,
			FileID:    productInTx.FileID, // ← Simpan FileID
			SellerID:  productInTx.UserID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		snapshots = append(snapshots, snapshot)
		sellerTotals[productInTx.UserID] += int(productInTx.Price) * itemReq.Qty

		//  Tambahkan ke purchasedItems — tanpa FileURI dulu
		purchasedItems = append(purchasedItems, models.ProductResponse{
			ProductID:        productInTx.ID,
			Name:             productInTx.Name,
			Category:         productInTx.Category,
			Qty:              itemReq.Qty,
			Price:            int(productInTx.Price),
			SKU:              productInTx.Sku,
			FileID:           productInTx.FileID, // ← Simpan FileID
			FileURI:          "",                 // ← Akan diisi di service
			FileThumbnailURI: "",                 // ← Akan diisi di service
			CreatedAt:        now,
			UpdatedAt:        now,
		})
	}

	//  Generate payment details
	var paymentDetails []models.PaymentDetail
	for sellerID, total := range sellerTotals {
		userInTx, err := q.GetUserByID(ctx, sellerID)
		if err != nil {
			paymentDetails = append(paymentDetails, models.PaymentDetail{
				BankAccountName:   "",
				BankAccountHolder: "",
				BankAccountNumber: "",
				TotalPrice:        total,
			})
			continue
		}

		bankAccountName := ""
		if userInTx.BankAccountName != nil {
			bankAccountName = *userInTx.BankAccountName
		}

		bankAccountHolder := ""
		if userInTx.BankAccountHolder != nil {
			bankAccountHolder = *userInTx.BankAccountHolder
		}

		bankAccountNumber := ""
		if userInTx.BankAccountNumber != nil {
			bankAccountNumber = *userInTx.BankAccountNumber
		}

		paymentDetails = append(paymentDetails, models.PaymentDetail{
			BankAccountName:   bankAccountName,
			BankAccountHolder: bankAccountHolder,
			BankAccountNumber: bankAccountNumber,
			TotalPrice:        total,
		})
	}

	//  Hitung grand total
	grandTotal := 0
	for _, total := range sellerTotals {
		grandTotal += total
	}

	//  Serialize dan simpan
	snapshotsJSON, err := json.Marshal(snapshots)
	if err != nil {
		return models.PurchaseResponse{}, err
	}

	paymentDetailsJSON, err := json.Marshal(paymentDetails)
	if err != nil {
		return models.PurchaseResponse{}, err
	}

	createParams := database.CreatePurchaseParams{
		ID:                  purchaseID,
		SenderName:          req.SenderName,
		SenderContactType:   req.SenderContactType,
		SenderContactDetail: req.SenderContactDetail,
		PurchasedItems:      snapshotsJSON,
		PaymentDetails:      paymentDetailsJSON,
		TotalPrice:          int32(grandTotal),
	}
	err = q.CreatePurchase(ctx, createParams)
	if err != nil {
		return models.PurchaseResponse{}, err
	}

	//  Commit
	if err := tx.Commit(ctx); err != nil {
		return models.PurchaseResponse{}, errors.New("failed to commit transaction")
	}

	//  Return response — tanpa FileURI (akan diisi di service)
	return models.PurchaseResponse{
		PurchaseID:     purchaseID,
		PurchasedItems: purchasedItems,
		TotalPrice:     grandTotal,
		PaymentDetails: paymentDetails,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}
