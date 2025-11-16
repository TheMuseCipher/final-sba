package users

import (
	"database/sql"
	"errors"
	"time"

	"ims-go/auth"
	"ims-go/models"
)

type Database interface {
	GetDB() *sql.DB
}

func CreateUser(db Database, username, password string, canRead, canTransaction, canRevenue bool) (*models.User, error) {
	// Check if username already exists
	var count int
	err := db.GetDB().QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("username already exists")
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	var canReadInt, canTransactionInt, canRevenueInt int
	if canRead {
		canReadInt = 1
	}
	if canTransaction {
		canTransactionInt = 1
	}
	if canRevenue {
		canRevenueInt = 1
	}

	result, err := db.GetDB().Exec(
		"INSERT INTO users (username, password_hash, can_read, can_transaction, can_revenue) VALUES (?, ?, ?, ?, ?)",
		username, hashedPassword, canReadInt, canTransactionInt, canRevenueInt,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetUserByID(db, int(id))
}

func GetUserByID(db Database, id int) (*models.User, error) {
	var user models.User
	var createdAt time.Time
	var isRootAdmin, canRead, canTransaction, canRevenue int

	err := db.GetDB().QueryRow(
		"SELECT id, username, is_root_admin, can_read, can_transaction, can_revenue, created_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &isRootAdmin, &canRead, &canTransaction, &canRevenue, &createdAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	user.IsRootAdmin = isRootAdmin == 1
	user.CanRead = canRead == 1
	user.CanTransaction = canTransaction == 1
	user.CanRevenue = canRevenue == 1
	user.CreatedAt = createdAt
	return &user, nil
}

func GetAllUsers(db Database) ([]models.User, error) {
	rows, err := db.GetDB().Query(
		"SELECT id, username, is_root_admin, can_read, can_transaction, can_revenue, created_at FROM users ORDER BY username",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var createdAt time.Time
		var isRootAdmin, canRead, canTransaction, canRevenue int

		err := rows.Scan(&user.ID, &user.Username, &isRootAdmin, &canRead, &canTransaction, &canRevenue, &createdAt)
		if err != nil {
			return nil, err
		}

		user.IsRootAdmin = isRootAdmin == 1
		user.CanRead = canRead == 1
		user.CanTransaction = canTransaction == 1
		user.CanRevenue = canRevenue == 1
		user.CreatedAt = createdAt
		users = append(users, user)
	}

	return users, nil
}

func UpdateUserPermissions(db Database, id int, canRead, canTransaction, canRevenue bool) error {
	var canReadInt, canTransactionInt, canRevenueInt int
	if canRead {
		canReadInt = 1
	}
	if canTransaction {
		canTransactionInt = 1
	}
	if canRevenue {
		canRevenueInt = 1
	}

	_, err := db.GetDB().Exec(
		"UPDATE users SET can_read = ?, can_transaction = ?, can_revenue = ? WHERE id = ?",
		canReadInt, canTransactionInt, canRevenueInt, id,
	)
	return err
}

func DeleteUser(db Database, id int) error {
	// Prevent deleting root admin
	var isRootAdmin int
	err := db.GetDB().QueryRow("SELECT is_root_admin FROM users WHERE id = ?", id).Scan(&isRootAdmin)
	if err != nil {
		return err
	}
	if isRootAdmin == 1 {
		return errors.New("cannot delete root admin")
	}

	_, err = db.GetDB().Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func UpdateUserPassword(db Database, id int, newPassword string) error {
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	_, err = db.GetDB().Exec("UPDATE users SET password_hash = ? WHERE id = ?", hashedPassword, id)
	return err
}

