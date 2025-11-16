package gui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"ims-go/auth"
	"ims-go/database"
	"ims-go/inventory"
	"ims-go/models"
)

func ShowMainWindow(myApp fyne.App, appState *auth.AppState, user *models.User) {
	mainWindow := myApp.NewWindow("Inventory Management System")
	mainWindow.Resize(fyne.NewSize(1000, 700))
	mainWindow.CenterOnScreen()

	// Create tabs based on user permissions
	tabs := container.NewAppTabs()

	// Inventory tab (if user can read)
	if user.CanRead || user.IsRootAdmin {
		tabs.Append(&container.TabItem{Text: "Inventory", Content: createInventoryTab(mainWindow, appState, user)})
	}

	// Transaction mode (if user has transaction permission)
	if user.CanTransaction || user.IsRootAdmin {
		tabs.Append(&container.TabItem{Text: "Transaction", Content: createTransactionTab(mainWindow, appState, user)})
	}

	// Revenue tab (if user has revenue permission)
	if user.CanRevenue || user.IsRootAdmin {
		tabs.Append(&container.TabItem{Text: "Revenue", Content: createRevenueTab(mainWindow, appState, user)})
	}

	// Transaction log (if user has transaction permission)
	if user.CanTransaction || user.IsRootAdmin {
		tabs.Append(&container.TabItem{Text: "Transaction Log", Content: createTransactionLogTab(mainWindow, appState, user)})
	}

	// User management (only for root admin)
	if user.IsRootAdmin {
		tabs.Append(&container.TabItem{Text: "User Management", Content: createUserManagementTab(mainWindow, appState)})
	}

	// Logout button
	logoutBtn := widget.NewButton("Logout", func() {
		appState.SetUser(nil)
		mainWindow.Close()
		ShowLoginWindow(appState)
	})

	// Reset database button (only for root admin)
	var resetBtn *widget.Button
	if user.IsRootAdmin {
		resetBtn = widget.NewButton("Reset Database", func() {
			// Show confirmation dialog
			dialog.ShowConfirm(
				"Reset Database",
				"WARNING: This will delete ALL data including users, items, and transactions. This action cannot be undone.\n\nAre you sure you want to reset the database?",
				func(confirmed bool) {
					if confirmed {
						// Show second confirmation for safety
						dialog.ShowConfirm(
							"Final Confirmation",
							"This will permanently delete ALL data. Type 'RESET' in the next dialog to confirm.",
							func(confirmed2 bool) {
								if confirmed2 {
									// Show input dialog for typing "RESET"
									confirmEntry := widget.NewEntry()
									confirmEntry.SetPlaceHolder("Type RESET to confirm")
									
									confirmContent := container.NewVBox(
										widget.NewLabel("Type 'RESET' (all caps) to confirm database reset:"),
										confirmEntry,
									)
									
									dialog.ShowCustomConfirm(
										"Type RESET to Confirm",
										"Cancel",
										"Reset",
										confirmContent,
										func(confirmed3 bool) {
											if confirmed3 && confirmEntry.Text == "RESET" {
												// Perform reset
												err := appState.GetDB().ResetDatabase()
												if err != nil {
													dialog.ShowError(fmt.Errorf("Failed to reset database: %v", err), mainWindow)
													return
												}
												
												// Show success message
												dialog.ShowInformation("Database Reset", "Database has been reset successfully. You will be logged out.", mainWindow)
												
												// Log out and return to login
												appState.SetUser(nil)
												mainWindow.Close()
												ShowLoginWindow(appState)
											} else if confirmed3 {
												dialog.ShowInformation("Invalid Confirmation", "You must type 'RESET' exactly to confirm.", mainWindow)
											}
										},
										mainWindow,
									)
								}
							},
							mainWindow,
						)
					}
				},
				mainWindow,
			)
		})
		resetBtn.Importance = widget.DangerImportance
	}

	// User icon and label
	userIcon := widget.NewIcon(theme.AccountIcon())
	userText := user.Username
	if user.IsRootAdmin {
		userText = fmt.Sprintf("%s (Admin)", user.Username)
	}
	userLabel := widget.NewLabel(userText)
	userLabel.TextStyle = fyne.TextStyle{Bold: true}

	userContainer := container.NewHBox(
		userIcon,
		userLabel,
	)

	// Build header with appropriate buttons
	var headerButtons []fyne.CanvasObject
	headerButtons = append(headerButtons, container.NewPadded(userContainer))
	headerButtons = append(headerButtons, widget.NewSeparator())
	if resetBtn != nil {
		headerButtons = append(headerButtons, resetBtn)
		headerButtons = append(headerButtons, widget.NewSeparator())
	}
	headerButtons = append(headerButtons, logoutBtn)

	header := container.NewHBox(headerButtons...)
	content := container.NewBorder(header, nil, nil, nil, tabs)

	mainWindow.SetContent(content)
	mainWindow.Show()
}

