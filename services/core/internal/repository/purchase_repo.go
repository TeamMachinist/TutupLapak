package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/teammachinist/tutuplapak/services/core/internal/database"
	"github.com/teammachinist/tutuplapak/services/core/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PurchaseRepositoryInterface interface {
	CreatePurchase(ctx context.Context, req model.PurchaseRequest) (model.PurchaseResponse, error)
	GetPurchaseByid(ctx context.Context, purchaseId string) (model.PurchaseResponse, error)
	UpdatePurchaseStatus(ctx context.Context, purchaseId string, newStatus database.PurchaseStatus) error
}

type PurchaseRepository struct {
	db     *pgxpool.Pool
	dbSqlc database.Querier
}

// GetPurchaseByid implements PurchaseRepositoryInterface.
func (r *PurchaseRepository) GetPurchaseByid(ctx context.Context, purchaseId string) (model.PurchaseResponse, error) {
	parsedId, err := uuid.Parse(purchaseId)
	if err != nil {
		return model.PurchaseResponse{}, err
	}

	row, err := r.dbSqlc.GetPurchaseByID(ctx, parsedId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PurchaseResponse{}, nil
		}
		return model.PurchaseResponse{}, err
	}

	var paymentDetails []model.PaymentDetail
	if err := json.Unmarshal(row.PaymentDetails, &paymentDetails); err != nil {
		return model.PurchaseResponse{}, err
	}

	var purchasedItems []model.ProductResponse
	if err := json.Unmarshal(row.PurchasedItems, &purchasedItems); err != nil {
		return model.PurchaseResponse{}, err
	}

	return model.PurchaseResponse{
		PurchaseID:     row.ID,
		PurchasedItems: purchasedItems,
		TotalPrice:     row.TotalPrice,
		PaymentDetails: paymentDetails,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
		Status:         model.PurchaseStatus(row.Status),
	}, nil
}

// UpdatePurchaseStatus implements PurchaseRepositoryInterface.
func (r *PurchaseRepository) UpdatePurchaseStatus(ctx context.Context, purchaseId string, newStatus database.PurchaseStatus) error {
	parsedId, err := uuid.Parse(purchaseId)
	if err != nil {
		return err
	}

	return r.dbSqlc.UpdatePurchaseStatus(ctx, database.UpdatePurchaseStatusParams{
		Purchaseid: parsedId,
		Status:     newStatus,
	})
}

func (r *PurchaseRepository) CreatePurchase(ctx context.Context, req model.PurchaseRequest) (model.PurchaseResponse, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return model.PurchaseResponse{}, err
	}
	defer tx.Rollback(ctx)

	q := database.New(tx)
	now := time.Now().UTC()
	purchaseID := uuid.Must(uuid.NewV7())

	var snapshots []model.PurchasedItemSnapshot
	sellerTotals := make(map[uuid.UUID]int)
	var purchasedItems []model.ProductResponse // Akan diisi tanpa FileURI dulu

	for _, itemReq := range req.PurchasedItems {
		//  Ambil produk langsung dari DB (dalam transaksi)
		productInTx, err := q.GetProductByID(ctx, itemReq.ProductID)
		if err != nil {
			return model.PurchaseResponse{}, errors.New("product not found")
		}

		//  Validasi stok
		if itemReq.Qty > productInTx.Qty {
			return model.PurchaseResponse{}, errors.New("insufficient stock for product: " + productInTx.Name)
		}

		//  Update stok (update stock ketika upload file)
		// updateParams := database.UpdateProductQtyParams{
		// 	ID:  itemReq.ProductID,
		// 	Qty: itemReq.Qty,
		// }
		// rowsAffected, err := q.UpdateProductQty(ctx, updateParams)
		// if err != nil {
		// 	return model.PurchaseResponse{}, err
		// }
		// if rowsAffected == 0 {
		// 	return model.PurchaseResponse{}, errors.New("failed to update stock")
		// }

		//  Simpan snapshot — tanpa FileURI (akan diisi di service)
		snapshot := model.PurchasedItemSnapshot{
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
		purchasedItems = append(purchasedItems, model.ProductResponse{
			ProductID:        productInTx.ID,
			Name:             productInTx.Name,
			Category:         productInTx.Category,
			Qty:              itemReq.Qty,
			Price:            int(productInTx.Price),
			SKU:              productInTx.Sku,
			FileID:           productInTx.FileID,
			UserID:           productInTx.UserID, // ← isi untuk kebutuhan internal
			FileURI:          "",
			FileThumbnailURI: "",
			CreatedAt:        now,
			UpdatedAt:        now,
		})
	}

	//  Generate payment details
	var paymentDetails []model.PaymentDetail
	for sellerID, total := range sellerTotals {
		userInTx, err := q.GetUserByID(ctx, sellerID)
		if err != nil {
			paymentDetails = append(paymentDetails, model.PaymentDetail{
				BankAccountName:   "",
				BankAccountHolder: "",
				BankAccountNumber: "",
				TotalPrice:        total,
			})
			continue
		}

		bankAccountName := ""
		// if userInTx.BankAccountName != nil {
		if userInTx.BankAccountName != "" {
			// bankAccountName = *userInTx.BankAccountName
			bankAccountName = userInTx.BankAccountName
		}

		bankAccountHolder := ""
		// if userInTx.BankAccountHolder != nil {
		if userInTx.BankAccountHolder != "" {
			// bankAccountHolder = *userInTx.BankAccountHolder
			bankAccountHolder = userInTx.BankAccountHolder
		}

		bankAccountNumber := ""
		// if userInTx.BankAccountNumber != nil {
		if userInTx.BankAccountNumber != "" {
			// bankAccountNumber = *userInTx.BankAccountNumber
			bankAccountNumber = userInTx.BankAccountNumber
		}

		paymentDetails = append(paymentDetails, model.PaymentDetail{
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
		return model.PurchaseResponse{}, err
	}

	paymentDetailsJSON, err := json.Marshal(paymentDetails)
	if err != nil {
		return model.PurchaseResponse{}, err
	}

	createParams := database.CreatePurchaseParams{
		ID:                  purchaseID,
		SenderName:          req.SenderName,
		SenderContactType:   req.SenderContactType,
		SenderContactDetail: req.SenderContactDetail,
		PurchasedItems:      snapshotsJSON,
		PaymentDetails:      paymentDetailsJSON,
		TotalPrice:          grandTotal,
		// TotalPrice:          int32(grandTotal),
	}
	err = q.CreatePurchase(ctx, createParams)
	if err != nil {
		return model.PurchaseResponse{}, err
	}

	//  Commit
	if err := tx.Commit(ctx); err != nil {
		return model.PurchaseResponse{}, errors.New("failed to commit transaction")
	}

	//  Return response — tanpa FileURI (akan diisi di service)
	return model.PurchaseResponse{
		PurchaseID:     purchaseID,
		PurchasedItems: purchasedItems,
		TotalPrice:     grandTotal,
		PaymentDetails: paymentDetails,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func NewPurchaseRepository(db *pgxpool.Pool, dbSqlc database.Querier) PurchaseRepositoryInterface {
	return &PurchaseRepository{db: db, dbSqlc: dbSqlc}
}
