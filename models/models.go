package models

import "time"

type User struct {
	ID             int
	Username       string
	IsRootAdmin    bool
	CanRead        bool
	CanTransaction bool
	CanRevenue     bool
	CreatedAt      time.Time
}

type Item struct {
	ID          int
	Name        string
	Code        string
	Description string
	Price       float64
	Cost        float64
	Quantity    int
	InStockDate time.Time
	ExpiryDate  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ItemStock struct {
	ID          int
	ItemID      int
	Quantity    int
	InStockDate time.Time
	ExpiryDate  *time.Time
}

type Transaction struct {
	ID          int
	UserID      int
	TotalAmount float64
	CreatedAt   time.Time
	Items       []TransactionItem
}

type TransactionItem struct {
	ID            int
	TransactionID int
	ItemID        int
	ItemName      string
	Quantity      int
	Price         float64
}

