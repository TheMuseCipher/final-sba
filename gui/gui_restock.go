package gui

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"ims-go/auth"
	"ims-go/database"
	"ims-go/inventory"
	"ims-go/models"
)

func showRestockDialog(parent fyne.Window, appState *auth.AppState, item *models.Item, onSuccess func()) {
	// Get current quantity
	db := appState.GetDB().(*database.Database)
	currentQty, err := inventory.GetItemQuantity(db, item.ID)
	if err != nil {
		dialog.ShowError(err, parent)
		return
	}

	restockQtyEntry := widget.NewEntry()
	restockQtyEntry.SetPlaceHolder("Quantity to add")
	expiryDateEntry := widget.NewEntry()
	expiryDateEntry.SetPlaceHolder("Expiry Date (YYYY-MM-DD, optional)")

	totalQtyLabel := widget.NewLabel(fmt.Sprintf("Total after restock: %d", currentQty))
	totalQtyLabel.TextStyle = fyne.TextStyle{Bold: true}

	updateTotal := func() {
		restockQty, err := strconv.Atoi(restockQtyEntry.Text)
		if err == nil && restockQty > 0 {
			totalQtyLabel.SetText(fmt.Sprintf("Total after restock: %d", currentQty+restockQty))
		} else {
			totalQtyLabel.SetText(fmt.Sprintf("Total after restock: %d", currentQty))
		}
	}

	restockQtyEntry.OnChanged = func(_ string) {
		updateTotal()
	}

	formContent := container.NewVBox(
		createStyledFormField("Current Quantity", widget.NewLabel(fmt.Sprintf("%d", currentQty))),
		createStyledFormField("Quantity to Add", restockQtyEntry),
		createStyledFormField("Expiry Date (optional)", expiryDateEntry),
		createStyledFormField("", totalQtyLabel),
	)

	onAction := func() {
		restockQty, err := strconv.Atoi(restockQtyEntry.Text)
		if err != nil || restockQty <= 0 {
			dialog.ShowError(fmt.Errorf("invalid quantity"), parent)
			return
		}

		var expiryDate *time.Time
		if expiryDateEntry.Text != "" {
			parsedDate, err := time.Parse("2006-01-02", expiryDateEntry.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid date format. Use YYYY-MM-DD"), parent)
				return
			}
			expiryDate = &parsedDate
		}

		err = inventory.RestockItem(db, item.ID, restockQty, expiryDate)
		if err != nil {
			dialog.ShowError(err, parent)
			return
		}

		showStyledInformation(parent, "Success", fmt.Sprintf("Item restocked successfully. New quantity: %d", currentQty+restockQty))
		onSuccess()
	}

	showStyledDialog(parent, "Restock Item", formContent, "Restock", onAction, nil)
}

