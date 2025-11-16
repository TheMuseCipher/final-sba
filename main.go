package main

import (
	"ims-go/auth"
	"ims-go/database"
	"ims-go/gui"
)

func main() {
	// Initialize database
	db, err := database.NewDatabase()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Initialize app state
	appState := auth.NewAppState(db)

	// Show login window
	gui.ShowLoginWindow(appState)
}
