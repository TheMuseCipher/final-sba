package transactions

import (
	"database/sql"
	"testing"

	"ims-go/models"

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

	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		is_root_admin INTEGER DEFAULT 0,
		can_read INTEGER DEFAULT 0,
		can_transaction INTEGER DEFAULT 0,
		can_revenue INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
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

	_, err = db.Exec(`CREATE TABLE transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		total_amount REAL NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("Failed to create transactions table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE transaction_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		transaction_id INTEGER NOT NULL,
		item_id INTEGER NOT NULL,
		quantity INTEGER NOT NULL,
		price REAL NOT NULL
	)`)
	if err != nil {
		t.Fatalf("Failed to create transaction_items table: %v", err)
	}

	_, err = db.Exec(`INSERT INTO users (username, password_hash, can_transaction) VALUES ('testuser', 'hash', 1)`)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = db.Exec(`INSERT INTO items (name, code, price, cost, quantity) VALUES ('Apple', 'APL001', 1.50, 1.00, 100)`)
	if err != nil {
		t.Fatalf("Failed to insert test item: %v", err)
	}

	_, err = db.Exec(`INSERT INTO items (name, code, price, cost, quantity) VALUES ('Banana', 'BAN001', 0.75, 0.50, 50)`)
	if err != nil {
		t.Fatalf("Failed to insert test item: %v", err)
	}

	return &MockDB{db: db}
}

func TestCreateTransaction(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.db.Close()

	items := []models.TransactionItem{
		{ItemID: 1, ItemName: "Apple", Quantity: 2, Price: 1.50},
		{ItemID: 2, ItemName: "Banana", Quantity: 3, Price: 0.75},
	}

	transaction, err := CreateTransaction(mockDB, 1, items)
	if err != nil {
		t.Fatalf("CreateTransaction failed: %v", err)
	}

	expectedTotal := 5.25
	if transaction.TotalAmount != expectedTotal {
		t.Errorf("Expected total %.2f, got %.2f", expectedTotal, transaction.TotalAmount)
	}

	if transaction.UserID != 1 {
		t.Errorf("Expected user ID 1, got %d", transaction.UserID)
	}
}

func TestCreateTransaction_ReducesStock(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.db.Close()

	items := []models.TransactionItem{
		{ItemID: 1, ItemName: "Apple", Quantity: 10, Price: 1.50},
	}

	_, err := CreateTransaction(mockDB, 1, items)
	if err != nil {
		t.Fatalf("CreateTransaction failed: %v", err)
	}

	var newQuantity int
	err = mockDB.db.QueryRow("SELECT quantity FROM items WHERE id = 1").Scan(&newQuantity)
	if err != nil {
		t.Fatalf("Failed to get item quantity: %v", err)
	}

	if newQuantity != 90 {
		t.Errorf("Expected quantity 90, got %d", newQuantity)
	}
}

func TestGetTransactionByID(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.db.Close()

	items := []models.TransactionItem{
		{ItemID: 1, ItemName: "Apple", Quantity: 5, Price: 1.50},
	}

	created, err := CreateTransaction(mockDB, 1, items)
	if err != nil {
		t.Fatalf("CreateTransaction failed: %v", err)
	}

	found, err := GetTransactionByID(mockDB, created.ID)
	if err != nil {
		t.Fatalf("GetTransactionByID failed: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, found.ID)
	}

	if len(found.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(found.Items))
	}
}
