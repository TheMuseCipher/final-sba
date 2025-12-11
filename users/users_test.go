package users

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

	return &MockDB{db: db}
}

func TestCreateUser(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.db.Close()

	user, err := CreateUser(mockDB, "testuser", "password123", true, false, true)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}
	if !user.CanRead {
		t.Error("Expected CanRead to be true")
	}
	if user.CanTransaction {
		t.Error("Expected CanTransaction to be false")
	}
	if !user.CanRevenue {
		t.Error("Expected CanRevenue to be true")
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.db.Close()

	_, err := CreateUser(mockDB, "sameuser", "password1", true, true, true)
	if err != nil {
		t.Fatalf("First CreateUser failed: %v", err)
	}

	_, err = CreateUser(mockDB, "sameuser", "password2", false, false, false)
	if err == nil {
		t.Error("Expected error for duplicate username, got nil")
	}
}

func TestGetUserByID(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.db.Close()

	created, err := CreateUser(mockDB, "findme", "password", true, true, true)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	found, err := GetUserByID(mockDB, created.ID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}

	if found.Username != "findme" {
		t.Errorf("Expected username 'findme', got '%s'", found.Username)
	}
}

func TestUpdateUserPermissions(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.db.Close()

	user, err := CreateUser(mockDB, "updateme", "password", false, false, false)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	err = UpdateUserPermissions(mockDB, user.ID, true, true, true)
	if err != nil {
		t.Fatalf("UpdateUserPermissions failed: %v", err)
	}

	updated, err := GetUserByID(mockDB, user.ID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}

	if !updated.CanRead {
		t.Error("Expected CanRead to be true after update")
	}
	if !updated.CanTransaction {
		t.Error("Expected CanTransaction to be true after update")
	}
	if !updated.CanRevenue {
		t.Error("Expected CanRevenue to be true after update")
	}
}

func TestDeleteUser(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.db.Close()

	user, err := CreateUser(mockDB, "deleteme", "password", true, true, true)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	err = DeleteUser(mockDB, user.ID)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	_, err = GetUserByID(mockDB, user.ID)
	if err == nil {
		t.Error("GetUserByID should fail after user is deleted")
	}
}

func TestDeleteUser_CannotDeleteRootAdmin(t *testing.T) {
	mockDB := setupTestDB(t)
	defer mockDB.db.Close()

	_, err := mockDB.db.Exec(
		`INSERT INTO users (username, password_hash, is_root_admin, can_read, can_transaction, can_revenue)
		 VALUES ('admin', 'hash', 1, 1, 1, 1)`,
	)
	if err != nil {
		t.Fatalf("Failed to insert root admin: %v", err)
	}

	err = DeleteUser(mockDB, 1)
	if err == nil {
		t.Error("Should not be able to delete root admin")
	}
}
