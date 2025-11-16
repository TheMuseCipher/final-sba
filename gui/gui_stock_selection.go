package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"ims-go/auth"
	"ims-go/models"
)

func showItemStockSelectionDialog(parent fyne.Window, appState *auth.AppState, item *models.Item, batches []models.ItemStock, onSelect func(*models.ItemStock)) {
	selectionWindow := fyne.CurrentApp().NewWindow("Select Stock Batch")
	selectionWindow.Resize(fyne.NewSize(500, 400))
	selectionWindow.CenterOnScreen()

	list := widget.NewList(
		func() int {
			return len(batches)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(batches) {
				batch := batches[id]
				box := obj.(*fyne.Container)
				box.Objects[0].(*widget.Label).SetText(fmt.Sprintf("Qty: %d", batch.Quantity))
				if batch.ExpiryDate != nil {
					box.Objects[1].(*widget.Label).SetText(fmt.Sprintf("Expires: %s", batch.ExpiryDate.Format("2006-01-02")))
				} else {
					box.Objects[1].(*widget.Label).SetText("No expiry")
				}
				box.Objects[2].(*widget.Label).SetText(fmt.Sprintf("In stock: %s", batch.InStockDate.Format("2006-01-02")))
			}
		},
	)

	var selectedBatchID widget.ListItemID = -1
	list.OnSelected = func(id widget.ListItemID) {
		selectedBatchID = id
	}

	selectBtn := widget.NewButton("Select", func() {
		if selectedBatchID >= 0 && selectedBatchID < len(batches) {
			onSelect(&batches[selectedBatchID])
			selectionWindow.Close()
		}
	})

	cancelBtn := widget.NewButton("Cancel", func() {
		selectionWindow.Close()
	})

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel(fmt.Sprintf("Select stock batch for: %s", item.Name)),
			widget.NewSeparator(),
		),
		container.NewHBox(selectBtn, cancelBtn),
		nil,
		nil,
		container.NewScroll(list),
	)

	selectionWindow.SetContent(content)
	selectionWindow.Show()
}

