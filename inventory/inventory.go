package inventory

import (
	"database/sql"
	"errors"
	"time"

	"ims-go/models"
)

type Database interface {
	GetDB() *sql.DB
}

func CreateItem(db Database, name, code, description string, price, cost float64, quantity int) (*models.Item, error) {
	now := time.Now()
	result, err := db.GetDB().Exec(
		"INSERT INTO items (name, code, description, price, cost, quantity, in_stock_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		name, code, description, price, cost, quantity, now, now, now,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Create stock entry
	_, err = db.GetDB().Exec(
		"INSERT INTO item_stock (item_id, quantity, in_stock_date) VALUES (?, ?, ?)",
		id, quantity, now,
	)
	if err != nil {
		return nil, err
	}

	return GetItemByID(db, int(id))
}

func GetItemByID(db Database, id int) (*models.Item, error) {
	var item models.Item
	var createdAt, updatedAt time.Time

	var inStockDate time.Time
	var expiryDate sql.NullTime
	err := db.GetDB().QueryRow(
		"SELECT id, name, code, description, price, cost, quantity, in_stock_date, expiry_date, created_at, updated_at FROM items WHERE id = ?",
		id,
	).Scan(&item.ID, &item.Name, &item.Code, &item.Description, &item.Price, &item.Cost, &item.Quantity, &inStockDate, &expiryDate, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("item not found")
	}
	if err != nil {
		return nil, err
	}

	item.InStockDate = inStockDate
	if expiryDate.Valid {
		item.ExpiryDate = &expiryDate.Time
	}
	item.CreatedAt = createdAt
	item.UpdatedAt = updatedAt
	return &item, nil
}

func GetItemByCode(db Database, code string) (*models.Item, error) {
	var item models.Item
	var createdAt, updatedAt, inStockDate time.Time
	var expiryDate sql.NullTime

	err := db.GetDB().QueryRow(
		"SELECT id, name, code, description, price, cost, quantity, in_stock_date, expiry_date, created_at, updated_at FROM items WHERE code = ?",
		code,
	).Scan(&item.ID, &item.Name, &item.Code, &item.Description, &item.Price, &item.Cost, &item.Quantity, &inStockDate, &expiryDate, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("item not found")
	}
	if err != nil {
		return nil, err
	}

	item.InStockDate = inStockDate
	if expiryDate.Valid {
		item.ExpiryDate = &expiryDate.Time
	}
	item.CreatedAt = createdAt
	item.UpdatedAt = updatedAt
	return &item, nil
}

func SearchItems(db Database, query string) ([]models.Item, error) {
	rows, err := db.GetDB().Query(
		"SELECT id, name, code, description, price, cost, quantity, in_stock_date, expiry_date, created_at, updated_at FROM items WHERE name LIKE ? OR code LIKE ? ORDER BY name",
		"%"+query+"%", "%"+query+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		var createdAt, updatedAt, inStockDate time.Time
		var expiryDate sql.NullTime

		err := rows.Scan(&item.ID, &item.Name, &item.Code, &item.Description, &item.Price, &item.Cost, &item.Quantity, &inStockDate, &expiryDate, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		item.InStockDate = inStockDate
		if expiryDate.Valid {
			item.ExpiryDate = &expiryDate.Time
		}
		item.CreatedAt = createdAt
		item.UpdatedAt = updatedAt
		items = append(items, item)
	}

	return items, nil
}

func GetAllItems(db Database) ([]models.Item, error) {
	rows, err := db.GetDB().Query(
		"SELECT id, name, code, description, price, cost, quantity, in_stock_date, expiry_date, created_at, updated_at FROM items ORDER BY name",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		var createdAt, updatedAt, inStockDate time.Time
		var expiryDate sql.NullTime

		err := rows.Scan(&item.ID, &item.Name, &item.Code, &item.Description, &item.Price, &item.Cost, &item.Quantity, &inStockDate, &expiryDate, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		item.InStockDate = inStockDate
		if expiryDate.Valid {
			item.ExpiryDate = &expiryDate.Time
		}
		item.CreatedAt = createdAt
		item.UpdatedAt = updatedAt
		items = append(items, item)
	}

	return items, nil
}

func UpdateItem(db Database, id int, name, code, description string, price, cost float64, quantity int) error {
	_, err := db.GetDB().Exec(
		"UPDATE items SET name = ?, code = ?, description = ?, price = ?, cost = ?, quantity = ?, updated_at = ? WHERE id = ?",
		name, code, description, price, cost, quantity, time.Now(), id,
	)
	return err
}

func DeleteItem(db Database, id int) error {
	_, err := db.GetDB().Exec("DELETE FROM items WHERE id = ?", id)
	return err
}

func UpdateItemQuantity(db Database, id int, quantity int) error {
	_, err := db.GetDB().Exec(
		"UPDATE items SET quantity = ?, updated_at = ? WHERE id = ?",
		quantity, time.Now(), id,
	)
	return err
}

