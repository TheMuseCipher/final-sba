package gui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"ims-go/auth"
	"ims-go/database"
	"ims-go/models"
	transactionsPkg "ims-go/transactions"
)

func createTransactionLogTab(parent fyne.Window, appState *auth.AppState, user *models.User) *container.Scroll {
	var transactions []models.Transaction

	list := widget.NewList(
		func() int {
			var err error
			db := appState.GetDB().(*database.Database)
			txns, err := transactionsPkg.GetRecentTransactions(db, 1000)
			if err == nil {
				transactions = txns
			}
			if err != nil {
				return 0
			}
			return len(transactions)
		},
		func() fyne.CanvasObject {
			eyeBtn := widget.NewButtonWithIcon("", theme.VisibilityIcon(), nil)
			eyeBtn.Importance = widget.LowImportance
			idLabel := widget.NewLabel("")
			dateLabel := widget.NewLabel("")
			totalLabel := widget.NewLabel("")
			return container.NewHBox(
				container.NewBorder(nil, nil, nil, nil, idLabel),
				container.NewBorder(nil, nil, nil, nil, dateLabel),
				container.NewBorder(nil, nil, nil, nil, totalLabel),
				container.NewBorder(nil, nil, nil, nil, eyeBtn),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(transactions) {
				txn := transactions[id]
				box := obj.(*fyne.Container)
				idLabel := box.Objects[0].(*fyne.Container).Objects[0].(*widget.Label)
				idLabel.SetText(fmt.Sprintf("#%d", txn.ID))
				idLabel.Resize(fyne.NewSize(80, idLabel.MinSize().Height))
				dateLabel := box.Objects[1].(*fyne.Container).Objects[0].(*widget.Label)
				dateLabel.SetText(txn.CreatedAt.Format("2006-01-02 15:04:05"))
				dateLabel.Resize(fyne.NewSize(180, dateLabel.MinSize().Height))
				totalLabel := box.Objects[2].(*fyne.Container).Objects[0].(*widget.Label)
				totalLabel.SetText(fmt.Sprintf("$%.2f", txn.TotalAmount))
				totalLabel.Resize(fyne.NewSize(120, totalLabel.MinSize().Height))
				btn := box.Objects[3].(*fyne.Container).Objects[0].(*widget.Button)
				btn.OnTapped = func() {
					showTransactionDetails(parent, appState, &txn)
				}
			}
		},
	)

	// Handle double-click via OnSelected with a timer
	var lastSelectedID widget.ListItemID = -1
	var lastSelectedTime int64 = 0
	list.OnSelected = func(id widget.ListItemID) {
		now := time.Now().UnixNano()
		if id == lastSelectedID && (now-lastSelectedTime) < 500000000 { // 500ms double-click window
			if id < len(transactions) {
				showTransactionDetails(parent, appState, &transactions[id])
			}
		}
		lastSelectedID = id
		lastSelectedTime = now
	}

	refreshBtn := widget.NewButton("Refresh", func() {
		list.Refresh()
	})

	// Column headers for transaction log with fixed widths
	idHeader := widget.NewLabel("ID")
	idHeader.TextStyle = fyne.TextStyle{Bold: true}
	idHeader.Resize(fyne.NewSize(80, idHeader.MinSize().Height))
	dateHeader := widget.NewLabel("Date")
	dateHeader.TextStyle = fyne.TextStyle{Bold: true}
	dateHeader.Resize(fyne.NewSize(180, dateHeader.MinSize().Height))
	totalHeader := widget.NewLabel("Total")
	totalHeader.TextStyle = fyne.TextStyle{Bold: true}
	totalHeader.Resize(fyne.NewSize(120, totalHeader.MinSize().Height))
	headerRow := container.NewHBox(
		container.NewBorder(nil, nil, nil, nil, idHeader),
		container.NewBorder(nil, nil, nil, nil, dateHeader),
		container.NewBorder(nil, nil, nil, nil, totalHeader),
		container.NewBorder(nil, nil, nil, nil, widget.NewLabel("")),
	)

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Transaction Log"),
			widget.NewSeparator(),
			headerRow,
			widget.NewSeparator(),
		),
		refreshBtn,
		nil,
		nil,
		list,
	)

	list.Refresh()
	return container.NewScroll(content)
}

