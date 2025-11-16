package gui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"ims-go/auth"
	"ims-go/barcode"
	"ims-go/database"
	"ims-go/inventory"
	"ims-go/models"
	transactionsPkg "ims-go/transactions"
)

func createTransactionTab(parent fyne.Window, appState *auth.AppState, user *models.User) *container.Scroll {
	// Transaction items
	var transactionItems []models.TransactionItem
	var totalAmount float64

	// Transaction items list (declare early)
	var itemList *widget.List
	totalLabel := widget.NewLabel("Total: $0.00")
	totalLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Code entry for barcode/QR scanning
	codeEntry := widget.NewEntry()
	codeEntry.SetPlaceHolder("Enter or scan barcode/QR code...")
	codeEntry.OnSubmitted = func(code string) {
		code = strings.TrimSpace(code)
		if code == "" {
			return
		}

		db := appState.GetDB().(*database.Database)
		item, err := inventory.GetItemByCode(db, code)
		if err != nil {
			// Item not found - ask to add it
			dialog.ShowConfirm("Item Not Found", 
				fmt.Sprintf("Item with code '%s' not found. Would you like to add it to inventory?", code),
				func(confirmed bool) {
					if confirmed {
						showAddItemFromTransactionDialog(parent, appState, code, func(newItem *models.Item) {
							// Add to transaction after creating
							addItemToTransaction(newItem, 1, &transactionItems, &totalAmount, itemList, totalLabel)
							codeEntry.SetText("")
						})
					} else {
						codeEntry.SetText("")
					}
				}, parent)
			return
		}

		// Item found - check for different expiry dates
		batches, err := inventory.GetItemStockBatches(db, item.ID)
		if err == nil && len(batches) > 1 {
			// Check if batches have different expiry dates
			hasDifferentExpiry := false
			if len(batches) > 0 {
				firstExpiry := batches[0].ExpiryDate
				for _, batch := range batches[1:] {
					if (firstExpiry == nil) != (batch.ExpiryDate == nil) ||
						(firstExpiry != nil && batch.ExpiryDate != nil && !firstExpiry.Equal(*batch.ExpiryDate)) {
						hasDifferentExpiry = true
						break
					}
				}
			}
			
			if hasDifferentExpiry {
				showItemStockSelectionDialog(parent, appState, item, batches, func(selectedBatch *models.ItemStock) {
					addItemToTransaction(item, 1, &transactionItems, &totalAmount, itemList, totalLabel)
				})
				codeEntry.SetText("")
				return
			}
		}

		// Item found - add to transaction
		addItemToTransaction(item, 1, &transactionItems, &totalAmount, itemList, totalLabel)
		codeEntry.SetText("")
	}

	// Search entry
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search by name or code...")
		var searchResults []models.Item
	var searchList *widget.List

	// Create search list first with aligned columns
	searchList = widget.NewList(
		func() int {
			return len(searchResults)
		},
		func() fyne.CanvasObject {
			addBtn := widget.NewButton("Add", nil)
			addBtn.Importance = widget.LowImportance
			nameLabel := widget.NewLabel("")
			codeLabel := widget.NewLabel("")
			priceLabel := widget.NewLabel("")
			return container.NewHBox(
				container.NewBorder(nil, nil, nil, nil, nameLabel),
				container.NewBorder(nil, nil, nil, nil, codeLabel),
				container.NewBorder(nil, nil, nil, nil, priceLabel),
				container.NewBorder(nil, nil, nil, nil, addBtn),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(searchResults) {
				item := searchResults[id]
				box := obj.(*fyne.Container)
				nameLabel := box.Objects[0].(*fyne.Container).Objects[0].(*widget.Label)
				nameLabel.SetText(item.Name)
				nameLabel.Resize(fyne.NewSize(200, nameLabel.MinSize().Height))
				codeLabel := box.Objects[1].(*fyne.Container).Objects[0].(*widget.Label)
				codeLabel.SetText(item.Code)
				codeLabel.Resize(fyne.NewSize(150, codeLabel.MinSize().Height))
				priceLabel := box.Objects[2].(*fyne.Container).Objects[0].(*widget.Label)
				priceLabel.SetText(fmt.Sprintf("$%.2f", item.Price))
				priceLabel.Resize(fyne.NewSize(100, priceLabel.MinSize().Height))
				btn := box.Objects[3].(*fyne.Container).Objects[0].(*widget.Button)
				btn.OnTapped = func() {
					// Check for different expiry dates
					db := appState.GetDB().(*database.Database)
					batches, err := inventory.GetItemStockBatches(db, item.ID)
					if err == nil && len(batches) > 1 {
						hasDifferentExpiry := false
						if len(batches) > 0 {
							firstExpiry := batches[0].ExpiryDate
							for _, batch := range batches[1:] {
								if (firstExpiry == nil) != (batch.ExpiryDate == nil) ||
									(firstExpiry != nil && batch.ExpiryDate != nil && !firstExpiry.Equal(*batch.ExpiryDate)) {
									hasDifferentExpiry = true
									break
								}
							}
						}
						
						if hasDifferentExpiry {
							showItemStockSelectionDialog(parent, appState, &item, batches, func(selectedBatch *models.ItemStock) {
								addItemToTransaction(&item, 1, &transactionItems, &totalAmount, itemList, totalLabel)
							})
							return
						}
					}
					// Add item to transaction (will increment if already exists)
					addItemToTransaction(&item, 1, &transactionItems, &totalAmount, itemList, totalLabel)
				}
			}
		},
	)

	// Load all items by default
	loadSearchResults := func() {
		query := strings.TrimSpace(searchEntry.Text)
		var items []models.Item
		var err error

		db := appState.GetDB().(*database.Database)
		if query == "" {
			items, err = inventory.GetAllItems(db)
		} else {
			items, err = inventory.SearchItems(db, query)
		}

		if err == nil {
			searchResults = items
			searchList.Refresh()
		}
	}

	searchEntry.OnChanged = func(query string) {
		loadSearchResults()
	}

	// Load all items initially
	loadSearchResults()

	// Transaction items list with aligned columns
	itemList = widget.NewList(
		func() int {
			return len(transactionItems)
		},
		func() fyne.CanvasObject {
			// Create all widgets
			itemLabel := widget.NewLabel("")
			qtyEntry := widget.NewEntry()
			priceLabel := widget.NewLabel("")
			removeBtn := widget.NewButton("Remove", nil)
			
			// Create fixed-width containers for each column
			// Item name column: 200px - set width during creation
			itemLabelContainer := container.NewBorder(nil, nil, nil, nil, itemLabel)
			itemLabelContainer.Resize(fyne.NewSize(200, 0))
			
			// Quantity entry column: 80px - constrain entry to prevent expansion
			qtyContainer := container.NewBorder(nil, nil, nil, nil, qtyEntry)
			qtyContainer.Resize(fyne.NewSize(80, 0))
			
			// Price column: 100px
			priceLabelContainer := container.NewBorder(nil, nil, nil, nil, priceLabel)
			priceLabelContainer.Resize(fyne.NewSize(100, 0))
			
			// Remove button - no fixed width, let it size naturally
			removeBtnContainer := container.NewBorder(nil, nil, nil, nil, removeBtn)
			
			// Use HBox with fixed-width containers
			return container.NewHBox(
				itemLabelContainer,
				qtyContainer,
				priceLabelContainer,
				removeBtnContainer,
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(transactionItems) {
				ti := transactionItems[id]
				box := obj.(*fyne.Container)
				
				// Update item label
				itemLabelContainer := box.Objects[0].(*fyne.Container)
				itemLabel := itemLabelContainer.Objects[0].(*widget.Label)
				itemLabel.SetText(ti.ItemName)
				
				// Update quantity entry - get the container and entry
				qtyContainer := box.Objects[1].(*fyne.Container)
				qtyEntry := qtyContainer.Objects[0].(*widget.Entry)
				
				// Store the current item ID to avoid closure issues
				currentID := id
				
				// Update entry text - just update the text, don't resize
				qtyEntry.SetText(fmt.Sprintf("%d", ti.Quantity))
				
				// Update price label
				priceLabelContainer := box.Objects[2].(*fyne.Container)
				priceLabel := priceLabelContainer.Objects[0].(*widget.Label)
				
				// Function to update price when quantity changes
				updatePrice := func(qty int) {
					priceLabel.SetText(fmt.Sprintf("$%.2f", ti.Price*float64(qty)))
					totalAmount = 0
					for _, item := range transactionItems {
						totalAmount += item.Price * float64(item.Quantity)
					}
					totalLabel.SetText(fmt.Sprintf("Total: $%.2f", totalAmount))
				}
				
				// Set initial price
				priceLabel.SetText(fmt.Sprintf("$%.2f", ti.Price*float64(ti.Quantity)))
				
				// Set up quantity entry change handler
				qtyEntry.OnChanged = func(text string) {
					newQty, err := strconv.Atoi(text)
					if err == nil && newQty > 0 && currentID < len(transactionItems) {
						transactionItems[currentID].Quantity = newQty
						updatePrice(newQty)
						// Don't call Refresh() here as it can cause the field to disappear
					}
				}
				
				// Set up remove button
				removeBtnContainer := box.Objects[3].(*fyne.Container)
				btn := removeBtnContainer.Objects[0].(*widget.Button)
				btn.OnTapped = func() {
					if currentID < len(transactionItems) {
						// Remove item
						transactionItems = append(transactionItems[:currentID], transactionItems[currentID+1:]...)
						totalAmount = 0
						for _, item := range transactionItems {
							totalAmount += item.Price * float64(item.Quantity)
						}
						itemList.Refresh()
						totalLabel.SetText(fmt.Sprintf("Total: $%.2f", totalAmount))
					}
				}
			}
		},
	)

	// Buttons
	clearBtn := widget.NewButton("Clear Transaction", func() {
		transactionItems = []models.TransactionItem{}
		totalAmount = 0
		itemList.Refresh()
		totalLabel.SetText("Total: $0.00")
	})

	completeBtn := widget.NewButton("Complete Transaction", func() {
		if len(transactionItems) == 0 {
			dialog.ShowInformation("Empty Transaction", "Please add items to the transaction", parent)
			return
		}

		user := appState.GetCurrentUser()
		if user == nil {
			dialog.ShowError(fmt.Errorf("user not authenticated"), parent)
			return
		}

		db := appState.GetDB().(*database.Database)
		_, err := transactionsPkg.CreateTransaction(db, user.ID, transactionItems)
		if err != nil {
			dialog.ShowError(err, parent)
			return
		}

		showStyledInformation(parent, "Success", fmt.Sprintf("Transaction completed. Total: $%.2f", totalAmount))
		transactionItems = []models.TransactionItem{}
		totalAmount = 0
		itemList.Refresh()
		totalLabel.SetText("Total: $0.00")
	})

	// File upload button for barcode/QR image
	uploadBtn := widget.NewButton("Upload Barcode/QR Image", func() {
		showFilePickerWithMemory(parent, func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			// Read image and decode barcode/QR
			code, err := barcode.DecodeBarcodeFromImage(reader.URI().Path())
			if err != nil {
				dialog.ShowError(fmt.Errorf("failed to decode barcode/QR: %v", err), parent)
				return
			}

			// Process the code
			codeEntry.SetText(code)
			codeEntry.OnSubmitted(code)
		})
	})

	// Column headers for search results with fixed widths
	nameHeader := widget.NewLabel("Name")
	nameHeader.TextStyle = fyne.TextStyle{Bold: true}
	nameHeader.Resize(fyne.NewSize(200, nameHeader.MinSize().Height))
	codeHeader := widget.NewLabel("Code")
	codeHeader.TextStyle = fyne.TextStyle{Bold: true}
	codeHeader.Resize(fyne.NewSize(150, codeHeader.MinSize().Height))
	priceHeader := widget.NewLabel("Price")
	priceHeader.TextStyle = fyne.TextStyle{Bold: true}
	priceHeader.Resize(fyne.NewSize(100, priceHeader.MinSize().Height))
	searchHeaderRow := container.NewHBox(
		container.NewBorder(nil, nil, nil, nil, nameHeader),
		container.NewBorder(nil, nil, nil, nil, codeHeader),
		container.NewBorder(nil, nil, nil, nil, priceHeader),
		container.NewBorder(nil, nil, nil, nil, widget.NewLabel("")),
	)

	// Layout
	leftPanel := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Scan or Enter Code:"),
			codeEntry,
			uploadBtn,
			widget.NewSeparator(),
			widget.NewLabel("Search Items:"),
			searchEntry,
			widget.NewSeparator(),
			searchHeaderRow,
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		container.NewScroll(searchList),
	)

	// Column headers for transaction items with fixed widths
	itemHeader := widget.NewLabel("Item")
	itemHeader.TextStyle = fyne.TextStyle{Bold: true}
	itemHeader.Resize(fyne.NewSize(200, itemHeader.MinSize().Height))
	qtyHeader := widget.NewLabel("Qty")
	qtyHeader.TextStyle = fyne.TextStyle{Bold: true}
	qtyHeader.Resize(fyne.NewSize(80, qtyHeader.MinSize().Height))
	subtotalHeader := widget.NewLabel("Subtotal")
	subtotalHeader.TextStyle = fyne.TextStyle{Bold: true}
	subtotalHeader.Resize(fyne.NewSize(100, subtotalHeader.MinSize().Height))
	transactionHeaderRow := container.NewHBox(
		container.NewBorder(nil, nil, nil, nil, itemHeader),
		container.NewBorder(nil, nil, nil, nil, qtyHeader),
		container.NewBorder(nil, nil, nil, nil, subtotalHeader),
		container.NewBorder(nil, nil, nil, nil, widget.NewLabel("")),
	)

	rightPanel := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Transaction Items:"),
			widget.NewSeparator(),
			transactionHeaderRow,
			widget.NewSeparator(),
		),
		container.NewVBox(
			totalLabel,
			widget.NewSeparator(),
			container.NewHBox(clearBtn, completeBtn),
		),
		nil,
		nil,
		container.NewScroll(itemList),
	)

	content := container.NewHSplit(leftPanel, rightPanel)
	content.SetOffset(0.4)

	return container.NewScroll(content)
}

