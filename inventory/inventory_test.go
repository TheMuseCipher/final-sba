package inventory

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

type MockDB struct {
	db *sql.DB
}

func (m *MockDB) GetDB() *sql.DB {
	return m.db
}

func setupTestDB(t *testing.T) *MockDB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		code TEXT UNIQUE NOT NULL,
		description TEXT,
		price REAL NOT NULL,
		cost REAL NOT NULL DEFAULT 0,
		quantity INTEGER DEFAULT 0,
		in_stock_date DATETIME DEFAULT CURRENT_TIMESTAMP,
		expiry_date DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("Failed to create items table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE item_stock (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		item_id INTEGER NOT NULL,
		quantity INTEGER NOT NULL,
		in_stock_date DATETIME DEFAULT CURRENT_TIMESTAMP,
		expiry_date DATETIME
	)`)
	if err != nil {
		t.Fatalf("Failed to create item_stock table: %v", err)
	}

	return &MockDB{db: db}
}

func TestCreateItem(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.db.Close()

	item, err := CreateItem(mockDB, "Test Item", "CODE001", "A test item", 9.99, 5.00, 10)
	if err != nil {
		t.Fatalf("CreateItem failed: %v", err)
	}

	if item.Name != "Test Item" {
		t.Errorf("Expected name 'Test Item', got '%s'", item.Name)
	}
	if item.Code != "CODE001" {
		t.Errorf("Expected code 'CODE001', got '%s'", item.Code)
	}
	if item.Price != 9.99 {
		t.Errorf("Expected price 9.99, got %f", item.Price)
	}
	if item.Quantity != 10 {
		t.Errorf("Expected quantity 10, got %d", item.Quantity)
	}
}

func TestGetItemByID(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.db.Close()

	created, err := CreateItem(mockDB, "Find Me", "FIND001", "Item to find", 15.00, 10.00, 5)
	if err != nil {
		t.Fatalf("CreateItem failed: %v", err)
	}

	found, err := GetItemByID(mockDB, created.ID)
	if err != nil {
		t.Fatalf("GetItemByID failed: %v", err)
	}

	if found.Name != "Find Me" {
		t.Errorf("Expected name 'Find Me', got '%s'", found.Name)
	}
}

func TestGetItemByCode(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.db.Close()

	_, err := CreateItem(mockDB, "Barcode Item", "BAR123", "Item with barcode", 20.00, 12.00, 8)
	if err != nil {
		t.Fatalf("CreateItem failed: %v", err)
	}

	found, err := GetItemByCode(mockDB, "BAR123")
	if err != nil {
		t.Fatalf("GetItemByCode failed: %v", err)
	}

	if found.Name != "Barcode Item" {
		t.Errorf("Expected name 'Barcode Item', got '%s'", found.Name)
	}
}

func TestSearchItems(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.db.Close()

	CreateItem(mockDB, "Apple Juice", "AJ001", "Fresh apple juice", 3.50, 2.00, 20)
	CreateItem(mockDB, "Orange Juice", "OJ001", "Fresh orange juice", 3.50, 2.00, 15)
	CreateItem(mockDB, "Milk", "MK001", "Fresh milk", 2.00, 1.00, 30)

	results, err := SearchItems(mockDB, "Juice")
	if err != nil {
		t.Fatalf("SearchItems failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'Juice', got %d", len(results))
	}
}

func TestUpdateItemQuantity(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.db.Close()

	item, err := CreateItem(mockDB, "Stock Item", "ST001", "Item for stock test", 10.00, 6.00, 100)
	if err != nil {
		t.Fatalf("CreateItem failed: %v", err)
	}

	err = UpdateItemQuantity(mockDB, item.ID, 75)
	if err != nil {
		t.Fatalf("UpdateItemQuantity failed: %v", err)
	}

	updated, err := GetItemByID(mockDB, item.ID)
	if err != nil {
		t.Fatalf("GetItemByID failed: %v", err)
	}

	if updated.Quantity != 75 {
		t.Errorf("Expected quantity 75, got %d", updated.Quantity)
	}
}

func TestDeleteItem(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.db.Close()

	item, err := CreateItem(mockDB, "Delete Me", "DEL001", "Item to delete", 5.00, 3.00, 10)
	if err != nil {
		t.Fatalf("CreateItem failed: %v", err)
	}

	err = DeleteItem(mockDB, item.ID)
	if err != nil {
		t.Fatalf("DeleteItem failed: %v", err)
	}

	_, err = GetItemByID(mockDB, item.ID)
	if err == nil {
		t.Error("GetItemByID should fail after item is deleted")
	}
}

func TestGetLowStockItems(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.db.Close()

	CreateItem(mockDB, "Low Stock Item", "LOW001", "Only 5 left", 10.00, 5.00, 5)
	CreateItem(mockDB, "Medium Stock Item", "MED001", "Has 15", 10.00, 5.00, 15)
	CreateItem(mockDB, "High Stock Item", "HIGH001", "Has 100", 10.00, 5.00, 100)
	CreateItem(mockDB, "Very Low Stock", "VLOW001", "Only 2 left", 10.00, 5.00, 2)

	lowStockItems, err := GetLowStockItems(mockDB, 10)
	if err != nil {
		t.Fatalf("GetLowStockItems failed: %v", err)
	}

	if len(lowStockItems) != 2 {
		t.Errorf("Expected 2 low stock items, got %d", len(lowStockItems))
	}

	if len(lowStockItems) >= 2 {
		if lowStockItems[0].Quantity > lowStockItems[1].Quantity {
			t.Error("Low stock items should be sorted by quantity ascending")
		}
	}
}
