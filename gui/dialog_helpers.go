package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// showStyledDialog creates a custom dialog window with rounded corners and custom button labels
func showStyledDialog(parent fyne.Window, title string, content fyne.CanvasObject, actionLabel string, onAction func(), onDone func()) {
	// Get the app from parent window
	app := fyne.CurrentApp()
	dialogWindow := app.NewWindow(title)

	// Get screen size for 30% width (minimum 600px to ensure fields are visible)
	screenSize := parent.Canvas().Size()
	dialogWidth := float32(screenSize.Width) * 0.3
	if dialogWidth < 600 {
		dialogWidth = 600
	}

	dialogWindow.Resize(fyne.NewSize(dialogWidth, 600))
	dialogWindow.CenterOnScreen()
	dialogWindow.SetFixedSize(false)

	// Create buttons
	actionBtn := widget.NewButton(actionLabel, func() {
		if onAction != nil {
			onAction()
		}
		dialogWindow.Close()
	})

	doneBtn := widget.NewButton("Done", func() {
		if onDone != nil {
			onDone()
		}
		dialogWindow.Close()
	})

	buttons := container.NewHBox(
		actionBtn,
		widget.NewSeparator(),
		doneBtn,
	)

	// Create styled form with minimal padding
	styledContent := styleFormContent(content)

	// Centered title
	titleLabel := widget.NewLabel(title)
	titleLabel.Alignment = fyne.TextAlignCenter
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Ensure form content expands to fill available width
	formScroll := container.NewScroll(styledContent)

	mainContent := container.NewBorder(
		container.NewPadded(titleLabel),
		container.NewVBox(widget.NewSeparator(), buttons),
		nil,
		nil,
		formScroll, // Scroll for long forms
	)

	// Minimal padding around the entire content
	finalContent := container.NewPadded(mainContent)

	dialogWindow.SetContent(finalContent)
	dialogWindow.Show()
}

// styleFormContent wraps form fields with minimal padding
func styleFormContent(content fyne.CanvasObject) fyne.CanvasObject {
	// Minimal padding - content is already a VBox from the caller
	return container.NewPadded(content)
}

// createStyledFormField creates a form field with label on left, field on right (horizontal layout)
// All fields will be vertically aligned and expand to fill available width
func createStyledFormField(labelText string, fieldWidget fyne.CanvasObject) fyne.CanvasObject {
	label := widget.NewLabel(labelText)
	label.TextStyle = fyne.TextStyle{Bold: true}

	// Set fixed width for label to ensure alignment
	labelContainer := container.NewBorder(nil, nil, nil, nil, label)
	labelContainer.Resize(fyne.NewSize(120, 0)) // Fixed width for label column

	// Create a card-like container for the field (gray background box)
	fieldCard := widget.NewCard("", "", fieldWidget)

	// Use Border layout: label on left, field card in center (expands to fill remaining space)
	// This ensures the field expands properly and is fully visible
	return container.NewBorder(
		nil, nil,
		labelContainer, // Left: fixed-width label
		nil,
		fieldCard, // Center: field card that expands
	)
}

// showStyledInformation creates a styled information dialog with rounded corners
func showStyledInformation(parent fyne.Window, title, message string) {
	infoWindow := fyne.CurrentApp().NewWindow(title)
	infoWindow.Resize(fyne.NewSize(400, 200))
	infoWindow.CenterOnScreen()
	infoWindow.SetFixedSize(false)

	label := widget.NewLabel(message)
	label.Wrapping = fyne.TextWrapWord

	okBtn := widget.NewButton("OK", func() {
		infoWindow.Close()
	})

	content := container.NewBorder(
		nil,
		container.NewPadded(okBtn),
		nil,
		nil,
		container.NewPadded(label),
	)

	finalContent := container.NewPadded(content)
	infoWindow.SetContent(finalContent)
	infoWindow.Show()
}

// Store last file picker location
var lastFilePickerURI fyne.URI

// Opens a file picker in the last location that was used
func showFilePickerWithMemory(parent fyne.Window, callback func(fyne.URIReadCloser, error)) {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if reader != nil {
			lastFilePickerURI = reader.URI()
		}
		callback(reader, err)
	}, parent)
}