func showTransactionDetails(parent fyne.Window, appState *auth.AppState, txn *models.Transaction) {
	// Reload transaction with items
	db := appState.GetDB().(*database.Database)
	fullTxn, err := transactionsPkg.GetTransactionByID(db, txn.ID)
	if err != nil {
		dialog.ShowError(err, parent)
		return
	}

	// Create detail window
	detailWindow := fyne.CurrentApp().NewWindow(fmt.Sprintf("Transaction #%d Details", fullTxn.ID))
	detailWindow.Resize(fyne.NewSize(600, 500))
	detailWindow.CenterOnScreen()

	// Create highlighted header section
	transactionIDLabel := widget.NewLabel(fmt.Sprintf("Transaction #%d", fullTxn.ID))
	transactionIDLabel.TextStyle = fyne.TextStyle{Bold: true}
	dateLabel := widget.NewLabel(fmt.Sprintf("Date: %s", fullTxn.CreatedAt.Format("2006-01-02 15:04:05")))
	dateLabel.TextStyle = fyne.TextStyle{Bold: true}
	totalLabel := widget.NewLabel(fmt.Sprintf("Total: $%.2f", fullTxn.TotalAmount))
	totalLabel.TextStyle = fyne.TextStyle{Bold: true}
	
	infoSection := container.NewVBox(
		container.NewPadded(transactionIDLabel),
		container.NewPadded(dateLabel),
		container.NewPadded(totalLabel),
	)

	// Column headers for items with fixed widths
	itemNameHeader := widget.NewLabel("Item Name")
	itemNameHeader.TextStyle = fyne.TextStyle{Bold: true}
	itemNameHeader.Resize(fyne.NewSize(200, itemNameHeader.MinSize().Height))
	qtyHeader := widget.NewLabel("Quantity")
	qtyHeader.TextStyle = fyne.TextStyle{Bold: true}
	qtyHeader.Resize(fyne.NewSize(100, qtyHeader.MinSize().Height))
	subtotalHeader := widget.NewLabel("Subtotal")
	subtotalHeader.TextStyle = fyne.TextStyle{Bold: true}
	subtotalHeader.Resize(fyne.NewSize(120, subtotalHeader.MinSize().Height))
	itemsHeaderRow := container.NewHBox(
		container.NewBorder(nil, nil, nil, nil, itemNameHeader),
		container.NewBorder(nil, nil, nil, nil, qtyHeader),
		container.NewBorder(nil, nil, nil, nil, subtotalHeader),
	)

	itemsList := widget.NewList(
		func() int {
			return len(fullTxn.Items)
		},
		func() fyne.CanvasObject {
			itemLabel := widget.NewLabel("")
			qtyLabel := widget.NewLabel("")
			subtotalLabel := widget.NewLabel("")
			return container.NewHBox(
				container.NewBorder(nil, nil, nil, nil, itemLabel),
				container.NewBorder(nil, nil, nil, nil, qtyLabel),
				container.NewBorder(nil, nil, nil, nil, subtotalLabel),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(fullTxn.Items) {
				item := fullTxn.Items[id]
				box := obj.(*fyne.Container)
				itemLabel := box.Objects[0].(*fyne.Container).Objects[0].(*widget.Label)
				itemLabel.SetText(item.ItemName)
				itemLabel.Resize(fyne.NewSize(200, itemLabel.MinSize().Height))
				qtyLabel := box.Objects[1].(*fyne.Container).Objects[0].(*widget.Label)
				qtyLabel.SetText(fmt.Sprintf("%d", item.Quantity))
				qtyLabel.Resize(fyne.NewSize(100, qtyLabel.MinSize().Height))
				subtotalLabel := box.Objects[2].(*fyne.Container).Objects[0].(*widget.Label)
				subtotalLabel.SetText(fmt.Sprintf("$%.2f", item.Price*float64(item.Quantity)))
				subtotalLabel.Resize(fyne.NewSize(120, subtotalLabel.MinSize().Height))
			}
		},
	)

	closeBtn := widget.NewButton("Close", func() {
		detailWindow.Close()
	})

	itemsSection := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Items:"),
			widget.NewSeparator(),
			itemsHeaderRow,
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		container.NewScroll(itemsList),
	)

	content := container.NewBorder(
		container.NewVBox(
			container.NewPadded(infoSection),
			widget.NewSeparator(),
		),
		container.NewPadded(closeBtn),
		nil,
		nil,
		container.NewPadded(itemsSection),
	)

	detailWindow.SetContent(content)
	detailWindow.Show()
}

