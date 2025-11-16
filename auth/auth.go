package auth

import (
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	
	"ims-go/models"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type AppState struct {
	db   interface {
		GetDB() *sql.DB
		ResetDatabase() error
	}
	user *models.User
}

func NewAppState(db interface {
	GetDB() *sql.DB
	ResetDatabase() error
}) *AppState {
	return &AppState{db: db}
}

func (a *AppState) Authenticate(username, password string) (*models.User, error) {
	var id int
	var usernameDB, passwordHash string
	var isRootAdmin, canRead, canTransaction, canRevenue int
	var createdAt time.Time

	err := a.db.GetDB().QueryRow(
		"SELECT id, username, password_hash, is_root_admin, can_read, can_transaction, can_revenue, created_at FROM users WHERE username = ?",
		username,
	).Scan(&id, &usernameDB, &passwordHash, &isRootAdmin, &canRead, &canTransaction, &canRevenue, &createdAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("invalid credentials")
	}
	if err != nil {
		return nil, err
	}

	if !CheckPasswordHash(password, passwordHash) {
		return nil, errors.New("invalid credentials")
	}

	user := &models.User{
		ID:             id,
		Username:       usernameDB,
		IsRootAdmin:    isRootAdmin == 1,
		CanRead:        canRead == 1,
		CanTransaction: canTransaction == 1,
		CanRevenue:     canRevenue == 1,
		CreatedAt:       createdAt,
	}

	a.user = user
	return user, nil
}

func (a *AppState) GetCurrentUser() *models.User {
	return a.user
}

func (a *AppState) SetUser(user *models.User) {
	a.user = user
}

func (a *AppState) GetDB() interface {
	GetDB() *sql.DB
	ResetDatabase() error
} {
	return a.db
}