func createInventoryTab(parent fyne.Window, appState *auth.AppState, user *models.User) *container.Scroll {
	// Search bar
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search by name or code...")

	// Store items list
	var currentItems []models.Item
	var selectedID widget.ListItemID = -1

	// Item list with aligned columns using fixed-width containers
	list := widget.NewList(
		func() int {
			return len(currentItems)
		},
		func() fyne.CanvasObject {
			nameLabel := widget.NewLabel("")
			codeLabel := widget.NewLabel("")
			priceLabel := widget.NewLabel("")
			qtyLabel := widget.NewLabel("")
			// Use fixed-width containers
			return container.NewHBox(
				container.NewBorder(nil, nil, nil, nil, nameLabel),
				container.NewBorder(nil, nil, nil, nil, codeLabel),
				container.NewBorder(nil, nil, nil, nil, priceLabel),
				container.NewBorder(nil, nil, nil, nil, qtyLabel),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(currentItems) {
				item := currentItems[id]
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
				qtyLabel := box.Objects[3].(*fyne.Container).Objects[0].(*widget.Label)
				qtyLabel.SetText(fmt.Sprintf("%d", item.Quantity))
				qtyLabel.Resize(fyne.NewSize(100, qtyLabel.MinSize().Height))
			}
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		selectedID = id
	}

	// Refresh function
	refreshList := func() {
		query := strings.TrimSpace(searchEntry.Text)
		var items []models.Item
		var err error

		db := appState.GetDB().(*database.Database)
		if query == "" {
			items, err = inventory.GetAllItems(db)
		} else {
			items, err = inventory.SearchItems(db, query)
		}

		if err != nil {
			dialog.ShowError(err, parent)
			return
		}

		currentItems = items
		list.Refresh()
		selectedID = -1
	}

	searchEntry.OnChanged = func(_ string) {
		refreshList()
	}

	// Buttons
	var buttons *fyne.Container
	if user.IsRootAdmin {
		addBtn := widget.NewButton("Add Item", func() {
			showAddItemDialog(parent, appState, refreshList)
		})
		editBtn := widget.NewButton("Edit Item", func() {
			if selectedID < 0 || selectedID >= len(currentItems) {
				dialog.ShowInformation("No Selection", "Please select an item to edit", parent)
				return
			}
			showEditItemDialog(parent, appState, &currentItems[selectedID], refreshList)
		})
		restockBtn := widget.NewButton("Restock", func() {
			if selectedID < 0 || selectedID >= len(currentItems) {
				dialog.ShowInformation("No Selection", "Please select an item to restock", parent)
				return
			}
			showRestockDialog(parent, appState, &currentItems[selectedID], refreshList)
		})
		deleteBtn := widget.NewButton("Delete Item", func() {
			if selectedID < 0 || selectedID >= len(currentItems) {
				dialog.ShowInformation("No Selection", "Please select an item to delete", parent)
				return
			}
			showDeleteItemDialog(parent, appState, &currentItems[selectedID], refreshList)
		})
		buttons = container.NewHBox(addBtn, editBtn, restockBtn, deleteBtn)
	} else {
		buttons = container.NewHBox()
	}

	refreshBtn := widget.NewButton("Refresh", refreshList)
	buttons.Add(refreshBtn)

	// Column headers for inventory with fixed widths
	nameHeader := widget.NewLabel("Name")
	nameHeader.TextStyle = fyne.TextStyle{Bold: true}
	nameHeader.Resize(fyne.NewSize(200, nameHeader.MinSize().Height))
	codeHeader := widget.NewLabel("Code")
	codeHeader.TextStyle = fyne.TextStyle{Bold: true}
	codeHeader.Resize(fyne.NewSize(150, codeHeader.MinSize().Height))
	priceHeader := widget.NewLabel("Price")
	priceHeader.TextStyle = fyne.TextStyle{Bold: true}
	priceHeader.Resize(fyne.NewSize(100, priceHeader.MinSize().Height))
	qtyHeader := widget.NewLabel("Quantity")
	qtyHeader.TextStyle = fyne.TextStyle{Bold: true}
	qtyHeader.Resize(fyne.NewSize(100, qtyHeader.MinSize().Height))
	headerRow := container.NewHBox(
		container.NewBorder(nil, nil, nil, nil, nameHeader),
		container.NewBorder(nil, nil, nil, nil, codeHeader),
		container.NewBorder(nil, nil, nil, nil, priceHeader),
		container.NewBorder(nil, nil, nil, nil, qtyHeader),
	)

	content := container.NewBorder(
		container.NewVBox(
			searchEntry,
			widget.NewSeparator(),
			headerRow,
			widget.NewSeparator(),
		),
		buttons,
		nil,
		nil,
		container.NewScroll(list),
	)

	refreshList()
	return container.NewScroll(content)
}

func showAddItemDialog(parent fyne.Window, appState *auth.AppState, onSuccess func()) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Item Name")
	codeEntry := widget.NewEntry()
	codeEntry.SetPlaceHolder("Barcode/Code")
	descEntry := widget.NewMultiLineEntry()
	descEntry.SetPlaceHolder("Description")
	priceEntry := widget.NewEntry()
	priceEntry.SetPlaceHolder("Price")
	costEntry := widget.NewEntry()
	costEntry.SetPlaceHolder("Cost")
	quantityEntry := widget.NewEntry()
	quantityEntry.SetPlaceHolder("Quantity")

	formContent := container.NewVBox(
		createStyledFormField("Name", nameEntry),
		createStyledFormField("Code", codeEntry),
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
		_, err = inventory.CreateItem(db, nameEntry.Text, codeEntry.Text, descEntry.Text, price, cost, quantity)
		if err != nil {
			dialog.ShowError(err, parent)
			return
		}

		showStyledInformation(parent, "Success", "Item added successfully")
		onSuccess()
	}

	showStyledDialog(parent, "Add Item", formContent, "Add", onAction, nil)
}

func showEditItemDialog(parent fyne.Window, appState *auth.AppState, item *models.Item, onSuccess func()) {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(item.Name)
	codeEntry := widget.NewEntry()
	codeEntry.SetText(item.Code)
	descEntry := widget.NewMultiLineEntry()
	descEntry.SetText(item.Description)
	priceEntry := widget.NewEntry()
	priceEntry.SetText(fmt.Sprintf("%.2f", item.Price))
	costEntry := widget.NewEntry()
	costEntry.SetText(fmt.Sprintf("%.2f", item.Cost))
	quantityEntry := widget.NewEntry()
	quantityEntry.SetText(fmt.Sprintf("%d", item.Quantity))

	formContent := container.NewVBox(
		createStyledFormField("Name", nameEntry),
		createStyledFormField("Code", codeEntry),
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
		err = inventory.UpdateItem(db, item.ID, nameEntry.Text, codeEntry.Text, descEntry.Text, price, cost, quantity)
		if err != nil {
			dialog.ShowError(err, parent)
			return
		}

		showStyledInformation(parent, "Success", "Item updated successfully")
		onSuccess()
	}

	showStyledDialog(parent, "Edit Item", formContent, "Update", onAction, nil)
}

func showDeleteItemDialog(parent fyne.Window, appState *auth.AppState, item *models.Item, onSuccess func()) {
	dialog.ShowConfirm("Delete Item", fmt.Sprintf("Are you sure you want to delete '%s'?", item.Name), func(confirmed bool) {
		if confirmed {
			db := appState.GetDB().(*database.Database)
			err := inventory.DeleteItem(db, item.ID)
			if err != nil {
				dialog.ShowError(err, parent)
				return
			}
			showStyledInformation(parent, "Success", "Item deleted successfully")
			onSuccess()
		}
	}, parent)
}
