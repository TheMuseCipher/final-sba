package database

import (
	"database/sql"

	_ "modernc.org/sqlite"

	"ims-go/auth"
)

type Database struct {
	db *sql.DB
}

func NewDatabase() (*Database, error) {
	db, err := sql.Open("sqlite", "ims.db")
	if err != nil {
		return nil, err
	}

	d := &Database{db: db}
	if err := d.initSchema(); err != nil {
		return nil, err
	}

	// Create root admin if it doesn't exist
	if err := d.ensureRootAdmin(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) GetDB() *sql.DB {
	return d.db
}

func (d *Database) initSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			is_root_admin INTEGER DEFAULT 0,
			can_read INTEGER DEFAULT 0,
			can_transaction INTEGER DEFAULT 0,
			can_revenue INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS items (
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
		)`,
		`CREATE TABLE IF NOT EXISTS item_stock (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			item_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL,
			in_stock_date DATETIME DEFAULT CURRENT_TIMESTAMP,
			expiry_date DATETIME,
			FOREIGN KEY (item_id) REFERENCES items(id)
		)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			total_amount REAL NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS transaction_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			transaction_id INTEGER NOT NULL,
			item_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL,
			price REAL NOT NULL,
			FOREIGN KEY (transaction_id) REFERENCES transactions(id),
			FOREIGN KEY (item_id) REFERENCES items(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_items_code ON items(code)`,
		`CREATE INDEX IF NOT EXISTS idx_items_name ON items(name)`,
	}

	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return err
		}
	}

	// Add new columns if they don't exist (for existing databases)
	migrationQueries := []string{
		`ALTER TABLE users ADD COLUMN can_revenue INTEGER DEFAULT 0`,
		`ALTER TABLE items ADD COLUMN cost REAL NOT NULL DEFAULT 0`,
		`ALTER TABLE items ADD COLUMN in_stock_date DATETIME DEFAULT CURRENT_TIMESTAMP`,
		`ALTER TABLE items ADD COLUMN expiry_date DATETIME`,
	}

	for _, query := range migrationQueries {
		d.db.Exec(query) // Ignore error if column already exists
	}

	return nil
}

func (d *Database) ensureRootAdmin() error {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM users WHERE is_root_admin = 1").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// Create default root admin: admin/admin
		hashedPassword, err := auth.HashPassword("admin")
		if err != nil {
			return err
		}

		_, err = d.db.Exec(
			"INSERT INTO users (username, password_hash, is_root_admin, can_read, can_transaction, can_revenue) VALUES (?, ?, 1, 1, 1, 1)",
			"admin", hashedPassword,
		)
		return err
	}

	return nil
}

// ResetDatabase deletes all data and reinitializes the database
func (d *Database) ResetDatabase() error {
	// Delete all data from all tables (in reverse order of dependencies)
	tables := []string{
		"transaction_items",
		"transactions",
		"item_stock",
		"items",
		"users",
	}

	for _, table := range tables {
		if _, err := d.db.Exec("DELETE FROM " + table); err != nil {
			return err
		}
	}

	// Reset auto-increment counters
	resetQueries := []string{
		"DELETE FROM sqlite_sequence WHERE name IN ('users', 'items', 'item_stock', 'transactions', 'transaction_items')",
	}

	for _, query := range resetQueries {
		d.db.Exec(query) // Ignore error if sequence doesn't exist
	}

	// Recreate root admin
	return d.ensureRootAdmin()
}
