package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"kasir-api/models"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (repo *TransactionRepository) CreateTransaction(items []models.CheckoutItem) (*models.Transaction, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	totalAmount := 0
	details := make([]models.TransactionDetail, 0)

	for _, item := range items {
		var productPrice, stock int
		var productName string

		err := tx.QueryRow("SELECT name, price, stock FROM products WHERE id = $1", item.ProductID).Scan(&productName, &productPrice, &stock)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product id %d not found", item.ProductID)
		}
		if err != nil {
			return nil, err
		}

		subtotal := productPrice * item.Quantity
		totalAmount += subtotal

		_, err = tx.Exec("UPDATE products SET stock = stock - $1 WHERE id = $2", item.Quantity, item.ProductID)
		if err != nil {
			return nil, err
		}

		details = append(details, models.TransactionDetail{
			ProductID:   item.ProductID,
			ProductName: productName,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	var transactionID int
	err = tx.QueryRow("INSERT INTO transactions (total_amount) VALUES ($1) RETURNING id", totalAmount).Scan(&transactionID)
	if err != nil {
		return nil, err
	}

	// pakai ctx
	ctx := context.Background()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO transaction_details (transaction_id, product_id, quantity, subtotal)
     VALUES ($1, $2, $3, $4)`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for i := range details {
		details[i].TransactionID = transactionID

		// Validasi ringan opsional
		if details[i].Quantity <= 0 {
			return nil, fmt.Errorf("quantity untuk product_id %d harus > 0", details[i].ProductID)
		}
		if details[i].Subtotal < 0 {
			return nil, fmt.Errorf("subtotal untuk product_id %d tidak boleh negatif", details[i].ProductID)
		}

		res, err := stmt.ExecContext(ctx,
			details[i].TransactionID,
			details[i].ProductID,
			details[i].Quantity,
			details[i].Subtotal,
		)
		if err != nil {
			return nil, err
		}
		if rows, _ := res.RowsAffected(); rows == 0 {
			return nil, fmt.Errorf("gagal insert transaction_detail untuk product_id=%d", details[i].ProductID)
		}
	}

	// for i := range details {
	// 	details[i].TransactionID = transactionID
	// 	_, err = tx.Exec("INSERT INTO transaction_details (transaction_id, product_id, quantity, subtotal) VALUES ($1, $2, $3, $4)",
	// 		transactionID, details[i].ProductID, details[i].Quantity, details[i].Subtotal)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.Transaction{
		ID:          transactionID,
		TotalAmount: totalAmount,
		Details:     details,
	}, nil
}
