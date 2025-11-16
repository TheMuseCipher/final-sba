package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"ims-go/auth"
	"ims-go/database"
	"ims-go/models"
	usersPkg "ims-go/users"
)

func createUserManagementTab(parent fyne.Window, appState *auth.AppState) *container.Scroll {
	// User list
	var users []models.User
	var selectedID widget.ListItemID = -1

	list := widget.NewList(
		func() int {
			var err error
			db := appState.GetDB().(*database.Database)
			allUsers, err := usersPkg.GetAllUsers(db)
			if err != nil {
				return 0
			}
			users = allUsers
			return len(users)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(users) {
				user := users[id]
				box := obj.(*fyne.Container)
				box.Objects[0].(*widget.Label).SetText(user.Username)
				
				roleText := "User"
				if user.IsRootAdmin {
					roleText = "Root Admin"
				}
				box.Objects[1].(*widget.Label).SetText(roleText)

				perms := []string{}
				if user.CanRead {
					perms = append(perms, "Read")
				}
				if user.CanTransaction {
					perms = append(perms, "Transaction")
				}
				if user.CanRevenue {
					perms = append(perms, "Revenue")
				}
				if len(perms) == 0 {
					perms = append(perms, "None")
				}
				box.Objects[2].(*widget.Label).SetText(fmt.Sprintf("Permissions: %s", fmt.Sprint(perms)))
				box.Objects[3].(*widget.Label).SetText(fmt.Sprintf("Created: %s", user.CreatedAt.Format("2006-01-02")))
			}
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		selectedID = id
	}

	refreshList := func() {
		list.Refresh()
		selectedID = -1
	}

	// Buttons
	addBtn := widget.NewButton("Create User", func() {
		showCreateUserDialog(parent, appState, refreshList)
	})

	editBtn := widget.NewButton("Edit Permissions", func() {
		if selectedID < 0 || selectedID >= len(users) {
			dialog.ShowInformation("No Selection", "Please select a user to edit", parent)
			return
		}
		showEditUserDialog(parent, appState, &users[selectedID], refreshList)
	})

	deleteBtn := widget.NewButton("Delete User", func() {
		if selectedID < 0 || selectedID >= len(users) {
			dialog.ShowInformation("No Selection", "Please select a user to delete", parent)
			return
		}
		showDeleteUserDialog(parent, appState, &users[selectedID], refreshList)
	})

	refreshBtn := widget.NewButton("Refresh", refreshList)

	buttons := container.NewHBox(addBtn, editBtn, deleteBtn, refreshBtn)

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("User Management"),
			widget.NewSeparator(),
		),
		buttons,
		nil,
		nil,
		list,
	)

	refreshList()
	return container.NewScroll(content)
}

func showCreateUserDialog(parent fyne.Window, appState *auth.AppState, onSuccess func()) {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")
	canReadCheck := widget.NewCheck("Can Read", nil)
	canTransactionCheck := widget.NewCheck("Can Transaction", nil)
	canRevenueCheck := widget.NewCheck("Can Revenue", nil)

	formContent := container.NewVBox(
		createStyledFormField("Username", usernameEntry),
		createStyledFormField("Password", passwordEntry),
		createStyledFormField("Permissions", container.NewVBox(canReadCheck, canTransactionCheck, canRevenueCheck)),
	)

	onAction := func() {
		db := appState.GetDB().(*database.Database)
		_, err := usersPkg.CreateUser(db,
			usernameEntry.Text,
			passwordEntry.Text,
			canReadCheck.Checked,
			canTransactionCheck.Checked,
			canRevenueCheck.Checked,
		)
		if err != nil {
			dialog.ShowError(err, parent)
			return
		}

		showStyledInformation(parent, "Success", "User created successfully")
		onSuccess()
	}

	showStyledDialog(parent, "Create User", formContent, "Create", onAction, nil)
}

func showEditUserDialog(parent fyne.Window, appState *auth.AppState, user *models.User, onSuccess func()) {
	if user.IsRootAdmin {
		dialog.ShowInformation("Cannot Edit", "Root admin permissions cannot be modified", parent)
		return
	}

	canReadCheck := widget.NewCheck("Can Read", nil)
	canReadCheck.SetChecked(user.CanRead)
	canTransactionCheck := widget.NewCheck("Can Transaction", nil)
	canTransactionCheck.SetChecked(user.CanTransaction)
	canRevenueCheck := widget.NewCheck("Can Revenue", nil)
	canRevenueCheck.SetChecked(user.CanRevenue)

	formContent := container.NewVBox(
		createStyledFormField("Username", widget.NewLabel(user.Username)),
		createStyledFormField("Permissions", container.NewVBox(canReadCheck, canTransactionCheck, canRevenueCheck)),
	)

	onAction := func() {
		db := appState.GetDB().(*database.Database)
		err := usersPkg.UpdateUserPermissions(db,
			user.ID,
			canReadCheck.Checked,
			canTransactionCheck.Checked,
			canRevenueCheck.Checked,
		)
		if err != nil {
			dialog.ShowError(err, parent)
			return
		}

		showStyledInformation(parent, "Success", "User permissions updated successfully")
		onSuccess()
	}

	showStyledDialog(parent, "Edit User Permissions", formContent, "Update", onAction, nil)
}

func showDeleteUserDialog(parent fyne.Window, appState *auth.AppState, user *models.User, onSuccess func()) {
	if user.IsRootAdmin {
		dialog.ShowInformation("Cannot Delete", "Root admin cannot be deleted", parent)
		return
	}

	dialog.ShowConfirm("Delete User", fmt.Sprintf("Are you sure you want to delete user '%s'?", user.Username), func(confirmed bool) {
		if confirmed {
			db := appState.GetDB().(*database.Database)
			err := usersPkg.DeleteUser(db, user.ID)
			if err != nil {
				dialog.ShowError(err, parent)
				return
			}
			showStyledInformation(parent, "Success", "User deleted successfully")
			onSuccess()
		}
	}, parent)
}