func addItemToTransaction(item *models.Item, quantity int, transactionItems *[]models.TransactionItem, totalAmount *float64, itemList *widget.List, totalLabel *widget.Label) {
	// Check if item already in transaction
	for i, ti := range *transactionItems {
		if ti.ItemID == item.ID {
			(*transactionItems)[i].Quantity += quantity
			*totalAmount = 0
			for _, tItem := range *transactionItems {
				*totalAmount += tItem.Price * float64(tItem.Quantity)
			}
			itemList.Refresh()
			totalLabel.SetText(fmt.Sprintf("Total: $%.2f", *totalAmount))
			return
		}
	}

	// Add new item
	*transactionItems = append(*transactionItems, models.TransactionItem{
		ItemID:   item.ID,
		ItemName: item.Name,
		Quantity: quantity,
		Price:    item.Price,
	})

	*totalAmount = 0
	for _, tItem := range *transactionItems {
		*totalAmount += tItem.Price * float64(tItem.Quantity)
	}
	itemList.Refresh()
	totalLabel.SetText(fmt.Sprintf("Total: $%.2f", *totalAmount))
}

func showAddItemFromTransactionDialog(parent fyne.Window, appState *auth.AppState, code string, onSuccess func(*models.Item)) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Item Name")
	descEntry := widget.NewMultiLineEntry()
	descEntry.SetPlaceHolder("Description")
	priceEntry := widget.NewEntry()
	priceEntry.SetPlaceHolder("Price")
	costEntry := widget.NewEntry()
	costEntry.SetPlaceHolder("Cost")
	quantityEntry := widget.NewEntry()
	quantityEntry.SetPlaceHolder("Quantity")
	quantityEntry.SetText("0")

	formContent := container.NewVBox(
		createStyledFormField("Name", nameEntry),
		createStyledFormField("Code", widget.NewLabel(code)),
		createStyledFormField("Description", descEntry),
		createStyledFormField("Price", priceEntry),
		createStyledFormField("Cost", costEntry),
		createStyledFormField("Quantity", quantityEntry),
	)

	onAction := func() {
		price, err := strconv.ParseFloat(priceEntry.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid price"), parent)
			return
		}

		cost, err := strconv.ParseFloat(costEntry.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid cost"), parent)
			return
		}

		quantity, err := strconv.Atoi(quantityEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid quantity"), parent)
			return
		}

		db := appState.GetDB().(*database.Database)
		item, err := inventory.CreateItem(db, nameEntry.Text, code, descEntry.Text, price, cost, quantity)
		if err != nil {
			dialog.ShowError(err, parent)
			return
		}

		showStyledInformation(parent, "Success", "Item added successfully")
		onSuccess(item)
	}

	showStyledDialog(parent, "Add New Item", formContent, "Add", onAction, nil)
}

