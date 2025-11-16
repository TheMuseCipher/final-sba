package gui

import (
	"database/sql"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"ims-go/auth"
	"ims-go/database"
	"ims-go/models"
)

type RevenueItem struct {
	ItemID      int
	ItemName    string
	TotalRevenue float64
	QuantitySold int
}

func createRevenueTab(parent fyne.Window, appState *auth.AppState, user *models.User) *container.Scroll {
	// Revenue items list
	var revenueItems []RevenueItem
	var oldestItems []models.Item

	// Get revenue data
	revenueList := widget.NewList(
		func() int {
			return len(revenueItems)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(revenueItems) {
				item := revenueItems[id]
				box := obj.(*fyne.Container)
				box.Objects[0].(*widget.Label).SetText(item.ItemName)
				box.Objects[1].(*widget.Label).SetText(fmt.Sprintf("Sold: %d", item.QuantitySold))
				box.Objects[2].(*widget.Label).SetText(fmt.Sprintf("Revenue: $%.2f", item.TotalRevenue))
			}
		},
	)

	// Oldest items list
	oldestList := widget.NewList(
		func() int {
			return len(oldestItems)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(oldestItems) {
				item := oldestItems[id]
				box := obj.(*fyne.Container)
				box.Objects[0].(*widget.Label).SetText(item.Name)
				now := time.Now().Round(time.Minute)
				shelfLife := now.Sub(item.InStockDate.Round(time.Minute))
				days := int(shelfLife.Hours() / 24)
				box.Objects[1].(*widget.Label).SetText(fmt.Sprintf("On shelf: %d days", days))
				box.Objects[2].(*widget.Label).SetText(fmt.Sprintf("Qty: %d", item.Quantity))
			}
		},
	)

	refreshData := func() {
		// Get revenue data
		db := appState.GetDB().(*database.Database)
		rows, err := db.GetDB().Query(`
			SELECT 
				i.id,
				i.name,
				SUM((i.price - i.cost) * ti.quantity) as total_revenue,
				SUM(ti.quantity) as quantity_sold
			FROM items i
			JOIN transaction_items ti ON i.id = ti.item_id
			GROUP BY i.id, i.name
			ORDER BY total_revenue DESC
		`)
		if err == nil {
			defer rows.Close()
			revenueItems = []RevenueItem{}
			for rows.Next() {
				var item RevenueItem
				err := rows.Scan(&item.ItemID, &item.ItemName, &item.TotalRevenue, &item.QuantitySold)
				if err == nil {
					revenueItems = append(revenueItems, item)
				}
			}
		}

		// Get oldest items
		rows2, err := db.GetDB().Query(`
			SELECT id, name, code, description, price, cost, quantity, in_stock_date, expiry_date, created_at, updated_at
			FROM items
			WHERE quantity > 0
			ORDER BY in_stock_date ASC
		`)
		if err == nil {
			defer rows2.Close()
			oldestItems = []models.Item{}
			for rows2.Next() {
				var item models.Item
				var createdAt, updatedAt, inStockDate time.Time
				var expiryDate sql.NullTime
				err := rows2.Scan(&item.ID, &item.Name, &item.Code, &item.Description, &item.Price, &item.Cost, &item.Quantity, &inStockDate, &expiryDate, &createdAt, &updatedAt)
				if err == nil {
					item.InStockDate = inStockDate
					if expiryDate.Valid {
						item.ExpiryDate = &expiryDate.Time
					}
					item.CreatedAt = createdAt
					item.UpdatedAt = updatedAt
					oldestItems = append(oldestItems, item)
				}
			}
		}

		revenueList.Refresh()
		oldestList.Refresh()
	}

	refreshBtn := widget.NewButton("Refresh", refreshData)

	content := container.NewHSplit(
		container.NewBorder(
			container.NewVBox(
				widget.NewLabel("Top Revenue Items"),
				widget.NewSeparator(),
			),
			refreshBtn,
			nil,
			nil,
			container.NewScroll(revenueList),
		),
		container.NewBorder(
			container.NewVBox(
				widget.NewLabel("Oldest Items on Shelf"),
				widget.NewSeparator(),
			),
			nil,
			nil,
			nil,
			container.NewScroll(oldestList),
		),
	)
	content.SetOffset(0.5)

	refreshData()
	return container.NewScroll(content)
}

