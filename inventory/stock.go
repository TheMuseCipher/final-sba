package inventory

import (
	"database/sql"
	"time"

	"ims-go/models"
)

func RestockItem(db Database, itemID int, quantity int, expiryDate *time.Time) error {
	now := time.Now()
	// Update item quantity
	_, err := db.GetDB().Exec(
		"UPDATE items SET quantity = quantity + ?, in_stock_date = ?, updated_at = ? WHERE id = ?",
		quantity, now, now, itemID,
	)
	if err != nil {
		return err
	}

	// Create stock entry
	_, err = db.GetDB().Exec(
		"INSERT INTO item_stock (item_id, quantity, in_stock_date, expiry_date) VALUES (?, ?, ?, ?)",
		itemID, quantity, now, expiryDate,
	)
	return err
}

func GetItemStockBatches(db Database, itemID int) ([]models.ItemStock, error) {
	rows, err := db.GetDB().Query(
		"SELECT id, item_id, quantity, in_stock_date, expiry_date FROM item_stock WHERE item_id = ? ORDER BY in_stock_date ASC",
		itemID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var batches []models.ItemStock
	for rows.Next() {
		var batch models.ItemStock
		var inStockDate time.Time
		var expiryDate sql.NullTime

		err := rows.Scan(&batch.ID, &batch.ItemID, &batch.Quantity, &inStockDate, &expiryDate)
		if err != nil {
			return nil, err
		}

		batch.InStockDate = inStockDate
		if expiryDate.Valid {
			batch.ExpiryDate = &expiryDate.Time
		}
		batches = append(batches, batch)
	}

	return batches, nil
}

func GetItemQuantity(db Database, itemID int) (int, error) {
	var quantity int
	err := db.GetDB().QueryRow("SELECT quantity FROM items WHERE id = ?", itemID).Scan(&quantity)
	if err != nil {
		return 0, err
	}
	return quantity, nil
}

