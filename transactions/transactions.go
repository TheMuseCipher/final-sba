package transactions

import (
	"database/sql"
	"time"

	"ims-go/models"
)

type Database interface {
	GetDB() *sql.DB
}

func CreateTransaction(db Database, userID int, items []models.TransactionItem) (*models.Transaction, error) {
	// Calculate total
	var totalAmount float64
	for _, item := range items {
		totalAmount += item.Price * float64(item.Quantity)
	}

	// Create transaction
	result, err := db.GetDB().Exec(
		"INSERT INTO transactions (user_id, total_amount, created_at) VALUES (?, ?, ?)",
		userID, totalAmount, time.Now(),
	)
	if err != nil {
		return nil, err
	}

	transactionID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Create transaction items and update inventory
	for _, item := range items {
		_, err := db.GetDB().Exec(
			"INSERT INTO transaction_items (transaction_id, item_id, quantity, price) VALUES (?, ?, ?, ?)",
			transactionID, item.ItemID, item.Quantity, item.Price,
		)
		if err != nil {
			return nil, err
		}

		// Update item quantity
		var currentQuantity int
		err = db.GetDB().QueryRow("SELECT quantity FROM items WHERE id = ?", item.ItemID).Scan(&currentQuantity)
		if err != nil {
			return nil, err
		}

		newQuantity := currentQuantity - item.Quantity
		if newQuantity < 0 {
			newQuantity = 0
		}

		_, err = db.GetDB().Exec("UPDATE items SET quantity = ?, updated_at = ? WHERE id = ?", newQuantity, time.Now(), item.ItemID)
		if err != nil {
			return nil, err
		}
	}

	return GetTransactionByID(db, int(transactionID))
}

func GetTransactionByID(db Database, id int) (*models.Transaction, error) {
	var transaction models.Transaction
	var createdAt time.Time

	err := db.GetDB().QueryRow(
		"SELECT id, user_id, total_amount, created_at FROM transactions WHERE id = ?",
		id,
	).Scan(&transaction.ID, &transaction.UserID, &transaction.TotalAmount, &createdAt)

	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	transaction.CreatedAt = createdAt

	// Get transaction items
	rows, err := db.GetDB().Query(
		`SELECT ti.id, ti.transaction_id, ti.item_id, ti.quantity, ti.price, i.name 
		 FROM transaction_items ti 
		 JOIN items i ON ti.item_id = i.id 
		 WHERE ti.transaction_id = ?`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.TransactionItem
		err := rows.Scan(&item.ID, &item.TransactionID, &item.ItemID, &item.Quantity, &item.Price, &item.ItemName)
		if err != nil {
			return nil, err
		}
		transaction.Items = append(transaction.Items, item)
	}

	return &transaction, nil
}

func GetRecentTransactions(db Database, limit int) ([]models.Transaction, error) {
	rows, err := db.GetDB().Query(
		"SELECT id, user_id, total_amount, created_at FROM transactions ORDER BY created_at DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		var createdAt time.Time

		err := rows.Scan(&transaction.ID, &transaction.UserID, &transaction.TotalAmount, &createdAt)
		if err != nil {
			return nil, err
		}

		transaction.CreatedAt = createdAt

		// Get transaction items
		itemRows, err := db.GetDB().Query(
			`SELECT ti.id, ti.transaction_id, ti.item_id, ti.quantity, ti.price, i.name 
			 FROM transaction_items ti 
			 JOIN items i ON ti.item_id = i.id 
			 WHERE ti.transaction_id = ?`,
			transaction.ID,
		)
		if err != nil {
			return nil, err
		}

		for itemRows.Next() {
			var item models.TransactionItem
			err := itemRows.Scan(&item.ID, &item.TransactionID, &item.ItemID, &item.Quantity, &item.Price, &item.ItemName)
			if err != nil {
				itemRows.Close()
				return nil, err
			}
			transaction.Items = append(transaction.Items, item)
		}
		itemRows.Close()

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

