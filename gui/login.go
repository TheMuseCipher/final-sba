package gui

import (

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"ims-go/auth"
)

func ShowLoginWindow(appState *auth.AppState) {
	myApp := app.New()
	loginWindow := myApp.NewWindow("Login - Inventory Management System")
	loginWindow.Resize(fyne.NewSize(400, 300))
	loginWindow.CenterOnScreen()

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord

	loginBtn := widget.NewButton("Login", func() {
		username := usernameEntry.Text
		password := passwordEntry.Text

		if username == "" || password == "" {
			statusLabel.SetText("Please enter both username and password")
			return
		}

		user, err := appState.Authenticate(username, password)
		if err != nil {
			statusLabel.SetText("Invalid credentials")
			return
		}

		// Login successful, show main window
		loginWindow.Hide()
		ShowMainWindow(myApp, appState, user)
	})

	content := container.NewVBox(
		widget.NewLabel("Inventory Management System"),
		widget.NewSeparator(),
		usernameEntry,
		passwordEntry,
		loginBtn,
		statusLabel,
	)

	loginWindow.SetContent(container.NewCenter(content))
	loginWindow.ShowAndRun()
}

